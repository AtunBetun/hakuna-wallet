BINARY_NAME=out
IMAGE_NAME?=hakuna-wallet
IMAGE_TAG?=latest
MIGRATE_BIN?=migrate
MIGRATIONS_DIR?=migrations
MIGRATE_DATABASE_URL?=$(DATABASE_URL)

# Run the binary
run: build
	@echo "Running..."
	./$(BINARY_NAME) | jq

# Build the Go project
build:
	@echo "Building..."
	cd src && go build -o ../$(BINARY_NAME) cmd/ticket_generator/main.go

test:
	@echo "Testing"
	cd src && go test ./...

# Clean up built files
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)

docker-build:
	@echo "Building Docker image..."
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) . --progress plain

migrate-up:
	@echo "Running database migrations..."
	@if [ -z "$(MIGRATE_DATABASE_URL)" ]; then \
		echo "Error: set MIGRATE_DATABASE_URL or export DATABASE_URL"; \
		exit 1; \
	fi
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(MIGRATE_DATABASE_URL)" up

migrate-down:
	@echo "Rolling back last database migration..."
	@if [ -z "$(MIGRATE_DATABASE_URL)" ]; then \
		echo "Error: set MIGRATE_DATABASE_URL or export DATABASE_URL"; \
		exit 1; \
	fi
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(MIGRATE_DATABASE_URL)" down 1

.PHONY: build run clean deps docker-build migrate-up migrate-down
