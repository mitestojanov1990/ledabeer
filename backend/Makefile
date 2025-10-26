.PHONY: test build lint cover clean

# Test targets
test:
	go test -v ./...

test-race:
	go test -race -v ./...

test-coverage:
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Build targets
build:
	go build -o bin/node ./cmd/node

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/node-linux ./cmd/node

# Linting
lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

# Coverage
cover: test-coverage
	@echo "Coverage report generated: coverage.html"

# Clean
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Development
dev:
	go run ./cmd/node

# Docker
docker-build:
	docker build -t ledabeer-backend .

docker-test:
	docker-compose -f ../docker/docker-compose.yml up --abort-on-container-exit

# All
all: lint test build
