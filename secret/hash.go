package secret

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// Hash returns a URL-safe SHA-256 digest of data. It is suitable for storing
// high-entropy tokens at rest. Use [Sign] for short or low-entropy values.
func Hash(data []byte) string {
	sum := sha256.Sum256(data)
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// Sign returns a URL-safe HMAC-SHA256 signature of data.
func Sign(key, data []byte) string {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// Verify reports whether sig is a valid signature of data.
func Verify(key, data []byte, sig string) bool {
	got, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	return hmac.Equal(got, mac.Sum(nil))
}

// Compare reports whether a and b are equal, using constant-time comparison.
func Compare(a, b string) bool {
	return hmac.Equal([]byte(a), []byte(b))
}
