.PHONY: build install clean test release help

BINARY_NAME=check-projects
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building ${BINARY_NAME}..."
	go build ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/check-projects

install: build ## Build and install to /usr/local/bin
	@echo "Installing ${BINARY_NAME} to /usr/local/bin..."
	cp bin/${BINARY_NAME} /usr/local/bin/
	@echo "✔ ${BINARY_NAME} installed successfully!"

clean: ## Remove build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf dist/

test: ## Run tests
	go test -v ./...

deps: ## Download dependencies
	go mod download
	go mod tidy

release: clean ## Build binaries for all platforms
	@echo "Building release binaries..."
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-amd64 ./cmd/check-projects
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-arm64 ./cmd/check-projects
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-amd64 ./cmd/check-projects
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-arm64 ./cmd/check-projects
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-windows-amd64.exe ./cmd/check-projects
	@echo "✔ Release binaries built in dist/"

run: build ## Build and run
	./bin/${BINARY_NAME}

dev: ## Run without building (using go run)
	go run ./cmd/check-projects

lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: brew install golangci-lint"; exit 1)
	golangci-lint run
