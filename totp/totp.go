// Package totp implements RFC 6238 time-based one-time passwords with the
// interoperable defaults used by authenticator apps: SHA-1, a 30-second
// period, and six digits.
package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	period    = 30
	modulus   = 1_000_000
	secretLen = 20
)

// NewSecret returns a new 20-byte secret.
func NewSecret() []byte {
	b := make([]byte, secretLen)
	if _, err := rand.Read(b); err != nil {
		panic("totp: random: " + err.Error())
	}
	return b
}

// Generate returns the code for secret at time t.
func Generate(secret []byte, t time.Time) string {
	return generate(secret, t.Unix()/period)
}

// Validate reports whether code is valid for secret at time t. The skew is the
// number of adjacent periods accepted on each side; use 1 for typical clock
// drift.
func Validate(secret []byte, code string, t time.Time, skew int) bool {
	if skew < 0 {
		panic("totp: negative skew")
	}
	step := t.Unix() / period
	for i := -skew; i <= skew; i++ {
		counter := step + int64(i)
		if counter < 0 {
			continue
		}
		if hmac.Equal([]byte(code), []byte(generate(secret, counter))) {
			return true
		}
	}
	return false
}

// URI returns an otpauth:// URL for enrolling secret in an authenticator app.
func URI(secret []byte, issuer, account string) string {
	label := escapeLabel(account)
	if issuer != "" {
		label = escapeLabel(issuer) + ":" + label
	}
	q := url.Values{}
	q.Set("secret", base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret))
	if issuer != "" {
		q.Set("issuer", issuer)
	}
	q.Set("algorithm", "SHA1")
	q.Set("digits", "6")
	q.Set("period", strconv.Itoa(period))
	return "otpauth://totp/" + label + "?" + q.Encode()
}

// escapeLabel escapes one component of the "issuer:account" label. PathEscape
// leaves ':' intact, which would make the label ambiguous, so escape it too.
func escapeLabel(s string) string {
	return strings.ReplaceAll(url.PathEscape(s), ":", "%3A")
}

func generate(secret []byte, counter int64) string {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(counter))
	mac := hmac.New(sha1.New, secret)
	_, _ = mac.Write(buf[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	n := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	return fmt.Sprintf("%06d", n%modulus)
}
