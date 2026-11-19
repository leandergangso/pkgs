// Package email renders MJML and plain-text email templates with shared data.
//
// It does not send mail. Use [net/smtp] or a delivery service for that.
package email

import (
	"bytes"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"text/template"

	"github.com/preslavrachev/gomjml/mjml"
)

// Template is a parsed pair of MJML and plain-text templates. Either may be
// absent when only one format is needed.
type Template struct {
	mjml *template.Template
	text *template.Template
}

// Parse parses an MJML template and a plain-text template. Both use
// [text/template] syntax. The MJML template also provides an escape function
// for untrusted text; see [Escape].
//
// Either source may be empty to skip that format, but at least one must be set.
// [Template.Render] requires both, so use [Template.RenderHTML] or
// [Template.RenderText] for single-format templates.
func Parse(name, mjmlSource, textSource string) (*Template, error) {
	if mjmlSource == "" && textSource == "" {
		return nil, errors.New("email: no template source")
	}
	t := new(Template)
	if mjmlSource != "" {
		funcs := template.FuncMap{"escape": Escape}
		parsed, err := template.New(name + ".mjml").Funcs(funcs).Parse(mjmlSource)
		if err != nil {
			return nil, fmt.Errorf("email: parse mjml template: %w", err)
		}
		t.mjml = parsed
	}
	if textSource != "" {
		parsed, err := template.New(name + ".txt").Parse(textSource)
		if err != nil {
			return nil, fmt.Errorf("email: parse text template: %w", err)
		}
		t.text = parsed
	}
	return t, nil
}

// Render executes both templates with data and compiles the MJML into HTML.
// It returns the HTML and plain-text bodies. Both templates must have been
// provided to [Parse]. Templates may be rendered concurrently.
func (t *Template) Render(data any) (html, text string, err error) {
	if t.mjml == nil || t.text == nil {
		return "", "", errors.New("email: both templates are required")
	}
	html, err = t.RenderHTML(data)
	if err != nil {
		return "", "", err
	}
	text, err = t.RenderText(data)
	if err != nil {
		return "", "", err
	}
	return html, text, nil
}

// RenderHTML executes the MJML template with data and compiles it into HTML.
// It fails if the template has no MJML source.
func (t *Template) RenderHTML(data any) (string, error) {
	if t.mjml == nil {
		return "", errors.New("email: no mjml template")
	}
	var buf bytes.Buffer
	if err := t.mjml.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("email: render mjml template: %w", err)
	}
	html, err := mjml.Render(buf.String())
	if err != nil {
		return "", fmt.Errorf("email: compile mjml: %w", err)
	}
	return html, nil
}

// RenderText executes the plain-text template with data. It fails if the
// template has no text source.
func (t *Template) RenderText(data any) (string, error) {
	if t.text == nil {
		return "", errors.New("email: no text template")
	}
	var buf bytes.Buffer
	if err := t.text.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("email: render text template: %w", err)
	}
	return buf.String(), nil
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
// compiler decodes named entities such as &lt; back into markup before
// rendering mj-text; the standard library's html escaping is therefore not
// sufficient. Double-escaping with html would be safe but would show the raw
// entity to recipients instead of the intended character.
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
