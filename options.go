package platform

import (
	"github.com/benbjohnson/clock"
	"go.encore.dev/platform-sdk/internal/client"
	"go.encore.dev/platform-sdk/pkg/auth"
)

// Option is a function that can be passed to New to configure the SDK.
type Option func(config *client.Config)

// WithHost configures the SDK to use the specified host, overriding the default.
func WithHost(host string) Option {
	return func(config *client.Config) {
		config.Host = host
	}
}

// WithAppDetails configures the SDK to act on behalf of the specified application
// and environment.
func WithAppDetails(appSlug, envName string) Option {
	return func(config *client.Config) {
		config.AppSlug = appSlug
		config.EnvName = envName
	}
}

// WithAuthKeys configures the SDK to use the specified auth keys.
func WithAuthKeys(keys ...auth.Key) Option {
	return func(config *client.Config) {
		var latestKey auth.Key
		for _, key := range keys {
			if key.KeyID > latestKey.KeyID {
				latestKey = key
			}
		}

		config.AuthKeys = keys
		config.LatestAuthKey = latestKey
	}
}

// WithClock configures the SDK to use the specified clock.
//
// This is useful for testing with a mocked clock, if not
// specified a real clock will be used.
func WithClock(clock clock.Clock) Option {
	return func(config *client.Config) {
		config.Clock = clock
	}
}
