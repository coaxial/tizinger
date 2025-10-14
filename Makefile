test:
	go test -cover -race ./...

testwatch:
	watch -n 5 make test

ci:
	go test -race -coverprofile=coverage.out ./... && go tool covdata func -i=coverage.out

lint:
	golangci-lint run ./...

gettools:
	go install golang.org/x/lint/golint@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

testcovhtml:
	go test -coverprofile=coverage.out ./... && go tool covdata textfmt -i=coverage.out -o coverage.txt && go tool cover -html=coverage.txt
