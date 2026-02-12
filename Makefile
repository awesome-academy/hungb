.PHONY: run build docker-up docker-down migrate seed test lint clean

# Run with hot reload (requires air: go install github.com/air-verse/air@latest)
run:
	@echo "Starting server with hot reload..."
	air

# Run without hot reload
run-simple:
	go run ./cmd/server

# Build binary
build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server

# Docker Compose
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-build:
	docker compose up -d --build

docker-logs:
	docker compose logs -f

# Database
migrate:
	go run ./cmd/server -migrate

seed:
	go run ./cmd/server -seed

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
