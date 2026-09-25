.PHONY: test run build docker

test:
	go test ./...

run:
	go run ./cmd/api

build:
	go build ./cmd/api

docker:
	docker build -t jev-banking-routing:local .
