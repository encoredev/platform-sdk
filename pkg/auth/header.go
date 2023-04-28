package auth

import (
	"crypto/hmac"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Headers are the headers that are used to authenticate a request.
type Headers struct {
	Authorization string `header:"Authorization" encore:"sensitive"`
	Date          string `header:"Date"`
}

// Equal returns true if the headers are equal.
//
// It compares the Authorization and Date headers using
// hmac.Equal to prevent timing attacks.
func (h *Headers) Equal(other *Headers) bool {
	authMatches := hmac.Equal([]byte(h.Authorization), []byte(other.Authorization))
	dateMatches := hmac.Equal([]byte(h.Date), []byte(other.Date))
	return authMatches && dateMatches
}

// SigningComponents returns the components of the authorization header.
func (h *Headers) SigningComponents() (keyID uint32, appSlug, envName string, timestamp time.Time, operationHash OperationHash, err error) {
	const expectedComponentCount = 3
	switch {
	case h.Authorization == "":
		err = ErrNoAuthorizationHeader
		return
	case h.Date == "":
		err = ErrNoDateHeader
		return
	}

	// First parse the date header
	timestamp, err = http.ParseTime(h.Date)
	if err != nil {
		err = ErrNoDateHeader
		return
	}

	scheme, parametersStr, found := strings.Cut(h.Authorization, " ")
	if !found {
		err = fmt.Errorf("%w: unable to find scheme", ErrInvalidSignature)
		return
	} else if scheme != authScheme {
		err = fmt.Errorf("%w: unknown scheme", ErrInvalidSignature)
		return
	}

	// Extract the parameters parts
	parameters := strings.Split(parametersStr, ", ")
	if len(parameters) != expectedComponentCount {
		err = fmt.Errorf("%w: expected %d parameters", ErrInvalidSignature, expectedComponentCount)
		return
	}

	for _, parameter := range parameters {
		name, value, found := strings.Cut(parameter, "=")
		if !found {
			err = fmt.Errorf("%w: unable to find parameter name", ErrInvalidSignature)
			return
		}

		switch name {
		case "cred":
			// Unquote the value
			value, err = strconv.Unquote(value)
			if err != nil {
				err = fmt.Errorf("%w: unable to unquote credential string", ErrInvalidSignature)
				return
			}

			var date string
			keyID, appSlug, envName, date, err = parseCredentialString(value)
			if err != nil {
				return
			}

			// Verify the date matches the date header
			if date != timestamp.UTC().Format("20060102") {
				err = fmt.Errorf("%w: dates don't align", ErrInvalidSignature)
				return
			}

		case "op":
			operationHash = OperationHash(value)

		case "sig":
		// No need to do anything with the signature

		default:
			err = fmt.Errorf("%w: unknown parameter %q", ErrInvalidSignature, name)
			return
		}
	}

	return
}

// parseCredentialString parses the credential string from the authorization header and extracts the
// key ID, app slug, environment name, and date.
func parseCredentialString(str string) (keyID uint32, appSlug, envName string, date string, err error) {
	const expectedCredentialComponentCount = 4

	parts := strings.Split(str, "/")
	if len(parts) != expectedCredentialComponentCount {
		err = fmt.Errorf("%w: invalid credential string", ErrInvalidSignature)
		return
	}

	date = parts[0]
	appSlug = parts[1]
	envName = parts[2]

	keyID64, err := strconv.ParseUint(parts[3], 10, 32)
	if err != nil {
		err = fmt.Errorf("%w: invalid credential string: invalid key id", ErrInvalidSignature)
		return
	}
	keyID = uint32(keyID64)

	return
}
