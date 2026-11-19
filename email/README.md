# email

Render MJML and plain-text email bodies from shared data. It does not send
mail; pair it with `net/smtp` or a delivery service. MJML compilation uses
[gomjml](https://github.com/preslavrachev/gomjml).

```go
const mjml = `<mjml><mj-body><mj-section><mj-column>
	<mj-text>Hello {{.Name | escape}}</mj-text>
</mj-column></mj-section></mj-body></mjml>`

const text = "Hello {{.Name}},\n\nWelcome aboard."

t, err := email.Parse("welcome", mjml, text)
if err != nil {
	return err
}

html, plain, err := t.Render(data)
```

Templates use `text/template` syntax and may be rendered concurrently. Escape
untrusted values in the MJML source with the provided `escape` function; nothing
is escaped automatically. The standard library's `html` function is not enough
because MJML decodes named entities such as `&lt;` back into markup. Validate
the scheme of untrusted URLs before using them in `href` or `src`.

## One format only

Either source passed to `Parse` may be empty. Use `RenderHTML` or `RenderText`
to render just one body; `Render` requires both templates.

```go
t, err := email.Parse("welcome", mjml, "")
html, err := t.RenderHTML(data)
```

`email.Valid(addr)` is a strict, syntax-only address check. It rejects display
names, comments, and multiple addresses, and does not test deliverability.
