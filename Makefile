.PHONY: help build generate install test test-integration test-all coverage clean fmt vet lint

help:
	@echo "Available targets:"
	@echo "  build            Build the coffer binary"
	@echo "  generate         Generate version info"
	@echo "  install          Install coffer to /usr/local/bin"
	@echo "  test             Run unit tests"
	@echo "  test-integration Run integration tests"
	@echo "  test-all         Run all tests"
	@echo "  coverage         Generate coverage report"
	@echo "  clean            Remove build artifacts"
	@echo "  fmt              Format code"
	@echo "  vet              Run static analysis"
	@echo "  lint             Run fmt and vet"

build: generate
	@go build -o ./bin/coffer ./cmd/coffer

generate:
	@go generate ./...

install: build
	sudo cp ./bin/coffer /usr/local/bin/coffer

test: generate
	@go test ./...

test-integration: generate
	@go test -tags=integration ./...

test-all: test-integration

coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

clean:
	rm -rf ./bin coverage.out

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet
