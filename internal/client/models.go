// models.go - Data models for the KeepHQ API
package client

// Provider represents a KeepHQ provider
type Provider struct {
	ID               string            `json:"id,omitempty"`
	Name             string            `json:"name"`
	Type             string            `json:"type"`
	Config           map[string]string `json:"config,omitempty"`
	Installed        bool              `json:"installed,omitempty"`
	LastAlertReceived string            `json:"last_alert_received,omitempty"`
}

// CreateProviderRequest represents the request body for creating a provider
type CreateProviderRequest struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Config  map[string]string `json:"config"`
	Webhook *bool             `json:"webhook,omitempty"`
}

// UpdateProviderRequest represents the request body for updating a provider
type UpdateProviderRequest struct {
	Name   string            `json:"name,omitempty"`
	Config map[string]string `json:"config,omitempty"`
}

// ProviderResponse represents the API response for provider operations
type ProviderResponse struct {
	Provider Provider `json:"provider"`
}

// ListProvidersResponse represents the API response for listing providers
type ListProvidersResponse struct {
	Providers []Provider `json:"providers"`
}
