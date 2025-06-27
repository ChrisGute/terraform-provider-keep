// client.go - API client for KeepHQ
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	// DefaultTimeout is the default timeout for API requests
	DefaultTimeout = 30 * time.Second
	// DefaultBaseURL is the default base URL for the KeepHQ API
	DefaultBaseURL = "http://localhost:8080"
)

// Client is the KeepHQ API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
}

// NewClient creates a new KeepHQ API client
func NewClient(baseURL, apiKey string) (*Client, error) {
	// Set default base URL if not provided
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json"

	// Add API key to headers if provided
	if apiKey != "" {
		headers["X-API-KEY"] = apiKey
	}

	// Create a context with debug logging
	ctx := context.Background()
	ctx = tflog.SetField(ctx, "provider", "keep")
	
	tflog.Debug(ctx, "Creating new KeepHQ API client", 
		map[string]interface{}{
			"base_url":    baseURL,
			"api_key_set": apiKey != "",
			"headers":     headers,
	})

	return &Client{
		baseURL: baseURL,
		headers: headers,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}, nil
}

// doRequest performs an HTTP request with the given method, path, and body
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	// Add debug logging for the client configuration
	tflog.Debug(ctx, "Client configuration", map[string]interface{}{
		"baseURL": c.baseURL,
		"headers": c.headers,
	})

	var reqBody io.Reader = http.NoBody

	// Marshal the request body if provided
	var jsonBody []byte
	var err error
	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	// Create the request
	url := c.baseURL + path
	tflog.Debug(ctx, "Creating request", map[string]interface{}{
		"method": method,
		"url":    url,
	})

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add headers
	headers := make(map[string]string)
	for k, v := range c.headers {
		req.Header.Add(k, v)
		headers[k] = v
	}

	// Log the request (without sensitive data)
	logHeaders := make(map[string]string)
	for k, v := range headers {
		if k == "X-API-KEY" {
			logHeaders[k] = "[REDACTED]"
		} else {
			logHeaders[k] = v
		}
	}

	tflog.Debug(ctx, "Sending request with headers", map[string]interface{}{
		"method":  method,
		"url":     req.URL.String(),
		"path":    path,
		"headers": logHeaders,
	})

	// Log the actual request being sent
	headerList := make([]string, 0, len(req.Header))
	for k, v := range req.Header {
		headerList = append(headerList, fmt.Sprintf("%s: %v", k, v))
	}
	tflog.Debug(ctx, "Actual request headers", map[string]interface{}{
		"headers": headerList,
	})

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	tflog.Debug(ctx, "Received response", map[string]interface{}{
		"status":     resp.Status,
		"statusCode": resp.StatusCode,
	})

	// Check for error responses
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPost, path, body)
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPut, path, body)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodDelete, path, nil)
}

// ExtractionRule represents an extraction rule in Keep
// This mirrors the API response structure
// https://github.com/keephq/keep/blob/main/keep/api/models/db/extraction.py
// https://github.com/keephq/keep/blob/main/keep/api/routes/extraction.py
type ExtractionRule struct {
	ID          int        `json:"id,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Priority    int        `json:"priority"`
	Disabled    bool       `json:"disabled"`
	Pre         bool       `json:"pre"`
	Condition   string     `json:"condition,omitempty"`
	Attribute   string     `json:"attribute"`
	Regex       string     `json:"regex"`
	CreatedAt   string     `json:"created_at,omitempty"`
	UpdatedAt   *string    `json:"updated_at,omitempty"`
	CreatedBy   string     `json:"created_by,omitempty"`
	UpdatedBy   *string    `json:"updated_by,omitempty"`
}

// CreateExtractionRule creates a new extraction rule
func (c *Client) CreateExtractionRule(ctx context.Context, rule map[string]interface{}) (map[string]interface{}, error) {
	path := "/extraction"
	body, err := c.Post(ctx, path, rule)
	if err != nil {
		return nil, fmt.Errorf("error creating extraction rule: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing extraction rule response: %w", err)
	}

	return result, nil
}

// GetExtractionRule retrieves an extraction rule by ID
func (c *Client) GetExtractionRule(ctx context.Context, id string) (map[string]interface{}, error) {
	// First, list all extraction rules
	rules, err := c.ListExtractionRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing extraction rules: %w", err)
	}

	// Find the rule with the matching ID
	for _, rule := range rules {
		ruleID, ok := rule["id"].(float64)
		if !ok {
			continue
		}
		
		if fmt.Sprintf("%.0f", ruleID) == id {
			return rule, nil
		}
	}

	return nil, fmt.Errorf("extraction rule with ID %s not found", id)
}

// UpdateExtractionRule updates an existing extraction rule
func (c *Client) UpdateExtractionRule(ctx context.Context, id string, rule map[string]interface{}) (map[string]interface{}, error) {
	path := fmt.Sprintf("/extraction/%s", id)
	body, err := c.Put(ctx, path, rule)
	if err != nil {
		return nil, fmt.Errorf("error updating extraction rule: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing updated extraction rule: %w", err)
	}

	return result, nil
}

// DeleteExtractionRule deletes an extraction rule by ID
func (c *Client) DeleteExtractionRule(ctx context.Context, id string) error {
	path := fmt.Sprintf("/extraction/%s", id)
	_, err := c.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("error deleting extraction rule: %w", err)
	}
	return nil
}

// ListExtractionRules retrieves all extraction rules
func (c *Client) ListExtractionRules(ctx context.Context) ([]map[string]interface{}, error) {
	path := "/extraction"
	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("error listing extraction rules: %w", err)
	}

	var rules []map[string]interface{}
	if err := json.Unmarshal(body, &rules); err != nil {
		return nil, fmt.Errorf("error parsing extraction rules list: %w", err)
	}

	return rules, nil
}

// Alert represents an alert in KeepHQ
type Alert struct {
	ID          string                 `json:"id,omitempty"`
	Fingerprint string                 `json:"fingerprint,omitempty"`
	Name        string                 `json:"name"`
	Status      string                 `json:"status,omitempty"`
	Severity    string                 `json:"severity,omitempty"`
	Environment string                 `json:"environment,omitempty"`
	Service     string                 `json:"service,omitempty"`
	Source      []string               `json:"source,omitempty"`
	Message     string                 `json:"message,omitempty"`
	Description string                 `json:"description,omitempty"`
	URL         string                 `json:"url,omitempty"`
	ImageURL    string                 `json:"image_url,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty"`
	LastReceived string                `json:"lastReceived,omitempty"`
}

// CreateAlert creates a new alert
func (c *Client) CreateAlert(ctx context.Context, alert Alert) (map[string]interface{}, error) {
	path := "/alerts/event"
	body, err := c.Post(ctx, path, alert)
	if err != nil {
		return nil, fmt.Errorf("error creating alert: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing alert response: %w", err)
	}

	return result, nil
}

// GetAlert retrieves an alert by fingerprint
func (c *Client) GetAlert(ctx context.Context, fingerprint string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/alerts/%s", fingerprint)
	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("error getting alert: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing alert: %w", err)
	}

	return result, nil
}

// SearchAlerts searches for alerts
func (c *Client) SearchAlerts(ctx context.Context, query map[string]interface{}) ([]map[string]interface{}, error) {
	path := "/alerts/search"
	body, err := c.Post(ctx, path, query)
	if err != nil {
		return nil, fmt.Errorf("error searching alerts: %w", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing search results: %w", err)
	}

	return result, nil
}

// EnrichAlert enriches an alert with additional data
func (c *Client) EnrichAlert(ctx context.Context, fingerprint string, enrichData map[string]interface{}) (map[string]interface{}, error) {
	path := "/alerts/enrich"
	body, err := c.Post(ctx, path, enrichData)
	if err != nil {
		return nil, fmt.Errorf("error enriching alert: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing enrich response: %w", err)
	}

	return result, nil
}

// DeleteAlert deletes an alert by fingerprint
func (c *Client) DeleteAlert(ctx context.Context, fingerprint string) error {
	_, err := c.Delete(ctx, fmt.Sprintf("/alerts/%s", fingerprint))
	if err != nil {
		return fmt.Errorf("failed to delete alert: %w", err)
	}

	return nil
}

// MappingRule represents a mapping rule in Keep
// This mirrors the API response structure
// https://github.com/keephq/keep/blob/main/keep/api/models/db/mapping.py
// https://github.com/keephq/keep/blob/main/keep/api/routes/mapping.py
type MappingRule struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Priority    int                    `json:"priority"`
	Disabled    bool                   `json:"disabled"`
	Matchers    map[string]string      `json:"matchers"`
	CSVData     string                 `json:"csv_data,omitempty"`
	CreatedAt   string                 `json:"created_at,omitempty"`
	UpdatedAt   *string                `json:"updated_at,omitempty"`
	CreatedBy   string                 `json:"created_by,omitempty"`
	UpdatedBy   *string                `json:"updated_by,omitempty"`
}

// CreateMappingRule creates a new mapping rule
func (c *Client) CreateMappingRule(ctx context.Context, rule map[string]interface{}) (map[string]interface{}, error) {
	tflog.Debug(ctx, "Creating mapping rule", map[string]interface{}{
		"request": rule,
	})

	resp, err := c.Post(ctx, "/mapping", rule)
	if err != nil {
		tflog.Error(ctx, "Failed to create mapping rule", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create mapping rule: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		tflog.Error(ctx, "Failed to unmarshal response", map[string]interface{}{
			"error": err.Error(),
			"response": string(resp),
		})
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	tflog.Debug(ctx, "Created mapping rule", map[string]interface{}{
		"response": result,
	})

	return result, nil
}

// GetMappingRule retrieves a mapping rule by ID
func (c *Client) GetMappingRule(ctx context.Context, id string) (map[string]interface{}, error) {
	tflog.Debug(ctx, "Getting mapping rule", map[string]interface{}{
		"id": id,
	})

	// Try to get the rule directly by ID first
	resp, err := c.Get(ctx, "/mapping/"+id)
	if err == nil {
		var rule map[string]interface{}
		if err := json.Unmarshal(resp, &rule); err == nil {
			tflog.Debug(ctx, "Found mapping rule by direct ID lookup", map[string]interface{}{
				"id":   id,
				"rule": rule,
			})
			return rule, nil
		}
	}

	tflog.Debug(ctx, "Direct lookup failed, falling back to listing all rules", map[string]interface{}{
		"id":    id,
		"error": err.Error(),
	})

	// Fall back to listing all rules if direct lookup fails
	rules, err := c.ListMappingRules(ctx)
	if err != nil {
		tflog.Error(ctx, "Failed to list mapping rules", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to list mapping rules: %w", err)
	}

	tflog.Debug(ctx, "Searching through rules for matching ID", map[string]interface{}{
		"id":     id,
		"count":  len(rules),
		"rules":  rules,
	})

	// Find the rule with the matching ID
	for i, rule := range rules {
		ruleID, ok := rule["id"].(string)
		if !ok {
			tflog.Warn(ctx, "Mapping rule has invalid ID", map[string]interface{}{
				"rule": rule,
			})
			continue
		}

		tflog.Debug(ctx, fmt.Sprintf("Checking rule %d", i), map[string]interface{}{
			"rule_id": ruleID,
			"wanted_id": id,
		})

		if ruleID == id {
			tflog.Debug(ctx, "Found mapping rule in list", map[string]interface{}{
				"id":   id,
				"rule": rule,
			})
			return rule, nil
		}
	}

	tflog.Error(ctx, "Mapping rule not found", map[string]interface{}{
		"id":          id,
		"rules_count": len(rules),
	})

	return nil, fmt.Errorf("mapping rule with ID %s not found", id)
}

// UpdateMappingRule updates an existing mapping rule
func (c *Client) UpdateMappingRule(ctx context.Context, id string, rule map[string]interface{}) (map[string]interface{}, error) {
	// In Keep, we need to delete and recreate the rule to update it
	// First, delete the existing rule
	if err := c.DeleteMappingRule(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to delete existing mapping rule: %w", err)
	}

	// Then create a new rule with the updated data
	return c.CreateMappingRule(ctx, rule)
}

// DeleteMappingRule deletes a mapping rule by ID
func (c *Client) DeleteMappingRule(ctx context.Context, id string) error {
	_, err := c.Delete(ctx, fmt.Sprintf("/mapping/%s", id))
	if err != nil {
		return fmt.Errorf("failed to delete mapping rule: %w", err)
	}

	return nil
}

// ListMappingRules retrieves all mapping rules
func (c *Client) ListMappingRules(ctx context.Context) ([]map[string]interface{}, error) {
	tflog.Debug(ctx, "Listing all mapping rules")

	resp, err := c.Get(ctx, "/mapping")
	if err != nil {
		tflog.Error(ctx, "Failed to list mapping rules", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to list mapping rules: %w", err)
	}

	// Log the raw response for debugging
	tflog.Debug(ctx, "Raw mapping rules response", map[string]interface{}{
		"response": string(resp),
	})

	var rules []map[string]interface{}
	if err := json.Unmarshal(resp, &rules); err != nil {
		tflog.Error(ctx, "Failed to unmarshal mapping rules response", map[string]interface{}{
			"error": err.Error(),
			"response": string(resp),
		})
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Log the parsed rules for debugging
	for i, rule := range rules {
		tflog.Debug(ctx, fmt.Sprintf("Rule %d", i), map[string]interface{}{
			"rule": rule,
			"id":   rule["id"],
			"name": rule["name"],
		})
	}

	tflog.Debug(ctx, "Retrieved mapping rules", map[string]interface{}{
		"count": len(rules),
	})

	return rules, nil
}
