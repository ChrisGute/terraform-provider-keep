# keep_extraction_rule

Manages an extraction rule in Keep. Extraction rules define how to extract specific attributes from alerts using regular expressions.

## Example Usage

```hcl
# Basic extraction rule
resource "keep_extraction_rule" "example" {
  name        = "extract-service-name"
  description = "Extract service name from alert name"
  attribute   = "service"
  regex       = "service-(?P<service>[a-z-]+)-alert"
  priority    = 10
  disabled    = false
  pre         = false
  condition   = "alert.name.matches('service-.*-alert')"
}

# Pre-processing extraction rule
resource "keep_extraction_rule" "preprocess" {
  name      = "normalize-severity"
  attribute = "severity"
  regex     = "(?i)crit(ical)?"
  priority  = 5
  pre       = true
  disabled  = false
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the extraction rule. Must be unique across all extraction rules.

* `description` - (Optional) A description of what the extraction rule does.

* `attribute` - (Required) The name of the attribute to extract from the alert.

* `regex` - (Required) The regular expression pattern used to extract the attribute value. The pattern should use named capture groups (e.g., `(?P<name>pattern)`) to extract specific values.

* `priority` - (Optional) The priority of the rule. Rules with lower numbers are evaluated first. Defaults to `10`.

* `disabled` - (Optional) Whether the extraction rule is disabled. Defaults to `false`.

* `pre` - (Optional) Whether this is a pre-processing rule that runs before other rules. Defaults to `false`.

* `condition` - (Optional) A CEL (Common Expression Language) expression that determines when this rule should be applied. If not specified, the rule will always be applied.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the extraction rule.

* `created_at` - The timestamp when the extraction rule was created.

* `updated_at` - The timestamp when the extraction rule was last updated.

## Import

Extraction rules can be imported using their ID:

```
$ terraform import keep_extraction_rule.example 123
```

## Best Practices

1. **Use named capture groups** in your regex patterns to make the extracted values more meaningful.

2. **Set appropriate priorities** to control the order in which rules are evaluated.

3. **Use conditions** to limit when extraction rules are applied, which can improve performance.

4. **Test your regex patterns** thoroughly to ensure they match expected inputs correctly.

5. **Use pre-processing rules** for normalizing data before other extraction rules run.
