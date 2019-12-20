.PHONY: go
go: ## go checks module packages
	$(ECHO) go list -u -m all
	$(ECHO) go mod verify
	$(ECHO) go mod tidy


.PHONY: build
build:
	$(ECHO) GOOS=linux go build main.go

.PHONY: release
release: build
	$(ECHO) zip function.zip main
	$(ECHO) aws lambda update-function-code --function-name beacon-lambda-insight-example --zip-file fileb://function.zip
