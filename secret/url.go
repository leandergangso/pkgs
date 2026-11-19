package secret

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

const (
	sigParam = "sig"
	expParam = "exp"
)

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
