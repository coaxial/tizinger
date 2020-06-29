test:
	go test -cover -race ./...

testwatch:
	watch -n 5 make test

ci:
	go get -t -d -v ./... && go test -race -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

lint:
	golint ./...

gettools:
	go get -u golang.org/x/lint/golint

testcovhtml:
	go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
