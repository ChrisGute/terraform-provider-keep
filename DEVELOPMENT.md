# Development Guide

This guide explains how to set up and work with the KeepHQ Terraform Provider during early development.

## Prerequisites

- Go 1.21 or later
- Terraform 1.0 or later
- Git

## Local Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/ChrisGute/terraform-provider-keep.git
   cd terraform-provider-keep
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

## Building the Provider

Build the provider for your local platform:

```bash
go build -o terraform-provider-keep
```

## Testing

### Unit Tests

```bash
go test -v ./...
```

### Acceptance Tests

1. Set up your environment variables:
   ```bash
   export KEEP_API_KEY=your_api_key
   export KEEP_API_URL=http://localhost:3000  # or your KeepHQ instance
   ```

2. Run acceptance tests:
   ```bash
   TF_ACC=1 go test -v ./...
   ```

## Using the Provider Locally

1. Create a `dev.tfrc` file in your home directory:
   ```hcl
   provider_installation {
     dev_overrides {
       "local/keep/keep" = "/path/to/your/terraform-provider-keep"
     }
     direct {}
   }
   ```

2. Set the environment variable:
   ```bash
   export TF_CLI_CONFIG_FILE=~/.terraformrc
   ```

3. Create a Terraform configuration that uses the provider:
   ```hcl
   terraform {
     required_providers {
       keep = {
         source = "local/keep/keep"
         version = "0.1.0"
       }
     }
   }
   ```

## Versioning

For early development (v0.1.0), we're using simple version tags. To create a new version:

```bash
git tag v0.1.0
git push origin v0.1.0
```

## Next Steps

As the provider matures, we'll:
1. Set up automated releases with GoReleaser
2. Add multi-platform builds
3. Configure signing for releases
4. Publish to the Terraform Registry
