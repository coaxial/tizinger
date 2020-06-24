test:
	go test ./...

test_watch:
	watch -n 5 make test

ci:
	go get -t -d -v ./... && make test
