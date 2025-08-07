# GoWaveform Makefile
.PHONY: build test clean install examples help

# Default target
all: build test

# Build the CLI application
build:
	@echo "Building GoWaveform CLI..."
	go build -o gowaveform main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./waveform

# Build and run examples
examples: build
	@echo "Running library examples..."
	go run examples/comprehensive_demo.go

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -f gowaveform example_usage
	rm -f *.svg
	rm -f test/import_demo

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. ./waveform

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build -o dist/gowaveform-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -o dist/gowaveform-darwin-amd64 main.go
	GOOS=windows GOARCH=amd64 go build -o dist/gowaveform-windows-amd64.exe main.go

# Generate documentation
docs:
	@echo "Generating documentation..."
	go doc -all ./waveform

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the CLI application"
	@echo "  test       - Run unit tests"
	@echo "  examples   - Build and run library examples"
	@echo "  clean      - Clean build artifacts"
	@echo "  install    - Install dependencies"
	@echo "  bench      - Run benchmarks"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  docs       - Generate documentation"
	@echo "  help       - Show this help message"
