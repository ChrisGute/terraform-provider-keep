# KeepHQ Mapping Rule Resource

This resource allows you to manage mapping rules in KeepHQ. Mapping rules are used to enrich alerts with additional data from CSV files or topology data.

## Example Usage

```hcl
resource "keep_mapping_rule" "example" {
  name        = "example-mapping-rule"
  description = "Example mapping rule"
  priority    = 10
  
  matchers = {
    env  = "production"
    team = "sre"
  }

  csv_data = <<-EOT
env,team,owner,severity
production,sre,alice,critical
staging,dev,bob,warning
  EOT
}
```

## Argument Reference

* `name` - (Required) The name of the mapping rule.
* `description` - (Optional) A description of what the mapping rule does.
* `priority` - (Optional) The priority of the mapping rule. Lower numbers have higher priority. Defaults to `0`.
* `matchers` - (Optional) A map of matchers that determine when this rule should be applied.
* `csv_data` - (Optional) The CSV data to use for mapping. Each row should contain the matcher values and the fields to add to matching alerts.

### Notes

* The `disabled` field is currently not supported by the KeepHQ API and will be ignored. This is a known limitation documented in [issue #123](https://github.com/keephq/keep/issues/123).
* When importing existing mapping rules, the `csv_data` field may have formatting differences from what was originally provided. The provider normalizes this data, but you may see differences in whitespace or quoting when comparing the original and imported values.

## Import

Mapping rules can be imported using their ID:

```bash
terraform import keep_mapping_rule.example 123e4567-e89b-12d3-a456-426614174000
```

## Known Limitations

* The `disabled` field is currently not supported by the KeepHQ API. The field exists in the provider schema for future compatibility but will be ignored by the API. All rules are effectively always enabled.
