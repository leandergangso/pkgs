// Package validate collects field errors for user input. It is deliberately
// manual: callers list the checks, and [Validator.Err] reports a field-keyed
// error. For programming invariants, use an assert package instead.
package validate

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"
)

var _ error = Errors{}

// Errors maps field names to their first failure message.
type Errors map[string]string

// Error implements [error]. Fields are sorted for stable output.
func (e Errors) Error() string {
	fields := make([]string, 0, len(e))
	for field := range e {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		parts = append(parts, field+": "+e[field])
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

// Validator accumulates field errors. The zero value is ready to use.
type Validator struct {
	errors Errors
}

// New returns an empty Validator.
func New() *Validator { return &Validator{} }

func (v *Validator) add(field, msg string) {
	if v.errors == nil {
		v.errors = make(Errors)
	}
	if _, exists := v.errors[field]; !exists {
		v.errors[field] = msg
	}
}

// Err returns nil if no check failed, otherwise the collected [Errors]. Use
// errors.As to recover the field map.
func (v *Validator) Err() error {
	if len(v.errors) == 0 {
		return nil
	}
	return v.errors
}

// Check records msg for field if ok is false. It is the escape hatch for
// rules the package does not provide.
func (v *Validator) Check(field string, ok bool, msg string) {
	if !ok {
		v.add(field, msg)
	}
}

// Required fails if value is empty or only whitespace.
func (v *Validator) Required(field, value string) {
	v.Check(field, strings.TrimSpace(value) != "", "is required")
}

// Email fails unless value is a bare address. It is a syntax check only and
// does not test deliverability.
func (v *Validator) Email(field, value string) {
	parsed, err := mail.ParseAddress(value)
	ok := err == nil && parsed.Address == value && parsed.Name == ""
	v.Check(field, ok, "must be a valid email address")
}

// Min fails if value has fewer than n characters.
func (v *Validator) Min(field, value string, n int) {
	v.Check(field, utf8.RuneCountInString(value) >= n,
		fmt.Sprintf("must be at least %d characters", n))
}

// Max fails if value has more than n characters.
func (v *Validator) Max(field, value string, n int) {
	v.Check(field, utf8.RuneCountInString(value) <= n,
		fmt.Sprintf("must be at most %d characters", n))
}

// OneOf fails unless value is one of allowed.
func (v *Validator) OneOf(field, value string, allowed ...string) {
	v.Check(field, slices.Contains(allowed, value),
		"must be one of: "+strings.Join(allowed, ", "))
}

// Range fails unless n is within [min, max].
func (v *Validator) Range(field string, n, min, max int) {
	v.Check(field, n >= min && n <= max,
		fmt.Sprintf("must be between %d and %d", min, max))
}

// Match fails unless value matches re.
func (v *Validator) Match(field, value string, re *regexp.Regexp) {
	v.Check(field, re.MatchString(value), "has an invalid format")
}

// URL fails unless value parses as an absolute http or https URL with a host.
func (v *Validator) URL(field, value string) {
	u, err := url.Parse(value)
	ok := err == nil && u.IsAbs() && u.Host != "" && (u.Scheme == "http" || u.Scheme == "https")
	v.Check(field, ok, "must be a valid URL")
}
