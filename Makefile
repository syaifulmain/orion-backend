APP_NAME = orion-backend
MAIN_PATH = cmd/api/main.go
BIN_DIR = bin
BIN_PATH = $(BIN_DIR)/api

.PHONY: all build run clean tidy help gen-key docs

all: build

## build: Compile the application binary
build:
	@echo "==> Building binary $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_PATH) $(MAIN_PATH)
	@echo "==> Done: $(BIN_PATH)"

## run: Run the application directly
run:
	@echo "==> Running $(APP_NAME)..."
	@go run $(MAIN_PATH)

## gen-key: Generate a short HMAC-SHA256 API key (e.g. make gen-key NAME=client-name DAYS=30)
gen-key:
	@go run cmd/genkey/main.go -name $(if $(NAME),$(NAME),local-client) -expires-in-days $(if $(DAYS),$(DAYS),30)

## tidy: Download and prune Go module dependencies
tidy:
	@echo "==> Tidying Go modules..."
	@go mod tidy

## clean: Remove built binary files
clean:
	@echo "==> Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@echo "==> Done."

## docs: Generate Swagger API documentation files
docs:
	@echo "==> Generating Swagger docs..."
	@swag init --parseDependency -g $(MAIN_PATH) -o docs
	@echo "==> Done. Docs generated at docs/"

## help: Display available Makefile commands
help:
	@echo "Available commands:"
	@echo "  make run                   - Run the application server"
	@echo "  make build                 - Compile binary to $(BIN_PATH)"
	@echo "  make docs                  - Generate Swagger API documentation"
	@echo "  make gen-key NAME=<name>   - Generate a Bearer API key"
	@echo "  make tidy                  - Run 'go mod tidy'"
	@echo "  make clean                 - Remove build directory ($(BIN_DIR))"
	@echo "  make help                  - Display this help message"
