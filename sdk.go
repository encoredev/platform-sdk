package platform

import (
	"github.com/benbjohnson/clock"
	"go.encore.dev/platform-sdk/encorecloud"
	"go.encore.dev/platform-sdk/internal/client"
)

// NewSDK creates a new SDK with the specified options.
func NewSDK(options ...Option) *SDK {
	// Create the raw client
	cfg := &client.Config{
		Clock: clock.New(),
	}
	for _, option := range options {
		option(cfg)
	}
	rawClient := client.New(cfg)

	// Now create the SDK struct
	return &SDK{
		EncoreCloud: encorecloud.NewClient(rawClient),
	}
}

// SDK is the main SDK for communicating with the Encore Platform services.
type SDK struct {
	// EncoreCloud is the client for services hosted specifically
	// to support applications deployed within the Encore Cloud.
	EncoreCloud *encorecloud.Client
}
