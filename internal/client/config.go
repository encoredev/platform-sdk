package client

import (
	"github.com/benbjohnson/clock"
	"go.encore.dev/platform-sdk/pkg/auth"
)

// Config is the configuration for the client.
type Config struct {
	Host          string      // The host to use
	Clock         clock.Clock // The clock to use
	AppSlug       string      // The app slug to use
	EnvName       string      // The environment name to use
	LatestAuthKey auth.Key    // The auth key to use when signing new requests
	AuthKeys      []auth.Key  // All known auth keys (used to verify data sent from the Encore Platform services)
}
