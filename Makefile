test:
	go test -race ./...

test_watch:
	watch -n 5 make test

ci:
	go get -t -d -v ./... && make test

lint:
	go get -u golang.org/x/lint/golint && golint ./...
