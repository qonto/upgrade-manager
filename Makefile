test:
	go test -v -cover -race ./...

dev:
	go run main.go start -v debug  --config-file config/config.yml

start:
	go run main.go start -v debug

lint:
	golangci-lint run

.PHONY: test start lint