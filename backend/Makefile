.PHONY: help build run test clean migrate docker-up docker-down deps lint fmt

# Variables
APP_NAME=cbalite-backend
MAIN_PATH=cmd/api
BINARY_NAME=main
DOCKER_COMPOSE=docker-compose

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m

help: ## Show this help message
	@echo "${GREEN}Available commands:${NC}"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  ${YELLOW}%-15s${NC} %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download Go dependencies
	@echo "${GREEN}Downloading dependencies...${NC}"
	go mod download
	go mod tidy

build: ## Build the application
	@echo "${GREEN}Building application...${NC}"
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)/*.go

run: ## Run the application locally
	@echo "${GREEN}Running application...${NC}"
	go run $(MAIN_PATH)/*.go

test: ## Run tests
	@echo "${GREEN}Running tests...${NC}"
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	@echo "${GREEN}Generating coverage report...${NC}"
	go tool cover -html=coverage.out -o coverage.html
	@echo "${GREEN}Coverage report generated: coverage.html${NC}"

lint: ## Run linter
	@echo "${GREEN}Running linter...${NC}"
	@if ! which golangci-lint > /dev/null; then \
		echo "${YELLOW}Installing golangci-lint...${NC}"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run

fmt: ## Format code
	@echo "${GREEN}Formatting code...${NC}"
	go fmt ./...
	gofmt -s -w .

clean: ## Clean build artifacts
	@echo "${GREEN}Cleaning build artifacts...${NC}"
	rm -rf bin/
	rm -f coverage.out coverage.html

migrate-up: ## Run database migrations up
	@echo "${GREEN}Running migrations up...${NC}"
	@if ! which migrate > /dev/null; then \
		echo "${YELLOW}Please install golang-migrate first${NC}"; \
		echo "Visit: https://github.com/golang-migrate/migrate"; \
		exit 1; \
	fi
	migrate -path migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up

migrate-down: ## Run database migrations down
	@echo "${GREEN}Running migrations down...${NC}"
	@if ! which migrate > /dev/null; then \
		echo "${YELLOW}Please install golang-migrate first${NC}"; \
		echo "Visit: https://github.com/golang-migrate/migrate"; \
		exit 1; \
	fi
	migrate -path migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down

migrate-create: ## Create a new migration (usage: make migrate-create name=migration_name)
	@echo "${GREEN}Creating migration: $(name)${NC}"
	@if ! which migrate > /dev/null; then \
		echo "${YELLOW}Please install golang-migrate first${NC}"; \
		echo "Visit: https://github.com/golang-migrate/migrate"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations -seq $(name)

docker-build: ## Build Docker image
	@echo "${GREEN}Building Docker image...${NC}"
	docker build -t $(APP_NAME):latest .

docker-up: ## Start services with docker-compose
	@echo "${GREEN}Starting services...${NC}"
	$(DOCKER_COMPOSE) up -d

docker-down: ## Stop services with docker-compose
	@echo "${GREEN}Stopping services...${NC}"
	$(DOCKER_COMPOSE) down

docker-logs: ## Show docker-compose logs
	$(DOCKER_COMPOSE) logs -f

docker-clean: ## Clean docker resources
	@echo "${GREEN}Cleaning Docker resources...${NC}"
	$(DOCKER_COMPOSE) down -v
	docker system prune -f

dev: ## Start development environment
	@echo "${GREEN}Starting development environment...${NC}"
	@if [ ! -f .env ]; then \
		echo "${YELLOW}Creating .env file from .env.example...${NC}"; \
		cp .env.example .env; \
	fi
	@echo "${GREEN}Starting database and cache services...${NC}"
	$(DOCKER_COMPOSE) up -d postgres-local redis-local
	@echo "${GREEN}Waiting for services to be ready...${NC}"
	@sleep 5
	@echo "${GREEN}Running application...${NC}"
	go run $(MAIN_PATH)/*.go

stop: ## Stop all services
	@echo "${GREEN}Stopping all services...${NC}"
	$(DOCKER_COMPOSE) down
	@pkill -f "go run" || true

install-tools: ## Install development tools
	@echo "${GREEN}Installing development tools...${NC}"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/swaggo/swag/cmd/swag@latest

check: lint test ## Run all checks (lint and test)
	@echo "${GREEN}All checks passed!${NC}"

.DEFAULT_GOAL := help