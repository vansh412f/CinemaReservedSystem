BINARY_NAME := bin/server

.PHONY: all build run clean deps

all: build

build:
	@echo "🔨 Building..."
	@mkdir -p bin
	@go build -o $(BINARY_NAME) ./cmd/server

run:
	@echo "🚀 Running..."
	@go run ./cmd/server

clean:
	@echo "🧹 Cleaning..."
	@rm -rf bin
	@rm -f cinema.db

deps:
	@echo "📦 Downloading dependencies..."
	@go mod download