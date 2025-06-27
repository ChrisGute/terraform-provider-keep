terraform {
  required_providers {
    keep = {
      source  = "keephq/keep"
      version = "~> 0.1"
    }
  }
}

# Configure the KeepHQ Provider
provider "keep" {
  # API key for KeepHQ. Can also be set via KEEP_API_KEY environment variable
  api_key = "your-keephq-api-key"
  
  # URL of the KeepHQ API. Defaults to http://localhost:8080
  # api_url = "https://your-keephq-instance.com"
}

# Example: Create a provider
# resource "keep_provider" "example" {
#   type = "datadog"
#   name = "example-datadog"
#   config = {
#     api_key = "your-datadog-api-key"
#     app_key = "your-datadog-app-key"
#   }
# }

# Example: Create an alert
# resource "keep_alert" "example" {
#   name     = "example-alert"
#   severity = "high"
#   message  = "This is an example alert"
#   labels = {
#     environment = "production"
#     team        = "devops"
#   }
# }
