test:
	go test -v -cover -race ./...

start:
	go run main.go start --debug

lint:
	golangci-lint run

.PHONY: test start lint