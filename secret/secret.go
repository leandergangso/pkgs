// Package secret provides small, dependency-free helpers for generating and
// protecting secret values. Keys are supplied by the caller; nothing is stored
// between calls.
package secret

import "errors"

var (
	// ErrInvalid indicates a malformed or unauthenticated value.
	ErrInvalid = errors.New("secret: invalid value")
	// ErrExpired indicates an expired signed URL.
	ErrExpired = errors.New("secret: expired")
)
