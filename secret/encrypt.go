package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

const nonceSize = 12

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

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("secret: invalid key: %w", err)
	}
	return cipher.NewGCM(block)
}
