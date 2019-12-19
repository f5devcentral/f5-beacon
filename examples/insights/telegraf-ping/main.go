package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"

	"github.com/aws/aws-lambda-go/lambda"
)

type FunctionInput struct {
	// F5CS account user name
	Username string `json:"username"`
	// F5CS account password
	Password string `json:"password"`
	// F5CS account prefered account
	// See for details:https://clouddocs.f5.com/cloud-services/latest/f5-cloud-services-Beacon-WorkWith.html#specify-preferred-account-header-in-a-multiple-accounts-divisions-scenario
	Account string `json:"account"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

type InsightRequest struct {
	Title           string `json:"title,omitempty"`
	Description     string `json:"description,omitempty"`
	MarkdownContent string `json:"markdownContent,omitempty"`
	Category        string `json:"category,omitempty"`
	Severity        string `json:"severity,omitempty"`
}

type Insight struct {
	Id string `json:"id"`
}

type InsightResults struct {
	Insights []Insight `json:"insights"`
}

type QueryBody struct {
	Query string `json:"query"`
}

// var basePath = "https://api.cloudservices.f5.com"
var basePath = "https://api.dev.f5aas.com"

func handler(functionInput FunctionInput) error {

	fmt.Printf("Starting processing Ping Insight\n")

	loginResp, err := login(LoginRequest{
		Username: functionInput.Username,
		Password: functionInput.Password,
	})

	if err != nil {
		return err
	}

	insight, err := buildPingInsight(loginResp.AccessToken, functionInput.Account)
	if err != nil {
		return err
	}

	publishInsight(loginResp.AccessToken, functionInput.Account, insight)
	return nil
}

// Login to F5CS and get access token
func login(login LoginRequest) (*LoginResponse, error) {

	urlPath := basePath + "/v1/svc-auth/login"
	loginBytes, err := json.Marshal(login)
	if err != nil {
		fmt.Printf("Failed to create json from body new request %s\n", err.Error())
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, urlPath, bytes.NewBuffer(loginBytes))
	if err != nil {
		fmt.Printf("Failed to create new request %s\n", err.Error())
		return nil, err
	}

	responseBody, err := executeRequest(req)
	if err != nil {
		return nil, err
	}
	loginResp := LoginResponse{}
	err = json.Unmarshal(responseBody, &loginResp)
	if err != nil {
		fmt.Printf("Failed to unmarshal login reply %s\n", err.Error())
		return nil, err
	}

	return &loginResp, err
}

// Build a Beacon insight base on ping metric results
func buildPingInsight(token string, account string) (*InsightRequest, error) {

	// Query the ping metric
	pingPacketLossPercentage, err := queryPingMetric(token, account)
	if err != nil {
		fmt.Printf("Unable to query for Ping metrics %s\n", err.Error())
		return nil, err
	}

	severity := "INS_SEV_MODERATE"
	if pingPacketLossPercentage == 0 {
		severity = "INS_SEV_INFORMATIONAL"
	}

	// Build insight request body
	var insightRequest = InsightRequest{
		Title:           "Ping Insight",
		Description:     "Periodic Ping Test Insight",
		MarkdownContent: fmt.Sprintf("### Ping result\n Ping average lost packet %.2f", pingPacketLossPercentage),
		Category:        "INS_CAT_OPERATIONS",
		Severity:        severity,
	}
	return &insightRequest, nil
}

// Query for ping packet loss percentage
func queryPingMetric(token string, account string) (float64, error) {

	// Query data for the last 2 hours
	timeSince := fmt.Sprint(time.Now().Add(-1 * time.Hour).Format("2006-01-02T15:04:05Z"))
	urlPath := basePath + "/beacon/v1/metrics"

	var queryBody = QueryBody{
		Query: fmt.Sprintf("SELECT mean(\"percent_packet_loss\") AS \"mean_percent_packet_loss\" FROM \"ping\" WHERE time > '%s' GROUP BY time(15m) FILL(none)", timeSince),
	}

	queryBytes, err := json.Marshal(queryBody)

	if err != nil {
		fmt.Printf("Failed to create json from body new request %s\n", err.Error())
		return -1, err
	}

	req, err := createHttpRequest(http.MethodPost, urlPath, queryBytes, token, account)
	if err != nil {
		return -1, err
	}

	responseBody, err := executeRequest(req)
	if err != nil {
		return -1, err
	}
	fmt.Printf("Result %s\n", string(responseBody))

	var response client.Response
	err = json.Unmarshal(responseBody, &response)

	if len(response.Results) == 0 {
		err = errors.New("missing result from metric query")
		fmt.Println(err.Error())
		return -1, err
	}

	if len(response.Results[0].Series) == 0 {
		err = errors.New("missing result series from metric query")
		fmt.Println(err.Error())
		return -1, err
	}

	if len(response.Results[0].Series[0].Values) == 0 {
		err = errors.New("missing result series values from metric query")
		fmt.Println(err.Error())
		return -1, err
	}
	valLen := len(response.Results[0].Series[0].Values)

	fmt.Printf("Values index %d\n", valLen)

	// Get last hour ping average failure
	return response.Results[0].Series[0].Values[valLen-1][1].(float64), err
}

// Publish the insight to Beacon
func publishInsight(token string, account string, insight *InsightRequest) {

	urlPath := basePath + "/beacon/v1/insights"
	verb := http.MethodPost

	insightBytes, err := json.Marshal(insight)
	if err != nil {
		fmt.Printf("Failed to create json from body new request %s\n", err.Error())
		return
	}

	req, err := createHttpRequest(verb, urlPath, insightBytes, token, account)
	if err != nil {
		return
	}

	_, err = executeRequest(req)
	if err != nil {
		return
	}
}

func createHttpRequest(verb string, urlPath string, inputByte []byte, token string, account string) (*http.Request, error) {
	req, err := http.NewRequest(verb, urlPath, bytes.NewBuffer(inputByte))
	if err != nil {
		fmt.Printf("Failed to create request %s\n", err.Error())
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-F5aas-Preferred-Account-Id", account)
	return req, nil
}

func executeRequest(req *http.Request) ([]byte, error) {

	httpClient := http.Client{Timeout: time.Second * 10}
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("%s %v failed, error: %s\n", req.Method, req.URL, err.Error())
		return []byte{}, err
	}
	fmt.Printf("%s %v returned %v\n", req.Method, req.URL, resp.Status)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return []byte{}, errors.New("unexpected request status code")
	}
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, errors.New(fmt.Sprintf("unable to read response body %s", err.Error()))
	}
	return responseBody, nil
}

func main() {
	lambda.Start(handler)
}
