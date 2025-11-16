.PHONY: build run test lint clean docker-up docker-down migrate-up migrate-down load-test e2e-test

build:
	go build -o bin/app cmd/app/main.go

run:
	go run cmd/app/main.go

test:
	go test -v -cover ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down -v

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

e2e-test:
	go test -v ./tests/e2e/...

load-test:
	go run tests/loadtest/main.go

help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run unit tests"
	@echo "  test-coverage - Generate test coverage report"
	@echo "  lint         - Run linter"
	@echo "  lint-fix     - Run linter with auto-fix"
	@echo "  docker-up    - Start services with docker-compose"
	@echo "  docker-down  - Stop and remove docker containers"
	@echo "  clean        - Clean build artifacts"
	@echo "  e2e-test     - Run E2E tests"
	@echo "  load-test    - Run load tests"
