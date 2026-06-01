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
	@echo "  make download-libs - Download ONNX Runtime libraries"

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
	go run ./cmd/cli $(ARGS)

clean:
	rm -rf bin/ coverage.out

fmt:
	go fmt ./...

lint:
	@golangci-lint --version > /dev/null || (echo "Installing golangci-lint..." && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin)
	golangci-lint run

libs: download-libs

download-libs:
	@echo "Downloading ONNX Runtime libraries..."
	@mkdir -p libs
	@if [ "$(shell uname -s)" = "Linux" ]; then \
		if [ "$(shell uname -m)" = "aarch64" ]; then \
			curl -L https://github.com/microsoft/onnxruntime/releases/download/v1.17.1/onnxruntime-linux-aarch64-1.17.1.tgz -o libs/onnxruntime.tgz; \
		else \
			curl -L https://github.com/microsoft/onnxruntime/releases/download/v1.17.1/onnxruntime-linux-x64-1.17.1.tgz -o libs/onnxruntime.tgz; \
		fi \
		tar -xzf libs/onnxruntime.tgz -C libs; \
		mv libs/onnxruntime-linux-*/lib/libonnxruntime.so.1.17.1 libs/ 2>/dev/null || true; \
		rm -rf libs/onnxruntime-linux-* libs/onnxruntime.tgz; \
		ln -sf libonnxruntime.so.1.17.1 libs/libonnxruntime.so; \
		echo "Downloaded: libs/libonnxruntime.so"; \
	elif [ "$(shell uname -s)" = "Darwin" ]; then \
		if [ "$(shell uname -m)" = "arm64" ]; then \
			curl -L https://github.com/microsoft/onnxruntime/releases/download/v1.17.1/onnxruntime-osx-arm64-1.17.1.tgz -o libs/onnxruntime.tgz; \
		else \
			curl -L https://github.com/microsoft/onnxruntime/releases/download/v1.17.1/onnxruntime-osx-x86_64-1.17.1.tgz -o libs/onnxruntime.tgz; \
		fi \
		tar -xzf libs/onnxruntime.tgz -C libs; \
		mv libs/onnxruntime-osx-*/lib/libonnxruntime.dylib libs/ 2>/dev/null || true; \
		rm -rf libs/onnxruntime-osx-* libs/onnxruntime.tgz; \
		echo "Downloaded: libs/libonnxruntime.dylib"; \
	else \
		curl -L https://github.com/microsoft/onnxruntime/releases/download/v1.17.1/onnxruntime-win-x64-1.17.1.zip -o libs/onnxruntime.zip; \
		unzip -q libs/onnxruntime.zip -d libs; \
		mv libs/onnxruntime-win-x64-*/onnxruntime.dll libs/ 2>/dev/null || true; \
		rm -rf libs/onnxruntime-win-x64-* libs/onnxruntime.zip; \
		echo "Downloaded: libs/onnxruntime.dll"; \
	fi