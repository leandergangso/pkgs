package secret

import (
	"crypto/rand"
	"encoding/base64"
)

// Token returns a URL-safe random string containing byteLen random bytes.
func Token(byteLen int) string {
	if byteLen <= 0 {
		panic("secret: invalid token length")
	}
	b := make([]byte, byteLen)
	mustRandom(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// Code returns a random decimal code with exactly digits characters.
func Code(digits int) string {
	if digits <= 0 {
		panic("secret: invalid code length")
	}
	b := make([]byte, digits)
	mustRandom(b)
	for i := range b {
		for b[i] >= 250 {
			mustRandom(b[i : i+1])
		}
		b[i] = '0' + b[i]%10
	}
	return string(b)
}

// NewKey returns a new 32-byte key.
func NewKey() []byte {
	b := make([]byte, 32)
	mustRandom(b)
	return b
}

func mustRandom(b []byte) {
	if _, err := rand.Read(b); err != nil {
		panic("secret: random: " + err.Error())
	}
}
