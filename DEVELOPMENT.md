# Development Guide

This guide explains how to set up, develop, and release the KeepHQ Terraform Provider.

## Prerequisites

- Go 1.21 or later
- Terraform 1.0 or later
- Git
- GPG (for release signing)
- GoReleaser (for releases)

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

Run all unit tests (excludes acceptance tests):

```bash
go test -v -short ./...
```

### Acceptance Tests

1. Set up your environment variables:
   ```bash
   export KEEP_API_KEY=your_api_key
   export KEEP_API_URL=http://localhost:3000  # or your KeepHQ instance URL
   ```

2. Run acceptance tests (takes longer as it creates real resources):
   ```bash
   TF_ACC=1 go test -v ./...
   ```

## Using the Provider Locally

### For Development

1. Build the provider:
   ```bash
   go build -o terraform-provider-keep
   ```

2. Create or update your Terraform CLI config (`~/.terraformrc`):
   ```hcl
   provider_installation {
     dev_overrides {
       "local/keep/keep" = "/path/to/terraform-provider-keep"
     }
     direct {}
   }
   ```

3. In your Terraform configuration:
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

### Using Released Versions

For production use, reference the provider from the Terraform Registry:

```hcl
terraform {
  required_providers {
    keep = {
      source  = "chrisgute/keep"
      version = "~> 0.1.0"
    }
  }
}

provider "keep" {
  api_key = var.keep_api_key
  url     = var.keep_api_url
}
```

## Release Process

### Prerequisites

1. Install GoReleaser:
   ```bash
   brew install goreleaser
   ```

2. Ensure you have a GPG key set up for signing releases.

### Creating a Release

1. Update the version in `main.go` and any other relevant files.

2. Update `CHANGELOG.md` with the changes for the new version.

3. Commit your changes:
   ```bash
   git add .
   git commit -m "Prepare for vX.Y.Z"
   ```

4. Create an annotated tag:
   ```bash
   git tag -a vX.Y.Z -m "vX.Y.Z"
   ```

5. Push the tag to trigger the GitHub Actions release workflow:
   ```bash
   git push origin vX.Y.Z
   ```

6. The GitHub Actions workflow will:
   - Run tests
   - Build binaries for all supported platforms
   - Generate SHA256 checksums
   - Sign the checksums file
   - Create a GitHub release with all artifacts

### Manual Release (if needed)

If you need to create a release manually:

```bash
goreleaser release --clean
```

## Supported Platforms

The provider is built for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

## GPG Signing

All releases are signed with GPG for verification. The public key is available in the repository at `.github/gpg/terraform-provider-keep-ci.pub`.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for your changes
5. Run the test suite
6. Submit a pull request

## Troubleshooting

### GPG Signing Issues

If you encounter GPG signing issues:
1. Ensure `gpg` is installed and in your PATH
2. Verify your GPG key is available: `gpg --list-secret-keys`
3. Check file permissions in `~/.gnupg`

### Acceptance Test Failures

- Ensure your API key has the necessary permissions
- Verify the API URL is correct and accessible
- Check for any rate limiting on the API
