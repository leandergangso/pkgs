# secret

Small, dependency-free helpers for generating and protecting secret values. The
caller owns all keys and storage; nothing is kept between calls.

```go
key := secret.NewKey()

token := secret.Token(32) // URL-safe random token
code := secret.Code(6)    // decimal login code

sig := secret.Sign(key, data)
if secret.Verify(key, data, sig) {
	// ...
}

blob, err := secret.Encrypt(key, data)
plain, err := secret.Decrypt(key, blob)

signed, err := secret.SignURL(key, "https://example.com/file", time.Hour)
err = secret.VerifyURL(key, signed)
```

## Storage

Store a digest, not the value: `Hash` for high-entropy tokens, or `Sign` for
short values such as email codes. Compare with `Compare`.

Signed URLs carry `exp` and `sig` query parameters. Verification checks the
signature before the expiry. Use `errors.Is(err, secret.ErrExpired)` to tell an
expired URL from an invalid one.

## Notes

- Random helpers panic if the system RNG fails.
- `Encrypt` uses AES-GCM with a fresh nonce. Keys must be 16, 24, or 32 bytes.
- This package is for small values: tokens, codes, URLs, and blobs. It is not a
  password hashing library.
