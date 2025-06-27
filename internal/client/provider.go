// provider.go - Provider-related API client methods
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"path"
)

// CreateProvider creates a new provider in KeepHQ by installing it
func (c *Client) CreateProvider(ctx context.Context, req CreateProviderRequest) (*Provider, error) {
	// The provider installation requires a specific format
	installReq := map[string]interface{}{
		"provider_id":   req.Type, // Use type as provider_id
		"provider_name": req.Name,
		"provider_type": req.Type,
		"pulling_enabled": true, // Default to true unless specified
	}

	// Add the config as top-level fields in the request
	// Since req.Config is map[string]string, we can directly assign values
	for k, v := range req.Config {
		installReq[k] = v
	}

	// Log the request payload for debugging
	log.Printf("Creating provider with payload: %+v", installReq)

	// Use the install endpoint for provider creation
	// Pass the raw map to Post, which will handle the JSON marshaling
	resp, err := c.Post(ctx, "/providers/install", installReq)
	if err != nil {
		log.Printf("Error response from API: %s", string(resp))
		return nil, fmt.Errorf("error installing provider: %w", err)
	}

	// Log the response for debugging
	log.Printf("Provider creation response: %s", string(resp))

	// The response might be just the provider object, not wrapped in a ProviderResponse
	var provider Provider
	if err := json.Unmarshal(resp, &provider); err != nil {
		log.Printf("Error parsing provider response: %v, response: %s", err, string(resp))
		return nil, fmt.Errorf("error parsing provider response: %w", err)
	}

	log.Printf("Successfully created provider: %+v", provider)
	return &provider, nil
}

// GetProvider retrieves a provider by ID
func (c *Client) GetProvider(ctx context.Context, id string) (*Provider, error) {
	urlPath := path.Join("/providers", url.PathEscape(id))
	resp, err := c.Get(ctx, urlPath)
	if err != nil {
		return nil, fmt.Errorf("error getting provider: %w", err)
	}

	var providerResp ProviderResponse
	if err := json.Unmarshal(resp, &providerResp); err != nil {
		return nil, fmt.Errorf("error parsing provider response: %w", err)
	}

	return &providerResp.Provider, nil
}

// UpdateProvider updates an existing provider
func (c *Client) UpdateProvider(ctx context.Context, id string, req UpdateProviderRequest) (*Provider, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	urlPath := path.Join("/providers", url.PathEscape(id))
	resp, err := c.Put(ctx, urlPath, body)
	if err != nil {
		return nil, fmt.Errorf("error updating provider: %w", err)
	}

	var providerResp ProviderResponse
	if err := json.Unmarshal(resp, &providerResp); err != nil {
		return nil, fmt.Errorf("error parsing provider response: %w", err)
	}

	return &providerResp.Provider, nil
}

// DeleteProvider deletes a provider by ID
func (c *Client) DeleteProvider(ctx context.Context, id string) error {
	urlPath := path.Join("/providers", url.PathEscape(id))
	_, err := c.Delete(ctx, urlPath)
	if err != nil {
		return fmt.Errorf("error deleting provider: %w", err)
	}

	return nil
}

// ListProviders retrieves all providers
func (c *Client) ListProviders(ctx context.Context) ([]Provider, error) {
	resp, err := c.Get(ctx, "/providers")
	if err != nil {
		return nil, fmt.Errorf("error listing providers: %w", err)
	}

	var listResp ListProvidersResponse
	if err := json.Unmarshal(resp, &listResp); err != nil {
		return nil, fmt.Errorf("error parsing providers list: %w", err)
	}

	return listResp.Providers, nil
}
