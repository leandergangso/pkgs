// Package email renders MJML and plain-text email templates with shared data.
//
// It does not send mail. Use [net/smtp] or a delivery service for that.
package email

import (
	"bytes"
	"fmt"
	"net/mail"
	"strings"
	"text/template"

	"github.com/preslavrachev/gomjml/mjml"
)

// Template is a parsed pair of MJML and plain-text templates.
type Template struct {
	mjml *template.Template
	text *template.Template
}

// Parse parses an MJML template and a plain-text template. Both use
// [text/template] syntax. The MJML template also provides an escape function
// for untrusted text; see [Escape].
func Parse(name, mjmlSource, textSource string) (*Template, error) {
	funcs := template.FuncMap{"escape": Escape}
	html, err := template.New(name + ".mjml").Funcs(funcs).Parse(mjmlSource)
	if err != nil {
		return nil, fmt.Errorf("email: parse mjml template: %w", err)
	}
	plain, err := template.New(name + ".txt").Parse(textSource)
	if err != nil {
		return nil, fmt.Errorf("email: parse text template: %w", err)
	}
	return &Template{mjml: html, text: plain}, nil
}

// Render executes both templates with data and compiles the MJML into HTML.
// It returns the HTML and plain-text bodies. Templates may be rendered
// concurrently.
func (t *Template) Render(data any) (html, text string, err error) {
	var mjmlBuf, textBuf bytes.Buffer
	if err := t.mjml.Execute(&mjmlBuf, data); err != nil {
		return "", "", fmt.Errorf("email: render mjml template: %w", err)
	}
	if err := t.text.Execute(&textBuf, data); err != nil {
		return "", "", fmt.Errorf("email: render text template: %w", err)
	}
	html, err = mjml.Render(mjmlBuf.String())
	if err != nil {
		return "", "", fmt.Errorf("email: compile mjml: %w", err)
	}
	return html, textBuf.String(), nil
}

var escapeReplacer = strings.NewReplacer(
	"&", "&#38;",
	"<", "&#60;",
	">", "&#62;",
	`"`, "&#34;",
	"'", "&#39;",
)

// Escape makes s safe to insert as text in an MJML template, for example
// {{.Name | escape}}. It uses numeric entity references because the MJML
// compiler decodes named XML entities before rendering mj-text, which would
// turn an escaped &lt; back into markup. The built-in html function is not
// sufficient.
//
// Escape does not make values safe for URL attributes. Validate the scheme of
// untrusted URLs before placing them in href or src.
func Escape(s string) string {
	return escapeReplacer.Replace(s)
}

// Valid reports whether addr is a bare email address. It is a syntax check
// only: it does not verify that the address exists or can receive mail.
// Display names, comments, and multiple addresses are rejected.
func Valid(addr string) bool {
	if addr == "" || strings.ContainsAny(addr, "\r\n") {
		return false
	}
	parsed, err := mail.ParseAddress(addr)
	return err == nil && parsed.Address == addr && parsed.Name == ""
}
