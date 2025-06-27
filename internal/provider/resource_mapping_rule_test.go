// resource_mapping_rule_test.go - Tests for the mapping rule resource
package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keephq/terraform-provider-keep/internal/client"
)

// testAccCheckMappingRuleExists checks if a mapping rule exists
func testAccCheckMappingRuleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no mapping rule ID is set")
		}

		// Create a new client using the test configuration
		apiKey := os.Getenv("KEEP_API_KEY")
		apiURL := os.Getenv("KEEP_API_URL")
		if apiURL == "" {
			apiURL = "http://localhost:8080"
		}

		c, err := client.NewClient(apiURL, apiKey)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		// Verify the mapping rule exists by trying to retrieve it
		_, err = c.GetMappingRule(context.Background(), rs.Primary.ID)
		return err
	}
}

// normalizeCSV normalizes CSV data by trimming whitespace and normalizing line endings
func normalizeCSV(csvData string) string {
	if csvData == "" {
		return ""
	}
	// Trim whitespace and normalize line endings
	normalized := strings.TrimSpace(csvData)
	normalized = strings.ReplaceAll(normalized, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	return normalized
}

// testAccCheckCSVDataEqual checks if the CSV data in the state matches the expected value after normalization
func testAccCheckCSVDataEqual(resourceName, attributeName, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		actual, ok := rs.Primary.Attributes[attributeName]
		if !ok {
			return fmt.Errorf("attribute %s not found in %s", attributeName, resourceName)
		}

		normalizedExpected := normalizeCSV(expected)
		normalizedActual := normalizeCSV(actual)

		if normalizedExpected != normalizedActual {
			return fmt.Errorf("CSV data does not match. Expected:\n%s\n\nGot:\n%s", normalizedExpected, normalizedActual)
		}

		return nil
	}
}

// testAccCheckMappingRuleDestroy checks if a mapping rule has been destroyed
func testAccCheckMappingRuleDestroy(s *terraform.State) error {
	// Create a new client using the test configuration
	apiKey := os.Getenv("KEEP_API_KEY")
	apiURL := os.Getenv("KEEP_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	client, err := client.NewClient(apiURL, apiKey)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "keep_mapping_rule" {
			continue
		}

		// Try to find the mapping rule
		_, err := client.GetMappingRule(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("mapping rule %s still exists", rs.Primary.ID)
		}

		// Verify the error is because the rule doesn't exist
		errMsg := err.Error()
		if matched, _ := regexp.MatchString("mapping rule with ID .* not found", errMsg); !matched {
			return fmt.Errorf("unexpected error: %s", errMsg)
		}
	}

	return nil
}

func TestAccMappingRuleResource(t *testing.T) {
	// Skip if acceptance testing is not enabled
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test as TF_ACC is not set")
	}

	// Ensure the test is run in sequence
	t.Parallel()

	// Test data for the mapping rule
	ruleName := "test-mapping-rule"
	updatedRuleName := "test-mapping-rule-updated"
	description := "Test mapping rule created by Terraform"
	updatedDescription := "Updated test mapping rule"

	// Test case for creating a basic mapping rule
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckMappingRuleDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccMappingRuleResourceBasic(ruleName, description, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMappingRuleExists("keep_mapping_rule.test"),
					resource.TestCheckResourceAttr("keep_mapping_rule.test", "name", ruleName),
					resource.TestCheckResourceAttr("keep_mapping_rule.test", "description", description),
					resource.TestCheckResourceAttr("keep_mapping_rule.test", "priority", "10"),
					resource.TestCheckResourceAttr("keep_mapping_rule.test", "matchers.%", "2"),
					resource.TestCheckResourceAttr("keep_mapping_rule.test", "matchers.env", "production"),
					resource.TestCheckResourceAttr("keep_mapping_rule.test", "matchers.team", "sre"),
					resource.TestMatchResourceAttr("keep_mapping_rule.test", "id", regexp.MustCompile(`^[0-9a-fA-F-]+$`)),
				),
			},
			// ImportState testing
			{
				ResourceName:            "keep_mapping_rule.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"csv_data"}, // Ignore csv_data as it may have formatting differences
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMappingRuleExists("keep_mapping_rule.test"),
					// Verify other important fields that should be preserved
					resource.TestCheckResourceAttr("keep_mapping_rule.test", "name", ruleName),
					resource.TestCheckResourceAttr("keep_mapping_rule.test", "description", description),
					resource.TestCheckResourceAttr("keep_mapping_rule.test", "priority", "10"),
				),
			},
			// Update and Read testing
			{
				Config: testAccMappingRuleResourceBasic(updatedRuleName, updatedDescription, 20),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMappingRuleExists("keep_mapping_rule.test"),
					resource.TestCheckResourceAttr("keep_mapping_rule.test", "name", updatedRuleName),
					resource.TestCheckResourceAttr("keep_mapping_rule.test", "description", updatedDescription),
					resource.TestCheckResourceAttr("keep_mapping_rule.test", "priority", "20"),
				),
			},
			// Test CSV data
			{
				Config: testAccMappingRuleResourceWithCSV(updatedRuleName, updatedDescription, 20),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMappingRuleExists("keep_mapping_rule.test"),
					resource.TestCheckResourceAttr("keep_mapping_rule.test", "name", updatedRuleName),
					testAccCheckCSVDataEqual("keep_mapping_rule.test", "csv_data", "env,team,owner\nproduction,sre,alice\nstaging,dev,bob"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// testAccMappingRuleResourceBasic returns the configuration for a basic mapping rule
func testAccMappingRuleResourceBasic(name, description string, priority int) string {
	return fmt.Sprintf(`
provider "keep" {
  api_key = "%s"
  api_url = "%s"
}

resource "keep_mapping_rule" "test" {
  name        = %q
  description = %q
  priority    = %d
  
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
`, os.Getenv("KEEP_API_KEY"), os.Getenv("KEEP_API_URL"), name, description, priority)
}

// testAccMappingRuleResourceWithCSV returns the configuration for a mapping rule with CSV data
func testAccMappingRuleResourceWithCSV(name, description string, priority int) string {
	return fmt.Sprintf(`
provider "keep" {
  api_key = "%s"
  api_url = "%s"
}

resource "keep_mapping_rule" "test" {
  name        = %q
  description = %q
  priority    = %d
  
  matchers = {
    env  = "production"
    team = "sre"
  }

  csv_data = <<-EOT
env,team,owner
production,sre,alice
staging,dev,bob
  EOT
}
`, os.Getenv("KEEP_API_KEY"), os.Getenv("KEEP_API_URL"), name, description, priority)
}
