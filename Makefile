.PHONY: build run clean test docker-build docker-run docker-stop

# Build the application
build:
	go build -o bin/telegram-bot main.go

# Run the application
run:
	go run main.go

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test ./...

# Install dependencies
deps:
	go mod tidy
	go mod download

# Build Docker image
docker-build:
	docker build -t telegram-bot .

# Run with Docker Compose
docker-run:
	docker-compose up -d

# Stop Docker containers
docker-stop:
	docker-compose down

# View logs
docker-logs:
	docker-compose logs -f bot

# Database setup (create database)
db-create:
	mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS telegram_bot CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Development setup
dev-setup: deps db-create
	@echo "Development environment setup complete!"
	@echo "Don't forget to update config.yaml with your settings"

# Production deployment
deploy: docker-build docker-run
	@echo "Application deployed successfully!"
