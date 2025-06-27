# Terraform Provider for KeepHQ

[![Go Reference](https://pkg.go.dev/badge/github.com/keephq/terraform-provider-keep.svg)](https://pkg.go.dev/github.com/keephq/terraform-provider-keep)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL_2.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
[![Tests](https://github.com/keephq/terraform-provider-keep/actions/workflows/test.yml/badge.svg)](https://github.com/keephq/terraform-provider-keep/actions)

This provider allows you to manage your [KeepHQ](https://keephq.dev) resources using Terraform. With this provider, you can manage your alert providers, alerts, and other KeepHQ resources as code.

> **Note**: This provider is currently in **beta**. The mapping rule resource is production-ready, while other resources are still under development.

## Features

- **Mapping Rules**: ✅ Production-ready - Manage your mapping rules as code
- **Extraction Rules**: ✅ Production-ready - Define and manage data extraction rules
- **Provider Management**: 🔧 In Development - Manage KeepHQ providers
- **Alert Management**: 🔧 In Development - Manage alerts and their configurations
- **Infrastructure as Code**: Define your KeepHQ resources in HCL and version control them
- **Integration with Terraform Workflows**: Use with Terraform Cloud, CI/CD pipelines, and more

## API Coverage

This section outlines the current coverage of the KeepHQ API by this Terraform provider.

### ✅ Production Ready

| API Endpoint | Resource | CRUD Operations | Notes |
|--------------|----------|-----------------|-------|
| `/mapping` | `keep_mapping_rule` | ✅ Create<br>✅ Read<br>✅ Update<br>✅ Delete | Full mapping rule lifecycle management |
| `/extraction` | `keep_extraction_rule` | ✅ Create<br>✅ Read<br>✅ Update<br>✅ Delete | Full extraction rule lifecycle management |
| `/providers` | `keep_provider` | ✅ Create<br>✅ Read<br>✅ Update<br>✅ Delete | Provider management |

### 🔧 In Development (Experimental)

| API Endpoint | Resource | Status | Notes |
|--------------|----------|--------|-------|
| `/alerts` | `keep_alert` | ⚠️ Experimental | Basic alert management |

### 📅 Planned

| API Endpoint | Resource | Priority | Notes |
|--------------|----------|----------|-------|
| `/workflows` | `keep_workflow` | High | Workflow automation |
| `/incidents` | `keep_incident` | High | Incident management |
| `/rules` | `keep_rule` | Medium | Alert routing rules |
| `/dashboard` | `keep_dashboard` | Medium | Dashboard management |
| `/settings` | `keep_setting` | Low | System and tenant settings |
| `/tags` | `keep_tag` | Low | Resource tagging |
| `/ai` | `keep_ai` | Low | AI-related features |
| `/auth` | `keep_auth` | Medium | Authentication endpoints |
| `/cel` | `keep_cel` | Low | CEL expression evaluation |
| `/deduplications` | `keep_deduplication` | Medium | Deduplication rules |
| `/facets` | `keep_facet` | Low | Faceted search |
| `/healthcheck` | - | Low | Health check endpoint |
| `/maintenance` | `keep_maintenance` | Low | Maintenance windows |
| `/metrics` | - | Low | System metrics |
| `/preset` | `keep_preset` | Low | Preset configurations |
| `/provider_images` | - | Low | Provider-specific images |
| `/pusher` | - | Low | Push notifications |
| `/status` | - | Low | System status |
| `/topology` | `keep_topology` | Low | System topology |
| `/whoami` | - | Low | Current user information |

> **Legend**:
> - ✅ Complete: Full CRUD support
> - 🔄 Partially Implemented: Some operations supported
> - 📅 Planned: In development pipeline
> - ❌ Not Started: No implementation yet

## Known Limitations

- The `disabled` field in mapping rules is currently not supported by the KeepHQ API. The field exists in the provider schema for future compatibility but will be ignored by the API. All rules are effectively always enabled.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (to build the provider plugin)

## Installation

### Local Installation (Recommended for Development)

To use the provider from your local development environment:

1. Build the provider from source:
   ```bash
   git clone https://github.com/ChrisGute/terraform-provider-keep.git
   cd terraform-provider-keep
   make build
   make install
   ```

2. Create a `versions.tf` file with the local provider:
   ```hcl
   terraform {
     required_providers {
       keep = {
         source  = "local/keep/keep"
         version = "0.1.0"
       }
     }
   }
   
   provider "keep" {
     api_key = "your-api-key-here"
     api_url = "http://localhost:3000"  # Update with your KeepHQ URL
   }
   ```

3. Run `terraform init` to initialize the provider.

### Using from GitHub (For Production)

To use this provider directly from the GitHub repository:

1. Create a `versions.tf` file with the GitHub source:
   ```hcl
   terraform {
     required_providers {
       keep = {
         source  = "ChrisGute/keep"
         version = "~> 0.1"
       }
     }
   }
   
   provider "keep" {
     api_key = "your-api-key-here"
     api_url = "https://your-keephq-instance.com"
   }
   ```

2. Run `terraform init` to install the provider from GitHub.

### Building from Source

1. Clone the repository:
   ```bash
   git clone git@github.com:ChrisGute/terraform-provider-keep.git
   cd terraform-provider-keep
   ```

2. Build the provider:
   ```bash
   make build
   ```

3. Install the provider locally:
   ```bash
   make install
   ```

   This will install the provider in the Terraform plugins directory for your user account.

## Configuration

### Provider Configuration

Configure the provider in your Terraform configuration:

```hcl
provider "keep" {
  # API key for authentication (required)
  api_key = "your-keephq-api-key"
  
  # API URL (optional, defaults to http://localhost:8080)
  # api_url = "https://your-keephq-instance.com"
}
```

### Environment Variables

You can also configure the provider using environment variables:

- `KEEP_API_KEY`: Your KeepHQ API key (required)
- `KEEP_API_URL`: KeepHQ API URL (optional, defaults to http://localhost:8080)

Example:

```bash
export KEEP_API_KEY="your-api-key-here"
export KEEP_API_URL="https://your-keephq-instance.com"
```

## Usage Examples

### Mapping Rule Example

```hcl
resource "keep_mapping_rule" "example" {
  name        = "example-mapping-rule"
  description = "Example mapping rule for production environment"
  priority    = 10
  
  # Matchers are used to filter which alerts this rule applies to
  matchers = {
    env  = "production"
    team = "sre"
  }

  # CSV data defines the mapping logic
  csv_data = <<-EOT
env,team,owner,severity
production,sre,alice,critical
staging,dev,bob,warning
  EOT
}
```

### Extraction Rule Example

```hcl
resource "keep_extraction_rule" "example" {
  name        = "extract-http-status"
  description = "Extract HTTP status code from log messages"
  
  # Pattern to match in the message
  pattern = "status=(?P<http_status>\\d+)"
  
  # Optional: Only apply this rule to specific providers
  # provider_id = "your-provider-id"
  
  # Optional: Add tags to extracted fields
  tags = {
    source = "http-logs"
    type   = "regex-extraction"
  }
}
```

### Provider Configuration Example

```hcl
resource "keep_provider" "example" {
  name        = "production-grafana"
  type        = "grafana"
  description = "Production Grafana instance"
  
  # Provider-specific configuration
  config = {
    url      = "https://grafana.example.com"
    username = "api-user"
    password = var.grafana_api_key
  }
  
  # Optional: Provider tags
  tags = {
    environment = "production"
    team        = "observability"
  }
}
```

## Authentication

### API Key

You'll need a KeepHQ API key to authenticate with the API. You can obtain this from your KeepHQ instance under Settings > API Keys.

### Authentication Methods

1. **Provider Configuration** (recommended for most cases):
   ```hcl
   provider "keep" {
     api_key = "your-api-key-here"
     api_url = "https://your-keephq-instance.com"
   }
   ```

2. **Environment Variables**:
   ```bash
   export KEEP_API_KEY="your-api-key-here"
   export KEEP_API_URL="https://your-keephq-instance.com"
   ```

3. **Terraform Variables** (for CI/CD):
   ```hcl
   variable "keep_api_key" {
     description = "KeepHQ API key"
     type        = string
     sensitive   = true
   }

   provider "keep" {
     api_key = var.keep_api_key
   }
   ```

   Then pass the key when running Terraform:
   ```bash
   terraform apply -var="keep_api_key=your-api-key-here"
   ```

## Troubleshooting

### Extraction Rule Creation Fails with HTML Response

If you encounter an error like this when creating an extraction rule:

```
Error: Could not create extraction rule, unexpected error: error parsing extraction rule response:
invalid character '<' looking for beginning of value
```

This typically means the API returned an HTML error page instead of JSON. Common causes include:

1. **Incorrect API URL**: Ensure the `api_url` in your provider configuration points to the correct KeepHQ instance.
2. **Authentication Failure**: Verify your API key is correct and has the necessary permissions.
3. **Server Error**: The KeepHQ server might be experiencing issues. Check the server logs for more details.
4. **Network Issues**: Ensure there are no network connectivity problems between your machine and the KeepHQ server.

### Debugging API Requests

To help diagnose issues, you can enable debug logging:

```hcl
provider "keep" {
  api_key = "your-api-key-here"
  api_url = "http://localhost:3000"
  
  # Enable debug logging
  debug = true
}
```

This will log detailed information about API requests and responses to help identify issues.

### Common Issues and Solutions

#### 1. Provider Not Found

If you see an error like:
```
Could not find required providers, but found possible local directory for "local/keep/keep"
```

Make sure you've built and installed the provider locally:
```bash
make build
make install
```

#### 2. Authentication Errors

If you receive authentication errors, verify:
- The API key is correct and not expired
- The API key has the necessary permissions
- The `api_url` is correctly set

#### 3. Resource Creation Fails

If resource creation fails, check:
- All required fields are provided
- Field values match the expected format
- The server logs for detailed error messages

## Development

### Running Tests

1. Set up your environment variables:
   ```bash
   export KEEP_API_KEY="your-api-key"
   export KEEP_API_URL="https://api.keephq.dev"
   ```

2. Run the tests:
   ```bash
   make test
   ```

   Or run acceptance tests (requires a running KeepHQ instance):
   ```bash
   TF_ACC=1 make testacc
   ```

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

This project is licensed under the MPL 2.0 License - see the [LICENSE](LICENSE) file for details.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.20 (to build the provider plugin)
- [KeepHQ](https://keephq.dev) account and API key

> **Note**: The mapping rule resource requires a KeepHQ API key with appropriate permissions.

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most common operations:

```sh
go get -u
go mod tidy
go mod vendor
```

## Using the Provider

```hcl
provider "keep" {
  # Configuration options
  api_key = "your-keephq-api-key"
  api_url = "https://your-keephq-instance.com" # Optional, defaults to http://localhost:8080
}
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
make testacc
```

## License

This project is licensed under the [Mozilla Public License 2.0](LICENSE).
