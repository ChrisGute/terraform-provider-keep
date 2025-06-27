package acceptance

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/keephq/terraform-provider-keep/test/acceptance/verification"
)

func TestAccExtractionRule_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	t.Parallel()

	// Get test configuration from environment
	apiKey := os.Getenv("KEEP_API_KEY")
	if apiKey == "" {
		t.Fatal("KEEP_API_KEY must be set for acceptance tests")
	}

	apiURL := os.Getenv("KEEP_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	// Create API client for verification
	apiClient := verification.NewAPIClient(apiURL, apiKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckExtractionRuleDestroy(apiClient),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccExtractionRuleConfig(apiKey, apiURL, "test-rule", "Test extraction rule created by acceptance test", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the extraction rule exists in Terraform state
					resource.TestCheckResourceAttr("keep_extraction_rule.test", "name", "test-rule"),
					resource.TestCheckResourceAttr("keep_extraction_rule.test", "description", "Test extraction rule created by acceptance test"),
					resource.TestCheckResourceAttr("keep_extraction_rule.test", "attribute", "message"),
					resource.TestCheckResourceAttr("keep_extraction_rule.test", "regex", "test-([A-Z0-9]+)"),
					resource.TestCheckResourceAttr("keep_extraction_rule.test", "priority", "1"),
					resource.TestCheckResourceAttr("keep_extraction_rule.test", "disabled", "false"),
					resource.TestCheckResourceAttr("keep_extraction_rule.test", "pre", "false"),

					// Verify the extraction rule exists in the API
					func(s *terraform.State) error {
						return verifyExtractionRuleInAPI(s, apiClient, "test-rule", "Test extraction rule created by acceptance test")
					},
				),
			},
			// ImportState testing
			{
				ResourceName:      "keep_extraction_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccExtractionRuleConfig(apiKey, apiURL, "test-rule-updated", "Updated description", 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("keep_extraction_rule.test", "name", "test-rule-updated"),
					resource.TestCheckResourceAttr("keep_extraction_rule.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("keep_extraction_rule.test", "priority", "2"),

					// Verify the extraction rule was updated in the API
					func(s *terraform.State) error {
						return verifyExtractionRuleInAPI(s, apiClient, "test-rule-updated", "Updated description")
					},
				),
			},
		},
	})
}

func testAccExtractionRuleConfig(apiKey, apiURL, name, description string, priority int) string {
	return fmt.Sprintf(`
provider "keep" {
  api_key = %q
  api_url = %q
}

resource "keep_extraction_rule" "test" {
  name        = %q
  description = %q
  attribute   = "message"
  regex       = "test-([A-Z0-9]+)"
  priority    = %d
  disabled    = false
  pre         = false
}
`, apiKey, apiURL, name, description, priority)
}

func testAccCheckExtractionRuleDestroy(apiClient *verification.APIClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keep_extraction_rule" {
				continue
			}

			exists, err := apiClient.VerifyExtractionRuleExists(ctx, rs.Primary.ID)
			if err != nil {
				return fmt.Errorf("error checking if extraction rule exists: %w", err)
			}

			if exists {
				return fmt.Errorf("extraction rule %s still exists in API", rs.Primary.ID)
			}
		}

		return nil
	}
}

func verifyExtractionRuleInAPI(s *terraform.State, apiClient *verification.APIClient, expectedName, expectedDescription string) error {
	rs, ok := s.RootModule().Resources["keep_extraction_rule.test"]
	if !ok {
		return fmt.Errorf("not found: %s", "keep_extraction_rule.test")
	}

	id := rs.Primary.ID
	if id == "" {
		return fmt.Errorf("no ID is set")
	}

	tflog.Info(context.Background(), "Checking if extraction rule exists in API", map[string]interface{}{"id": id})

	exists, err := apiClient.VerifyExtractionRuleExists(context.Background(), id)
	if err != nil {
		return fmt.Errorf("error checking if extraction rule exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("extraction rule %s does not exist in API", id)
	}

	// Verify attributes in API
	if match, err := apiClient.VerifyExtractionRuleAttribute(
		context.Background(),
		id,
		"name",
		expectedName,
	); err != nil {
		return fmt.Errorf("error verifying name: %w", err)
	} else if !match {
		return fmt.Errorf("name does not match in API")
	}

	if match, err := apiClient.VerifyExtractionRuleAttribute(
		context.Background(),
		id,
		"description",
		expectedDescription,
	); err != nil {
		return fmt.Errorf("error verifying description: %w", err)
	} else if !match {
		return fmt.Errorf("description does not match in API")
	}

	return nil
}
