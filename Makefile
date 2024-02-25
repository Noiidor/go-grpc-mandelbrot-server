PROJECT_NAME = go-grpc-mandelbrot
PROJECT_PATH = cmd/$(PROJECT_NAME).go
ENV_MODE ?= dev

.PHONY:run
run:
	go run $(PROJECT_PATH)

.PHONY:build
build:
	go build -o $(PROJECT_NAME) $(PROJECT_PATH)

.PHONY:test
test:
	go test ./...

.PHONY:lint
lint:
	golangci-lint run