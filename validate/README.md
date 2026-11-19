# validate

Collects field errors for user input. No reflection, no struct tags, no
dependencies. Callers list checks; `Err` returns a field-keyed error suitable
for a 400 response.

```go
v := validate.New()
v.Required("name", in.Name)
v.Email("email", in.Email)
v.Min("password", in.Password, 8)
v.Range("age", in.Age, 18, 120)
if err := v.Err(); err != nil {
	var fields validate.Errors
	if errors.As(err, &fields) {
		// fields["email"] == "must be a valid email address"
	}
	return err
}
```

Built-in rules: `Required`, `Email`, `Min`, `Max`, `OneOf`, `Range`, `Match`,
and `URL`. `Email` is a syntax-only check, like `email.Valid`. `Min` and `Max`
count Unicode characters, not bytes.

Only the first failure for each field is kept, so later checks on an invalid
field do not overwrite the first message. `Errors.Error()` sorts fields for
stable output.

For custom rules, use `Check`:

```go
v.Check("username", !strings.Contains(value, " "), "must not contain spaces")
```

Use this package for user input only. Programming invariants belong in an
assert package, and expected errors should be returned, not collected.
