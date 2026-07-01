.PHONY: build build-linux build-all test clean fmt vet deps help

OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)

ifeq ($(ARCH),x86_64)
  ARCH := amd64
endif
ifeq ($(ARCH),aarch64)
  ARCH := arm64
endif

VERSION ?= $(shell \
  git rev-parse --short HEAD 2>/dev/null || echo dev)
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build binary for the current platform
	go build $(LDFLAGS) -o ./build/autodiscover-$(OS)-$(ARCH) ./cmd/autodiscover

build-linux: ## Build binaries for linux/amd64 and linux/arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o ./build/autodiscover-linux-amd64 ./cmd/autodiscover
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o ./build/autodiscover-linux-arm64 ./cmd/autodiscover

build-all: ## Build binaries for all platforms
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o ./build/autodiscover-linux-amd64  ./cmd/autodiscover
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o ./build/autodiscover-linux-arm64  ./cmd/autodiscover
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o ./build/autodiscover-darwin-amd64 ./cmd/autodiscover
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o ./build/autodiscover-darwin-arm64 ./cmd/autodiscover
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o ./build/autodiscover-windows-amd64.exe ./cmd/autodiscover
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o ./build/autodiscover-windows-arm64.exe ./cmd/autodiscover

test: ## Run tests
	go test -v ./...

clean: ## Remove built binaries
	rm -rf build/* autodiscover autodiscover-*

deps: ## Download and verify dependencies
	go mod download
	go mod verify

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...
