.PHONY: run build up down logs restart docker-up docker-down docker-build docker-logs migrate migrate-up migrate-down seed test lint clean

# Run with hot reload (requires air: go install github.com/air-verse/air@latest)
dev:
	@echo "Starting server with hot reload..."
	@echo "Install air if not available: go install github.com/air-verse/air@latest"
	air

# Alias for dev
run: dev

# Run without hot reload
run-simple:
	go run ./cmd/server

# Build binary
build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server

# Docker shortcuts
up:
	docker compose up -d

# Start only database (for local development)
db:
	@echo "Starting database only..."
	docker compose up -d db

down:
	docker compose down

logs:
	docker compose logs -f

restart:
	docker compose restart

# Docker Compose
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-build:
	docker compose up -d --build

docker-logs:
	docker compose logs -f

# Database migrations
migrate-up:
	@echo "Running database migrations..."
	docker compose exec app go run ./cmd/app -migrate

migrate-down:
	@echo "Rolling back last migration..."
	docker compose exec app go run ./cmd/app -migrate-down

migrate:
	@echo "Running migrations locally..."
	go run ./cmd/app -migrate

seed:
	@echo "Seeding database..."
	docker compose exec app go run ./cmd/app -seed

# Testing
test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Linting (requires golangci-lint)
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

# Tidy dependencies
tidy:
	go mod tidy

# Format code
fmt:
	gofmt -s -w .
