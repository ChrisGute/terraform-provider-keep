package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// VerifyProviderExists checks if a provider exists in the API
func (c *APIClient) VerifyProviderExists(ctx context.Context, id string) (bool, error) {
	url := fmt.Sprintf("%s/providers", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-API-KEY", c.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var providers []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&providers); err != nil {
		return false, fmt.Errorf("error decoding response: %w", err)
	}

	for _, provider := range providers {
		if providerID, ok := provider["id"]; ok && fmt.Sprint(providerID) == id {
			tflog.Info(ctx, "Verified provider exists in API", map[string]interface{}{
				"id":       id,
				"provider": provider,
			})
			return true, nil
		}
	}

	return false, nil
}

// VerifyProviderAttribute verifies an attribute of a provider
func (c *APIClient) VerifyProviderAttribute(
	ctx context.Context,
	id string,
	attribute string,
	expectedValue interface{},
) (bool, error) {
	url := fmt.Sprintf("%s/providers", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-API-KEY", c.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var providers []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&providers); err != nil {
		return false, fmt.Errorf("error decoding response: %w", err)
	}

	for _, provider := range providers {
		if providerID, ok := provider["id"]; ok && fmt.Sprint(providerID) == id {
			// Handle nested attributes (e.g., config.webhook_url)
			value := provider
			parts := strings.Split(attribute, ".")
			for i, part := range parts {
				if m, ok := value[part].(map[string]interface{}); ok && i < len(parts)-1 {
					value = m
				} else if i == len(parts)-1 {
					actualValue, exists := value[part]
					if !exists {
						return false, fmt.Errorf("attribute %s does not exist", attribute)
					}

					tflog.Info(ctx, "Verifying provider attribute", map[string]interface{}{
						"id":       id,
						"attribute": attribute,
						"expected":  expectedValue,
						"actual":    actualValue,
					})


					return fmt.Sprint(actualValue) == fmt.Sprint(expectedValue), nil
				}
			}
			return false, fmt.Errorf("attribute %s not found in provider", attribute)
		}
	}

	return false, fmt.Errorf("provider with id %s not found", id)
}

// VerifyProviderDoesNotExist verifies that a provider does not exist
func (c *APIClient) VerifyProviderDoesNotExist(ctx context.Context, id string) (bool, error) {
	exists, err := c.VerifyProviderExists(ctx, id)
	if err != nil {
		return false, err
	}
	return !exists, nil
}
