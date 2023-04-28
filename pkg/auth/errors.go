package auth

import (
	"errors"
)

var (
	ErrNoAuthorizationHeader = errors.New("no authorization header provided")
	ErrNoDateHeader          = errors.New("no date header or invalid format provided")
	ErrAuthenticationExpired = errors.New("authentication expired")
	ErrAuthenticationFailed  = errors.New("authentication failed")
	ErrInvalidSignature      = errors.New("invalid signature")
)
