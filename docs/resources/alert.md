# keep_alert

Manages an alert in KeepHQ.

## Example Usage

```hcl
resource "keep_alert" "example" {
  name        = "example-alert"
  status      = "firing"
  severity    = "high"
  environment = "production"
  service     = "api-service"
  message     = "High error rate detected"
  description = "The error rate for the API service has exceeded the threshold"
  url         = "https://example.com/alerts/123"
  image_url   = "https://example.com/images/alert.png"
  
  labels = {
    "team"       = "sre"
    "environment" = "production"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the alert.
* `status` - (Optional) The status of the alert. Must be one of: `firing`, `resolved`, `acknowledged`, `suppressed`, `pending`. Default is `firing`.
* `severity` - (Optional) The severity of the alert. Must be one of: `critical`, `high`, `warning`, `info`, `low`. Default is `low`.
* `environment` - (Optional) The environment of the alert.
* `service` - (Optional) The service associated with the alert.
* `message` - (Optional) The message of the alert.
* `description` - (Optional) A detailed description of the alert.
* `url` - (Optional) A URL to provide more information about the alert.
* `image_url` - (Optional) A URL to an image related to the alert.
* `labels` - (Optional) A map of key-value pairs to attach to the alert as labels.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the alert.
* `fingerprint` - The fingerprint of the alert (used for deduplication).

## Import

Alerts can be imported using their ID:

```bash
terraform import keep_alert.example alert-1234567890
```

## Notes

- The `fingerprint` attribute is automatically generated if not provided and is used for deduplication of alerts.
- When updating an alert, only the fields that are specified will be updated. Other fields will remain unchanged.
- The `labels` field can be used to attach arbitrary metadata to the alert as key-value pairs.
