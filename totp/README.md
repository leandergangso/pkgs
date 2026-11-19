# totp

RFC 6238 time-based one-time passwords with the interoperable defaults used by
authenticator apps: SHA-1, a 30-second period, and six digits. Requires Go
1.22+.

```go
secret := totp.NewSecret() // store this per user, encrypted

code := totp.Generate(secret, time.Now())
ok := totp.Validate(secret, code, time.Now(), 1) // ±1 period

uri := totp.URI(secret, "Example", "user@example.com")
```

`URI` returns an `otpauth://` URL for QR-code enrollment. Show it once, and
require a valid code before enabling 2FA.

## Storage and use

- Store each user's secret encrypted. A plaintext secret in the database
  defeats 2FA.
- Accept a small `skew` (usually 1) for clock drift.
- Rate-limit verification, and reject a code once it has been used so it cannot
  be replayed within its period.
- Generate backup codes with `secret.Code(10)` from the `secret` package, store
  only `secret.Hash(code)`, and compare with `secret.Compare`.
