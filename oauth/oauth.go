// Package oauth provides a small, opinionated layer over [x/oauth2] and
// [go-oidc]. It always uses the authorization code flow with PKCE, and OIDC
// flows also carry a nonce.
//
// The caller stores a [Session] between the authorization request and the
// callback; this package keeps no state of its own.
//
// [x/oauth2]: https://pkg.go.dev/golang.org/x/oauth2
// [go-oidc]: https://pkg.go.dev/github.com/coreos/go-oidc/v3/oidc
package oauth

import (
	"context"
	"crypto/hmac"
	"errors"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	// ErrStateMismatch indicates that the callback state does not match.
	ErrStateMismatch = errors.New("oauth: state mismatch")
	// ErrNoIDToken indicates a token response without an OIDC ID token.
	ErrNoIDToken = errors.New("oauth: no id token")
	// ErrNotOIDC indicates an OIDC operation on a plain OAuth2 provider.
	ErrNotOIDC = errors.New("oauth: provider is not OIDC")
)

// Provider describes an OAuth2 or OIDC provider.
type Provider struct {
	Name     string
	Endpoint oauth2.Endpoint
	Scopes   []string
	Issuer   string // OIDC issuer URL; empty for plain OAuth2 providers.
}

// Custom returns a Provider for an arbitrary OAuth2 endpoint. Pass an issuer
// only if the provider supports OIDC discovery.
func Custom(name, authURL, tokenURL, issuer string, scopes ...string) Provider {
	return Provider{
		Name:     name,
		Endpoint: oauth2.Endpoint{AuthURL: authURL, TokenURL: tokenURL},
		Scopes:   scopes,
		Issuer:   issuer,
	}
}

// Config is a configured OAuth2 client.
type Config struct {
	config   *oauth2.Config
	clientID string
	issuer   string
}

// New configures a client for p.
func New(p Provider, clientID, clientSecret, redirectURL string) *Config {
	return &Config{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       p.Scopes,
			Endpoint:     p.Endpoint,
		},
		clientID: clientID,
		issuer:   p.Issuer,
	}
}

// Session holds the transient values that must survive the redirect. Store it
// keyed by State until the callback arrives.
type Session struct {
	State    string
	Verifier string // PKCE code verifier
	Nonce    string // OIDC nonce; empty for plain OAuth2
}

// CheckState reports whether got matches the session's state.
func (s Session) CheckState(got string) error {
	if got == "" || !hmac.Equal([]byte(got), []byte(s.State)) {
		return ErrStateMismatch
	}
	return nil
}

// Auth returns the authorization URL and the Session to store. It uses PKCE
// for every provider, and adds a nonce for OIDC providers.
func (c *Config) Auth() (string, Session) {
	s := Session{
		State:    oauth2.GenerateVerifier(),
		Verifier: oauth2.GenerateVerifier(),
	}
	opts := []oauth2.AuthCodeOption{oauth2.S256ChallengeOption(s.Verifier)}
	if c.issuer != "" {
		s.Nonce = oauth2.GenerateVerifier()
		opts = append(opts, oidc.Nonce(s.Nonce))
	}
	return c.config.AuthCodeURL(s.State, opts...), s
}

// Exchange exchanges code for a token using the session's PKCE verifier.
func (c *Config) Exchange(ctx context.Context, code string, s Session) (*oauth2.Token, error) {
	return c.config.Exchange(ctx, code, oauth2.VerifierOption(s.Verifier))
}

// IsOIDC reports whether the provider supports OIDC discovery.
func (c *Config) IsOIDC() bool {
	return c.issuer != ""
}

// AuthURL returns the provider's authorization endpoint. It is useful when
// constructing a redirect allowlist.
func (c *Config) AuthURL() string {
	return c.config.Endpoint.AuthURL
}
