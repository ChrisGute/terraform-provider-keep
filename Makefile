.PHONY: build install test testacc gen-docs clean

# Build the provider
build:
	@echo "==> Building the provider..."
	@go build -o terraform-provider-keep

# Install the provider
install: build
	@echo "==> Installing the provider..."
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/keephq/keep/0.1.0/darwin_$(shell uname -m)
	@cp terraform-provider-keep ~/.terraform.d/plugins/registry.terraform.io/keephq/keep/0.1.0/darwin_$(shell uname -m)/

# Run tests
test:
	@echo "==> Testing..."
	@go test -v ./...

# Run acceptance tests
testacc:
	@echo "==> Running acceptance tests..."
	@TF_ACC=1 go test -v ./...

# Generate documentation
gen-docs:
	@echo "==> Generating documentation..."
	@go generate ./...

# Clean build artifacts
clean:
	@echo "==> Cleaning..."
	@rm -f terraform-provider-keep

# Default target
all: build
