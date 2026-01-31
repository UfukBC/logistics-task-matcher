.PHONY: help build run test test-coverage test-bench clean docker-build docker-run docker-stop install lint

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install dependencies
	go mod download
	go mod tidy

build: ## Build the application
	go build -o task-assignment-engine .

run: ## Run the application
	go run .

test: ## Run tests
	go test -v

test-coverage: ## Run tests with coverage
	go test -v -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-bench: ## Run benchmarks
	go test -bench=. -benchmem

lint: ## Run linter (requires golangci-lint)
	golangci-lint run

clean: ## Clean build artifacts
	rm -f task-assignment-engine coverage.out coverage.html

docker-build: ## Build Docker image
	docker-compose build

docker-run: ## Run with Docker Compose
	docker-compose up

docker-run-detached: ## Run with Docker Compose in detached mode
	docker-compose up -d

docker-stop: ## Stop Docker Compose
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

docker-clean: ## Clean Docker resources
	docker-compose down -v
	docker system prune -f
