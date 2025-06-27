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

resource "keep_extraction_rule" "test" {
  name        = "test-rule"
  description = "Test extraction rule created by acceptance test"
  attribute   = "message"
  regex       = "test-([A-Z0-9]+)"
  priority    = 1
  disabled    = false
  pre         = false
}

output "extraction_rule_id" {
  value = keep_extraction_rule.test.id
}
