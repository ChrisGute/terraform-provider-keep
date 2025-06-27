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

func TestAccProvider_basic(t *testing.T) {
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
		CheckDestroy:             testAccCheckProviderDestroy(apiClient),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProviderConfig(apiKey, apiURL, "test-provider", "slack", "#alerts"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the provider exists in Terraform state
					resource.TestCheckResourceAttr("keep_provider.test", "name", "test-provider"),
					resource.TestCheckResourceAttr("keep_provider.test", "type", "slack"),
					resource.TestCheckResourceAttr("keep_provider.test", "config.channel", "#alerts"),

					// Verify the provider exists in the API
					func(s *terraform.State) error {
						return verifyProviderInAPI(s, apiClient, "test-provider", "slack", "#alerts")
					},
				),
			},
			// ImportState testing
			{
				ResourceName:      "keep_provider.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccProviderConfig(apiKey, apiURL, "test-provider-updated", "slack", "#general"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("keep_provider.test", "name", "test-provider-updated"),
					resource.TestCheckResourceAttr("keep_provider.test", "config.channel", "#general"),

					// Verify the provider was updated in the API
					func(s *terraform.State) error {
						return verifyProviderInAPI(s, apiClient, "test-provider-updated", "slack", "#general")
					},
				),
			},
		},
	})
}

func testAccProviderConfig(apiKey, apiURL, name, providerType, channel string) string {
	return fmt.Sprintf(`
provider "keep" {
  api_key = %q
  api_url = %q
}

resource "keep_provider" "test" {
  name = %q
  type = %q
  config = {
    webhook_url = "https://hooks.slack.com/services/xxx/yyy/zzz"
    channel     = %q
  }
}
`, apiKey, apiURL, name, providerType, channel)
}

func testAccCheckProviderDestroy(apiClient *verification.APIClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keep_provider" {
				continue
			}

			exists, err := apiClient.VerifyProviderExists(ctx, rs.Primary.ID)
			if err != nil {
				return fmt.Errorf("error checking if provider exists: %w", err)
			}

			if exists {
				return fmt.Errorf("provider %s still exists in API", rs.Primary.ID)
			}
		}

		return nil
	}
}

func verifyProviderInAPI(s *terraform.State, apiClient *verification.APIClient, expectedName, expectedType, expectedChannel string) error {
	rs, ok := s.RootModule().Resources["keep_provider.test"]
	if !ok {
		return fmt.Errorf("not found: %s", "keep_provider.test")
	}

	id := rs.Primary.ID
	if id == "" {
		return fmt.Errorf("no ID is set")
	}

	tflog.Info(context.Background(), "Checking if provider exists in API", map[string]interface{}{"id": id})

	exists, err := apiClient.VerifyProviderExists(context.Background(), id)
	if err != nil {
		return fmt.Errorf("error checking if provider exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("provider %s does not exist in API", id)
	}

	// Verify attributes in API
	if match, err := apiClient.VerifyProviderAttribute(
		context.Background(),
		id,
		"name",
		expectedName,
	); err != nil {
		return fmt.Errorf("error verifying name: %w", err)
	} else if !match {
		return fmt.Errorf("name does not match in API")
	}

	if match, err := apiClient.VerifyProviderAttribute(
		context.Background(),
		id,
		"type",
		expectedType,
	); err != nil {
		return fmt.Errorf("error verifying type: %w", err)
	} else if !match {
		return fmt.Errorf("type does not match in API")
	}

	if expectedChannel != "" {
		if match, err := apiClient.VerifyProviderAttribute(
			context.Background(),
			id,
			"config.channel",
			expectedChannel,
		); err != nil {
			return fmt.Errorf("error verifying channel: %w", err)
		} else if !match {
			return fmt.Errorf("channel does not match in API")
		}
	}

	return nil
}
