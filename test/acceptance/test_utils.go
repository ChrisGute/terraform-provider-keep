package acceptance

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/keephq/terraform-provider-keep/internal/provider"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"keep": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// testAccPreCheck validates the necessary test API keys exist
// in the testing environment
func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the environment to help test errors shine up.
	if v := os.Getenv("KEEP_API_KEY"); v == "" {
		t.Fatal("KEEP_API_KEY must be set for acceptance tests")
	}
}

// testAccPreCheck verifies the test environment is properly configured
func testAccProviderConfigure() {
	// Configure the provider with the test environment settings
	os.Setenv("KEEP_API_KEY", os.Getenv("KEEP_API_KEY"))
	if apiURL := os.Getenv("KEEP_API_URL"); apiURL != "" {
		os.Setenv("KEEP_API_URL", apiURL)
	}
}
