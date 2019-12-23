'use strict';
const dateFormat = require('dateformat');
const rp = require('request-promise');
var AWS = require("aws-sdk");

function validateInputBody(inputBody) {
    if (!inputBody.username) {
        return ({status: false, message: 'Please enter username for F5 Cloud Services Account. Eg, "username": "foo"'});
    } else if (!inputBody.password) {
        return ({
            status: false,
            message: 'Please enter password for F5 Cloud Services Account. Eg, "password": "fooPassword"'
        });
    } else if (!inputBody.accountid) {
        return ({
            status: false,
            message: 'Please enter Account-ID for Preferred F5 Cloud Services Account. Eg, "accountid": "a-fooAccountID"'
        });
    }
    return ({status: true, message: "All input parameters received."})
}

module.exports.costInsight = async event => {
    const self = this;

    function getAccessKey(Username, Password) {
        return new Promise((resolve) => {
            resolve(
                rp.post({
                    url: 'https://api.dev.f5aas.com/v1/svc-auth/login',
                    body: {username: Username, password: Password},
                    json: true
                }));
        });
    }

    function loginF5Portal(AccessToken) {
        return new Promise((resolve) => {
            resolve(
                rp.defaults({
                    auth: {bearer: AccessToken},
                    baseUrl: "https://api.dev.f5aas.com/beacon/v1/",
                    json: true
                })
            );
        });
    }

    function getCost(StartDate, EndDate) {
        var costexplorer = new AWS.CostExplorer();

        // Documentation: https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/ce-exploring-data.html
        var costParams = {
            TimePeriod: {
                End: EndDate,
                Start: StartDate
            },
            Granularity: 'MONTHLY',
            Metrics: ['AmortizedCost']
        };

        return new Promise((resolve, reject) => {
            costexplorer.getCostAndUsage(costParams, function (err, data) {
                if (err) {
                    reject(err);
                } else {
                    resolve(data);
                }
            })
        });
    }

    function createInsight(CostArray) {
        var MarkdownContent = "| Date | Amount | Unit |\n| :--- | :--- | :---- |\n";
        var totalCost = 0;
        const unit = CostArray.length > 0 ? CostArray[0].Total.AmortizedCost.Unit : 'USD';
        CostArray.ResultsByTime.forEach(result => {
            totalCost += Number(result.Total.AmortizedCost.Amount);
            MarkdownContent = MarkdownContent + `| ${result.TimePeriod.Start} | ${Number(result.Total.AmortizedCost.Amount).toFixed(2).toString()} | ${result.Total.AmortizedCost.Unit} |\n`;

        });
        totalCost = totalCost.toFixed(2);
        MarkdownContent = `\n ##### Total AWS Cost: ${totalCost} ${unit} \n` + MarkdownContent;
        const created_insight = {
            title: "AWS Cost Insight",
            description: `AWS cost incurred over the last 12 months along with the monthly distribution of the cost can be seen below.`,
            markdownContent: MarkdownContent,
            category: "INS_CAT_COST",
            severity: "INS_SEV_INFORMATIONAL"
        };
        return created_insight;
    }

    function publishCostInsight(createdInsight, accountid) {
        let options = {
            body: createdInsight,
        };

        options.uri = '/insights';
        options.method = "POST";
        if (accountid) {
            options.headers = {
                'X-F5aas-Preferred-Account-Id': accountid
            };
        }

        return new Promise((resolve) => {
            resolve(
                self.rpWithAuth.post(options)
            );
        });

    }

    const date = new Date();
    const StartDate = dateFormat(new Date(date.getFullYear(), date.getMonth() - 11, 1), "yyyy-mm-dd");
    const EndDate = dateFormat(new Date(date.getFullYear(), date.getMonth() + 1, 0), "yyyy-mm-dd");

    const jsonData = event;
    const validateMessageResult = validateInputBody(jsonData);
    // console.log(validateMessageResult.status, validateMessageResult.message);

    if (validateMessageResult.status) {
        try {
            const tokenRes = await getAccessKey(jsonData.username, jsonData.password);
            self.rpWithAuth = await loginF5Portal(tokenRes.access_token);
            console.log('Login Successful.');
            const costRes = await getCost(StartDate, EndDate);
            const createdInsight = createInsight(costRes);
            await publishCostInsight(createdInsight, jsonData.accountid);
            console.log('Cost Insight created and published.');
        } catch (e) {
            console.log(e);
        }
    }
    return 'Insight create/publish unsuccessful.';

};
