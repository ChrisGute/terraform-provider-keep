// resource_provider_test.go - Acceptance tests for the provider resource
package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccProviderResource(t *testing.T) {
	// Skip if running short tests
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Skip if acceptance testing is not enabled
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test as TF_ACC is not set")
	}

	resourceName := "keep_provider.test"
	providerName := "test-squadcast-provider"
	providerType := "squadcast"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProviderResourceConfig(providerName, providerType, "test-key", "test-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProviderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", providerName),
					resource.TestCheckResourceAttr(resourceName, "type", providerType),
					resource.TestCheckResourceAttr(resourceName, "config.test-key", "test-value"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "installed"),
				),
			},
			// ImportState testing
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config"}, // Config is sensitive and not returned in the same format
			},
			// Update and Read testing
			{
				Config: testAccProviderResourceConfig(providerName, providerType, "service_region", "EU"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProviderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", providerName),
					resource.TestCheckResourceAttr(resourceName, "config.updated-key", "updated-value"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccCheckProviderExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		return nil
	}
}

func testAccProviderResourceConfig(name, providerType, configKey, configValue string) string {
	return fmt.Sprintf(`
provider "keep" {
  api_key = "%s"
  api_url = "%s"
}

resource "keep_provider" "test" {
  name = %q
  type = %q
  config = {
    service_region = "US"
    %s = %q
  }
}
`,
		os.Getenv("KEEP_API_KEY"),
		os.Getenv("KEEP_API_URL"),
		name,
		providerType,
		configKey,
		configValue,
	)
}
