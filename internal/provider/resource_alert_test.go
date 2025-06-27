package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAlertResource(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	resourceName := "keep_alert.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAlertResourceConfig("test-alert", "firing", "high"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlertResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-alert"),
					resource.TestCheckResourceAttr(resourceName, "status", "firing"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "fingerprint"),
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
				Config: testAccAlertResourceConfig("test-alert-updated", "resolved", "low"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-alert-updated"),
					resource.TestCheckResourceAttr(resourceName, "status", "resolved"),
					resource.TestCheckResourceAttr(resourceName, "severity", "low"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccAlertResourceConfig(name, status, severity string) string {
	apiURL := os.Getenv("KEEP_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	apiKey := os.Getenv("KEEP_API_KEY")
	if apiKey == "" {
		panic("KEEP_API_KEY environment variable must be set for acceptance tests")
	}

	return fmt.Sprintf(`
terraform {
  required_providers {
    keep = {
      source = "registry.terraform.io/hashicorp/keep"
    }
  }
}

provider "keep" {
  api_key = %q
  api_url = %q
}

resource "keep_alert" "test" {
  name     = %q
  status   = %q
  severity = %q
  
  environment = "test"
  service     = "test-service"
  source      = ["test-source"]
  message     = "Test alert message"
  description = "Test alert description"
  
  labels = {
    "test" = "true"
  }

  # Add lifecycle block to prevent destroy for debugging
  lifecycle {
    prevent_destroy = false
  }
}

output "api_key_set" {
  value = var.KEEP_API_KEY != ""
}
`, 
	apiKey,
	apiURL,
	name, 
	status, 
	severity)
}

func testAccCheckAlertResourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		// Check if the resource has an ID
		if rs.Primary.ID == "" {
			return fmt.Errorf("No alert ID is set")
		}

		return nil
	}
}

func TestAccAlertResource_validation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping validation test in short mode")
	}

	// Skip if acceptance testing is not enabled
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test as TF_ACC is not set")
	}
	tests := []struct {
		name        string
		config      string
		expectError *regexp.Regexp
	}{
		{
			name: "missing required name",
			config: `
resource "keep_alert" "test" {
  status   = "firing"
  severity = "high"
}
`,
			expectError: regexp.MustCompile("The argument \"name\" is required"),
		},
		{
			name: "invalid status",
			config: `
resource "keep_alert" "test" {
  name     = "test"
  status   = "invalid"
  severity = "high"
}
`,
			expectError: regexp.MustCompile("expected status to be one of"),
		},
		{
			name: "invalid severity",
			config: `
resource "keep_alert" "test" {
  name     = "test"
  status   = "firing"
  severity = "invalid"
}
`,
			expectError: regexp.MustCompile("expected severity to be one of"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      tt.config,
						ExpectError: tt.expectError,
					},
				},
			})
		})
	}
}
