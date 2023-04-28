package encorecloud

import (
	"go.encore.dev/platform-sdk/internal/client"
)

// Client is the SDK for communicating with the Encore Cloud specific services.
type Client struct {
	client *client.Client
}

func NewClient(client *client.Client) *Client {
	return &Client{client}
}
