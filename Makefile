test:
	go test -race ./...

test_watch:
	watch -n 5 make test

ci:
	go get -t -d -v ./... && make test

lint:
	golint ./...

get_tools:
	go get -u golang.org/x/lint/golint
