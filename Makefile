.PHONY: build run test clean docker-build docker-run

# Build the application
build:
	go build -o bin/s3like cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build Docker image
docker-build:
	docker build -t s3like:latest .

# Run with Docker Compose
docker-run:
	docker-compose up --build

# Stop Docker Compose
docker-stop:
	docker-compose down

# Run database migrations
migrate:
	go run cmd/migrate/main.go

# Generate mocks (if using mockery)
mocks:
	mockery --all --output=mocks

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod download
	go mod tidy

# Create .env file from example
env:
	cp .env.example .env
