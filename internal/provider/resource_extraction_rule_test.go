// resource_extraction_rule_test.go - Acceptance tests for the extraction_rule resource
package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keephq/terraform-provider-keep/internal/client"
)

func TestAccExtractionRuleResource(t *testing.T) {
	// Skip if running short tests
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}

	resourceName := "keep_extraction_rule.test"
	ruleName := "test-extraction-rule"
	attribute := "test.attribute"
	regex := `(test-pattern-\d+)`
	updatedRegex := `(updated-pattern-\d+)`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccExtractionRuleResourceConfig(ruleName, attribute, regex, false, "test condition"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExtractionRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "attribute", attribute),
					resource.TestCheckResourceAttr(resourceName, "regex", regex),
					resource.TestCheckResourceAttr(resourceName, "disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "pre", "false"),
					resource.TestCheckResourceAttr(resourceName, "condition", "test condition"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccExtractionRuleResourceConfig(ruleName, attribute, updatedRegex, true, "updated condition"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExtractionRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "regex", updatedRegex),
					resource.TestCheckResourceAttr(resourceName, "disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "condition", "updated condition"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccCheckExtractionRuleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		// Get the ID from the resource
		id := rs.Primary.ID
		if id == "" {
			return fmt.Errorf("no ID is set")
		}

		// Verify the extraction rule exists by making an API call
		// This assumes you have a way to get a client in your test context
		client, err := getTestClient()
		if err != nil {
			return err
		}

		_, err = client.GetExtractionRule(context.Background(), id)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccExtractionRuleResourceConfig(name, attribute, regex string, disabled bool, condition string) string {
	disabledStr := "false"
	if disabled {
		disabledStr = "true"
	}

	apiURL := os.Getenv("KEEP_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	return fmt.Sprintf(`
provider "keep" {
  api_key = %q
  api_url = %q
}

resource "keep_extraction_rule" "test" {
  name        = %q
  description = "Test extraction rule"
  attribute   = %q
  regex       = %q
  disabled    = %s
  condition   = %q
  priority    = 10
  pre         = false
}`, 
	os.Getenv("KEEP_API_KEY"), 
	apiURL,
	name, 
	attribute, 
	regex, 
	disabledStr, 
	condition)
}

// getTestClient creates a test client using environment variables
func getTestClient() (*client.Client, error) {
	apiURL := os.Getenv("KEEP_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	apiKey := os.Getenv("KEEP_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("KEEP_API_KEY environment variable must be set for acceptance tests")
	}

	c, err := client.NewClient(apiURL, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	return c, nil
}
