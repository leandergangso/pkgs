// Package secret provides small, dependency-free helpers for generating and
// protecting secret values. Keys are supplied by the caller; nothing is stored
// between calls.
package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

var (
	// ErrInvalid indicates a malformed or unauthenticated value.
	ErrInvalid = errors.New("secret: invalid value")
	// ErrExpired indicates an expired signed URL.
	ErrExpired = errors.New("secret: expired")
)

const (
	nonceSize = 12
	sigParam  = "sig"
	expParam  = "exp"
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

// Encrypt encrypts plaintext with key and returns a URL-safe string. A fresh
// nonce is generated for each call. Key must be 16, 24, or 32 bytes.
func Encrypt(key, plaintext []byte) (string, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}
	blob := make([]byte, nonceSize, nonceSize+len(plaintext)+gcm.Overhead())
	mustRandom(blob)
	blob = gcm.Seal(blob, blob[:nonceSize], plaintext, nil)
	return base64.RawURLEncoding.EncodeToString(blob), nil
}

// Decrypt decrypts a value produced by [Encrypt].
func Decrypt(key []byte, value string) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	blob, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, ErrInvalid
	}
	if len(blob) < nonceSize+gcm.Overhead() {
		return nil, ErrInvalid
	}
	plaintext, err := gcm.Open(nil, blob[:nonceSize], blob[nonceSize:], nil)
	if err != nil {
		return nil, ErrInvalid
	}
	return plaintext, nil
}

// SignURL returns rawurl with exp and sig query parameters. The URL expires
// after ttl, which must be positive.
func SignURL(key []byte, rawurl string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		return "", errors.New("secret: invalid ttl")
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", fmt.Errorf("secret: malformed url: %w", err)
	}
	q := u.Query()
	q.Set(expParam, strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))
	q.Del(sigParam)
	u.RawQuery = q.Encode()
	sig := Sign(key, []byte(u.String()))
	q.Set(sigParam, sig)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// VerifyURL checks the signature and expiry of a URL produced by [SignURL].
// The signature is checked before the expiry.
func VerifyURL(key []byte, rawurl string) error {
	u, err := url.Parse(rawurl)
	if err != nil {
		return ErrInvalid
	}
	q := u.Query()
	sig := q.Get(sigParam)
	exp := q.Get(expParam)
	if sig == "" || exp == "" {
		return ErrInvalid
	}
	q.Del(sigParam)
	u.RawQuery = q.Encode()
	if !Verify(key, []byte(u.String()), sig) {
		return ErrInvalid
	}
	deadline, err := strconv.ParseInt(exp, 10, 64)
	if err != nil || deadline < time.Now().Unix() {
		return ErrExpired
	}
	return nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("secret: invalid key: %w", err)
	}
	return cipher.NewGCM(block)
}

func mustRandom(b []byte) {
	if _, err := rand.Read(b); err != nil {
		panic("secret: random: " + err.Error())
	}
}
