# Go packages

A set of common reusable packages.

## Packages

- [errmux](./errmux) error-returning router built on `net/http`'s `ServeMux`.
- [httpreq](./httpreq) small generic helpers for common HTTP API requests.
- [secret](./secret) small crypto helpers for tokens, codes, hashes, signatures, encryption, and signed URLs.
- [totp](./totp) RFC 6238 one-time passwords and authenticator enrollment.
- [email](./email) MJML and plain-text email template rendering.
- [assert](./assert) invariant checks for production code.
- [oauth](./oauth) OAuth2/OIDC logins with PKCE and nonce by default.
- [retry](./retry) bounded retries with exponential backoff and jitter.
- [log](./log) slog defaults and request-scoped loggers.
- [validate](./validate) field-keyed validation for user input.
