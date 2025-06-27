package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
)

// getStringPreview returns a preview of a string with the specified max length
func getStringPreview(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// getMapKeys returns a slice of keys from a map
func getMapKeys(m interface{}) []string {
	switch v := m.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		return keys
	case map[string]attr.Value:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		return keys
	default:
		return []string{fmt.Sprintf("unsupported type: %T", m)}
	}
}
