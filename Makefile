test:
	go test -cover -race ./...

testwatch:
	watch -n 5 make test

ci:
	go test -race -coverprofile=coverage ./... && go tool cover -func=coverage

lint:
	golangci-lint run ./...

gettools:
	go install golang.org/x/lint/golint@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

testcovhtml:
	go test -coverprofile=coverage ./... && go tool covdata textfmt -i=coverage -o=coverage.txt && go tool cover -html=coverage.txt
