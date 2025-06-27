variable "api_key" {
  type        = string
  description = "KeepHQ API key"
  sensitive   = true
}

variable "api_url" {
  type        = string
  description = "KeepHQ API URL"
  default     = "http://localhost:8080"
}

provider "keep" {
  api_key = var.api_key
  api_url = var.api_url
}

resource "keep_provider" "test" {
  name = "test-provider"
  type = "slack"
  config = {
    webhook_url = "https://hooks.slack.com/services/xxx/yyy/zzz"
    channel     = "#alerts"
  }
}

output "provider_id" {
  value = keep_provider.test.id
}
