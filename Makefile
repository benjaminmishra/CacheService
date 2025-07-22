BINARY_NAME=cache-service

.PHONY: all help build run test race up down logs

.DEFAULT_GOAL := help

help:
	@echo "Available commands:"
	@echo "  make build      Build the application binary"
	@echo "  make run        Run the application locally"
	@echo "  make test       Run tests"
	@echo "  make test-race  Run tests with race detector needs CGO_ENABLED=1 env variable set"
	@echo "  make coverage   Generate coverage report"
	@echo "  make up         Start the service with Docker"
	@echo "  make down       Stop the Docker service"

build:
	@echo "Building binary..."
	go build -o bin/$(BINARY_NAME) ./cmd/cache-service

run:
	go run ./cmd/cache-service

test:
	@echo "Running tests..."
	go test -v ./...

test-race:
	@echo "Running tests with race detector..."
	go test -race -v -timeout 2m ./...

benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

coverage:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

up:
	@echo "Starting Docker service..."
	docker compose up --build

down:
	@echo "Stopping Docker service..."
	docker compose down