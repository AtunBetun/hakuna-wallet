BINARY_NAME=out

# Run the binary
run: build
	@echo "Running..."
	./$(BINARY_NAME)

# Build the Go project
build:
	@echo "Building..."
	cd src && go build -o ../$(BINARY_NAME) cmd/batch/main.go


# Clean up built files
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)

.PHONY: build run clean deps
