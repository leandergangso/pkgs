# oauth

A thin, opinionated layer over `golang.org/x/oauth2` and
`github.com/coreos/go-oidc/v3/oidc`. Requires Go 1.24.4+.

Every flow uses the authorization code grant with PKCE. OIDC providers also get
a nonce. There is no implicit flow, and the package stores nothing: you keep the
returned `Session` until the callback.

```go
cfg := oauth.New(oauth.Google, clientID, clientSecret, redirectURL)

url, session := cfg.Auth()
// Store session keyed by session.State, then redirect to url.

// In the callback:
if err := session.CheckState(r.URL.Query().Get("state")); err != nil {
	return err
}
token, err := cfg.Exchange(ctx, r.URL.Query().Get("code"), session)
```

For OIDC providers, create one verifier at startup and check the ID token in
the callback:

```go
verifier, err := cfg.NewVerifier(ctx)
if err != nil {
	return err
}

id, err := verifier.Verify(ctx, token, session.Nonce)
if err != nil {
	return err
}
// id.Subject is the provider's stable user identifier.
```

The verifier checks signature, issuer, audience, and expiry, then compares the
nonce. Token refresh and user-info calls are left to `x/oauth2`.

## Providers

Built in: `Google`, `GitHub`, `Discord`, `GitLab`, `Microsoft`. Only providers
with OIDC discovery have an issuer set. Add any other provider with:

```go
p := oauth.Custom("my-provider", authURL, tokenURL, issuer, "openid", "email")
```

Providers that do not support PKCE generally ignore its parameters. Do not use
a provider that rejects unknown authorization parameters.
