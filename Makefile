.PHONY: build run test clean docker-build docker-run swagger-gen swagger-serve

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
	rm -rf docs/docs.go docs/swagger.json docs/swagger.yaml

# Build Docker image
docker-build:
	docker build -t s3like:latest .

# Run with Docker Compose
docker-run:
	docker-compose up --build

# Stop Docker Compose
docker-stop:
	docker-compose down

# Generate Swagger documentation
swagger-gen:
	@echo "🔄 Generating Swagger documentation..."
	@if ! command -v swag &> /dev/null; then \
		echo "📦 Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@swag init -g cmd/server/main.go -o docs/
	@echo "✅ Swagger documentation generated!"

# Serve Swagger documentation (requires swagger-ui)
swagger-serve:
	@echo "🌐 Serving Swagger UI..."
	@echo "📖 Open http://localhost:8080/swagger/index.html after starting the server"

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

# Full development setup
dev-setup: deps swagger-gen
	@echo "🚀 Development environment ready!"
	@echo "📝 Edit .env file with your configuration"
	@echo "🏃 Run 'make run' to start the server"
	@echo "📖 Swagger UI will be available at http://localhost:8080/swagger/index.html"

# Build and run with swagger
dev: swagger-gen run
