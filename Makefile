.PHONY: help build test test-cov clean run run-cli dev fmt lint migrate libs download-libs

help:
	@echo "Available targets:"
	@echo "  make build         - Build binaries"
	@echo "  make test          - Run tests"
	@echo "  make test-cov      - Run tests with coverage"
	@echo "  make dev           - Run with hot reload"
	@echo "  make run           - Run MCP server"
	@echo "  make run-cli       - Run CLI"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make fmt           - Format code"
	@echo "  make lint          - Run linter"
	@echo "  make migrate       - Run migrations"
	@echo "  make libs          - Download ONNX Runtime libraries"

build:
	@echo "Building..."
	@mkdir -p bin
	go build -o bin/memoos-server ./cmd/server
	go build -o bin/memoos-cli ./cmd/cli

test:
	go test -v ./...

test-cov:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

dev:
	@echo "Hot reload requires air: go install github.com/cosmtrek/air@latest"
	which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

run:
	go run ./cmd/server

run-cli:
	go run ./cmd/cli \$(ARGS)

clean:
	rm -rf bin/ coverage.out

fmt:
	go fmt ./...

lint:
	@golangci-lint --version > /dev/null || (echo "Installing golangci-lint..." && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(shell go env GOPATH)/bin)
	golangci-lint run

download-libs:
	@echo "Downloading ONNX Runtime libraries..."
	@mkdir -p libs
	@ORT_VERSION="1.26.0"; \
	OS=$$(uname -s); \
	ARCH=$$(uname -m); \
	case "$$OS" in \
		"Linux") \
			if [ "$$ARCH" = "aarch64" ]; then PLATFORM="linux-aarch64"; else PLATFORM="linux-x64"; fi; \
			URL="https://github.com/microsoft/onnxruntime/releases/download/v$$ORT_VERSION/onnxruntime-$$PLATFORM-$$ORT_VERSION.tgz"; \
			echo "Fetching $$URL..."; \
			curl -L $$URL -o libs/ort.tgz; \
			tar -xzf libs/ort.tgz -C libs; \
			mv libs/onnxruntime-$$PLATFORM-*/lib/libonnxruntime.so.$$ORT_VERSION libs/; \
			ln -sf libonnxruntime.so.$$ORT_VERSION libs/libonnxruntime.so; \
			rm -rf libs/onnxruntime-$$PLATFORM-* libs/ort.tgz; \
			;; \
		"Darwin") \
			if [ "$$ARCH" = "arm64" ]; then PLATFORM="osx-arm64"; else PLATFORM="osx-x86_64"; fi; \
			URL="https://github.com/microsoft/onnxruntime/releases/download/v$$ORT_VERSION/onnxruntime-$$PLATFORM-$$ORT_VERSION.tgz"; \
			echo "Fetching $$URL..."; \
			curl -L $$URL -o libs/ort.tgz; \
			tar -xzf libs/ort.tgz -C libs; \
			mv libs/onnxruntime-$$PLATFORM-*/lib/libonnxruntime.dylib libs/; \
			rm -rf libs/onnxruntime-$$PLATFORM-* libs/ort.tgz; \
			;; \
		*) \
			PLATFORM="win-x64"; \
			URL="https://github.com/microsoft/onnxruntime/releases/download/v$$ORT_VERSION/onnxruntime-$$PLATFORM-$$ORT_VERSION.zip"; \
			echo "Fetching $$URL..."; \
			curl -L $$URL -o libs/ort.zip; \
			unzip -q libs/ort.zip -d libs; \
			mv libs/onnxruntime-$$PLATFORM-*/onnxruntime.dll libs/; \
			rm -rf libs/onnxruntime-$$PLATFORM-* libs/ort.zip; \
			;; \
	esac
	@echo "Done."
