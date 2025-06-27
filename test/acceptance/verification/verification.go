package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type APIClient struct {
	BaseURL string
	APIKey  string
}

// NewAPIClient creates a new API client for verification
func NewAPIClient(baseURL, apiKey string) *APIClient {
	return &APIClient{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		APIKey:  apiKey,
	}
}

// VerifyExtractionRuleExists checks if an extraction rule exists in the API
func (c *APIClient) VerifyExtractionRuleExists(ctx context.Context, id string) (bool, error) {
	url := fmt.Sprintf("%s/extraction", c.BaseURL)
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

	var rules []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rules); err != nil {
		return false, fmt.Errorf("error decoding response: %w", err)
	}

	for _, rule := range rules {
		if ruleID, ok := rule["id"]; ok && fmt.Sprint(ruleID) == id {
			tflog.Info(ctx, "Verified extraction rule exists in API", map[string]interface{}{
				"id":   id,
				"rule": rule,
			})
			return true, nil
		}
	}

	return false, nil
}

// VerifyExtractionRuleDoesNotExist verifies that an extraction rule does not exist
func (c *APIClient) VerifyExtractionRuleDoesNotExist(ctx context.Context, id string) (bool, error) {
	exists, err := c.VerifyExtractionRuleExists(ctx, id)
	if err != nil {
		return false, err
	}
	return !exists, nil
}

// VerifyExtractionRuleAttribute verifies an attribute of an extraction rule
func (c *APIClient) VerifyExtractionRuleAttribute(
	ctx context.Context,
	id string,
	attribute string,
	expectedValue interface{},
) (bool, error) {
	url := fmt.Sprintf("%s/extraction", c.BaseURL)
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

	var rules []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rules); err != nil {
		return false, fmt.Errorf("error decoding response: %w", err)
	}

	for _, rule := range rules {
		if ruleID, ok := rule["id"]; ok && fmt.Sprint(ruleID) == id {
			actualValue, exists := rule[attribute]
			if !exists {
				return false, fmt.Errorf("attribute %s does not exist", attribute)
			}

			tflog.Info(ctx, "Verifying extraction rule attribute", map[string]interface{}{
				"id":       id,
				"attribute": attribute,
				"expected":  expectedValue,
				"actual":    actualValue,
			})


			return fmt.Sprint(actualValue) == fmt.Sprint(expectedValue), nil
		}
	}

	return false, fmt.Errorf("extraction rule with id %s not found", id)
}
