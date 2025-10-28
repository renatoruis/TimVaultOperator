package vault

import (
	"context"
	"fmt"

	vault "github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client
type Client struct {
	client *vault.Client
}

// NewClient creates a new Vault client
func NewClient(address, token string) (*Client, error) {
	config := vault.DefaultConfig()
	config.Address = address

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	client.SetToken(token)

	return &Client{client: client}, nil
}

// GetSecrets retrieves secrets from the specified path in Vault
func (c *Client) GetSecrets(ctx context.Context, path string) (map[string]string, error) {
	secret, err := c.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret from vault: %w", err)
	}

	if secret == nil {
		return nil, fmt.Errorf("secret not found at path: %s", path)
	}

	// Handle both KV v1 and KV v2
	var data map[string]interface{}
	if secret.Data["data"] != nil {
		// KV v2
		data = secret.Data["data"].(map[string]interface{})
	} else {
		// KV v1
		data = secret.Data
	}

	// Convert to map[string]string
	result := make(map[string]string)
	for k, v := range data {
		if str, ok := v.(string); ok {
			result[k] = str
		} else {
			result[k] = fmt.Sprintf("%v", v)
		}
	}

	return result, nil
}
