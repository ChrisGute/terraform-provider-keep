# keep_provider

Manages a provider in KeepHQ. This resource allows you to create, read, update, and delete providers in your KeepHQ instance.

## Example Usage

```hcl
resource "keep_provider" "datadog" {
  name = "production-datadog"
  type = "datadog"
  
  config = {
    api_key = "your-datadog-api-key"
    app_key = "your-datadog-app-key"
    # Additional provider-specific configuration
  }
}

output "provider_id" {
  value = keep_provider.datadog.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the provider. This is used to identify the provider in the KeepHQ UI and API.

* `type` - (Required, Forces new resource) The type of the provider. This determines what kind of service the provider connects to (e.g., "datadog", "newrelic", "pagerduty"). Once set, this cannot be changed without recreating the resource.

* `config` - (Required, Sensitive) A map of provider-specific configuration options. The keys and values depend on the provider type. This field is marked as sensitive and will not be displayed in logs or console output.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the provider in KeepHQ.

* `installed` - A boolean indicating whether the provider is successfully installed and connected.

* `last_alert_received` - The timestamp of the last alert received from this provider, if any.

## Import

Providers can be imported using their ID, e.g.,

```bash
terraform import keep_provider.example 12345abc-dead-beef-cafe-1234567890ab
```

## Provider-Specific Configuration

Different provider types require different configuration options in the `config` block. Below are examples for common provider types:

### Datadog

```hcl
resource "keep_provider" "datadog" {
  name = "production-datadog"
  type = "datadog"
  
  config = {
    api_key = "your-datadog-api-key"
    app_key = "your-datadog-app-key"
  }
}
```

### New Relic

```hcl
resource "keep_provider" "newrelic" {
  name = "production-newrelic"
  type = "newrelic"
  
  config = {
    api_key = "your-newrelic-api-key"
    account_id = "your-account-id"
    region = "US"  # or "EU"
  }
}
```

### PagerDuty

```hcl
resource "keep_provider" "pagerduty" {
  name = "production-pagerduty"
  type = "pagerduty"
  
  config = {
    api_key = "your-pagerduty-api-key"
    email = "your-email@example.com"
  }
}
```

## Best Practices

1. **Sensitive Data**: Always use environment variables or a secure secret management solution for sensitive values in the `config` block.

2. **Naming Conventions**: Use descriptive names that indicate the environment and purpose of the provider (e.g., `production-datadog`, `staging-slack`).

3. **Version Control**: Commit your Terraform configurations to version control, but use `.gitignore` to exclude files containing sensitive information.

4. **Provider Types**: Verify the supported provider types and their required configurations in the [KeepHQ documentation](https://keephq.dev/docs).

## Troubleshooting

### Provider Installation Fails

If a provider fails to install:

1. Verify that the `type` is spelled correctly and is a supported provider type.
2. Check that all required configuration options are provided in the `config` block.
3. Ensure that the API keys and other credentials are valid and have the necessary permissions.

### Provider Not Receiving Alerts

If the provider is installed but not receiving alerts:

1. Check the `installed` status of the provider in the Terraform state.
2. Verify that the provider's webhook URL is correctly configured in the external service.
3. Check the KeepHQ logs for any error messages related to the provider.

## Related Resources

* [KeepHQ Documentation](https://keephq.dev/docs)
* [Terraform Provider Documentation](https://www.terraform.io/docs/providers/keep/index.html)
