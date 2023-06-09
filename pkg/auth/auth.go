package auth

import (
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
)

// GetVerifiedOperationHash returns the operation hash from the request if the
// request is authenticated. If the request is not authenticated, it returns
// an error.
//
// Once the operation hash has been verified and extracted from the HTTP headers
// it is then can be used to verify the request body.
func GetVerifiedOperationHash(req *http.Request, keys []Key, clock clock.Clock) (OperationHash, error) {
	headers := &Headers{
		Authorization: req.Header.Get("Authorization"),
		Date:          req.Header.Get("Date"),
	}

	if headers.Authorization == "" && headers.Date == "" {
		// No auth header provided, so we can't authenticate the request.
		return "", ErrNoAuthorizationHeader
	}

	keyID, appSlug, envName, timestamp, opHash, err := headers.SigningComponents()
	if err != nil {
		return "", err
	}

	// First the timestamp, and don't do any work if it's too old or too new
	const allowedClockSkew = 2 * time.Minute
	if diff := clock.Since(timestamp); diff > allowedClockSkew || diff < -allowedClockSkew {
		return "", ErrAuthenticationExpired
	}

	// Find the key
	var key Key
	for _, k := range keys {
		if k.KeyID == keyID {
			key = k
			break
		}
	}
	if key.KeyID == 0 {
		return "", ErrAuthenticationFailed
	}

	// Rebuild the signature
	expectedHeaders := SignForVerification(&key, appSlug, envName, timestamp, opHash)

	// Verify the signature
	if !expectedHeaders.Equal(headers) {
		return "", ErrAuthenticationFailed
	}

	// Return the operation hash
	return opHash, nil
}
