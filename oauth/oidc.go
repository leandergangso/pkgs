package oauth

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// ErrNonceMismatch indicates that an ID token carried the wrong nonce.
var ErrNonceMismatch = errors.New("oauth: nonce mismatch")

// Verifier verifies ID tokens from one provider. Create one per application
// with [Config.NewVerifier]; it reuses the provider's signing keys.
type Verifier struct {
	inner *oidc.IDTokenVerifier
}

// NewVerifier discovers the provider's OIDC configuration and returns a
// verifier bound to this client's ID.
func (c *Config) NewVerifier(ctx context.Context) (*Verifier, error) {
	if c.issuer == "" {
		return nil, ErrNotOIDC
	}
	p, err := oidc.NewProvider(ctx, c.issuer)
	if err != nil {
		return nil, fmt.Errorf("oauth: discover %s: %w", c.issuer, err)
	}
	return &Verifier{inner: p.Verifier(&oidc.Config{ClientID: c.clientID})}, nil
}

// Verify checks the ID token's signature, issuer, audience, expiry, and nonce.
// nonce must come from the [Session] used to start this login.
func (v *Verifier) Verify(ctx context.Context, token *oauth2.Token, nonce string) (*oidc.IDToken, error) {
	raw, ok := token.Extra("id_token").(string)
	if !ok || raw == "" {
		return nil, ErrNoIDToken
	}
	id, err := v.inner.Verify(ctx, raw)
	if err != nil {
		return nil, err
	}
	if id.Nonce != nonce {
		return nil, ErrNonceMismatch
	}
	return id, nil
}
