.PHONY: help lint lint-fix build run test clean install-tools

help:
	@echo "Available commands:"
	@echo "  make lint          - Run golangci-lint"
	@echo "  make lint-fix      - Run golangci-lint with auto-fix"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the application"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make install-tools - Install development tools"

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

build:
	go build -o bin/bot ./cmd/bot

run:
	go run ./cmd/bot

test:
	go test -v -race -coverprofile=coverage.out ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

install-tools:
	@echo "Installing golangci-lint..."
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
	@echo "Tools installed successfully"

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

migrate-up:
	cd migrations && goose postgres "$(DATABASE_URL)" up

migrate-down:
	cd migrations && goose postgres "$(DATABASE_URL)" down

migrate-status:
	cd migrations && goose postgres "$(DATABASE_URL)" status
