BINARY_NAME=out
IMAGE_NAME?=hakuna-wallet
IMAGE_TAG?=latest

# Run the binary
run: build
	@echo "Running..."
	./$(BINARY_NAME) | jq

# Build the Go project
build:
	@echo "Building..."
	cd src && go build -o ../$(BINARY_NAME) cmd/batch/main.go

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

.PHONY: build run clean deps docker-build
