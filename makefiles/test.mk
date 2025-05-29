.PHONY: test coverage lint

test: ## Run unit tests
	go test ./...

coverage: ## Generate coverage report
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out

lint: ## Run linter
	golangci-lint run
