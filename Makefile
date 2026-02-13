.PHONY: build run serve setup chat test lint clean install init help

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"
CONFIG  ?= $(wildcard config.yaml)
VERBOSE ?=

# Build flags for serve
SERVE_FLAGS :=
ifneq ($(CONFIG),)
  SERVE_FLAGS += --config $(CONFIG)
endif
ifneq ($(VERBOSE),)
  SERVE_FLAGS += -v
endif

## build: Build the binary
build:
	go build $(LDFLAGS) -o bin/copilot ./cmd/copilot

## run: Build and start copilot serve
run: build
	./bin/copilot serve $(SERVE_FLAGS)

## serve: Alias for run
serve: run

## setup: Interactive setup wizard
setup: build
	./bin/copilot setup

## init: Create default config.yaml (non-interactive)
init: build
	./bin/copilot config init

## validate: Validate the configuration
validate: build
	./bin/copilot config validate $(if $(CONFIG),--config $(CONFIG))

## chat: Send a single message (usage: make chat MSG="hello")
chat: build
	./bin/copilot chat "$(MSG)"

## test: Run tests
test:
	go test ./... -v -race

## lint: Run linter
lint:
	golangci-lint run ./...

## clean: Remove build artifacts
clean:
	rm -rf bin/ dist/

## install: Install binary to GOPATH
install:
	go install $(LDFLAGS) ./cmd/copilot

## docker-build: Build Docker image
docker-build:
	docker compose build

## docker-up: Start via Docker Compose
docker-up:
	docker compose up -d

## docker-down: Stop containers
docker-down:
	docker compose down

## help: Show available commands
help:
	@echo "Usage:"
	@echo "  make setup             # Interactive setup wizard"
	@echo "  make run               # Build + serve (auto-detects config.yaml)"
	@echo "  make run VERBOSE=1     # Build + serve with debug logs"
	@echo "  make run CONFIG=x.yaml # Build + serve with specific config"
	@echo "  make init              # Create default config.yaml (non-interactive)"
	@echo "  make validate          # Validate configuration"
	@echo "  make chat MSG=\"hello\"  # Send a single message"
	@echo ""
	@echo "All commands:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | sort
