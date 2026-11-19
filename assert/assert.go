// Package assert checks programming invariants in production code.
//
// A failed assertion panics with a [Violation], which carries a message and
// the call site. Use assertions only for conditions that indicate a bug in the
// program: unreachable branches, violated preconditions, and impossible states.
// Never use them to validate user input or expected errors; return an error for
// those.
package assert

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
)

// Violation describes a failed assertion.
type Violation struct {
	Message string
	File    string
	Line    int
}

// Error implements [error].
func (v *Violation) Error() string {
	msg := v.Message
	if msg == "" {
		msg = "assertion failed"
	}
	if v.File != "" {
		return fmt.Sprintf("assert: %s (%s:%d)", msg, v.File, v.Line)
	}
	return "assert: " + msg
}

// True panics unless ok is true.
func True(ok bool, format string, args ...any) {
	if !ok {
		panic(newViolation(2, format, args...))
	}
}

// NotNil panics unless value is non-nil. It detects typed nils such as a nil
// pointer stored in an interface.
func NotNil(value any, format string, args ...any) {
	if isNil(value) {
		panic(newViolation(2, format, args...))
	}
}

// NoError panics unless err is nil. Use it only for errors that cannot happen
// unless the program is wrong.
func NoError(err error) {
	if err != nil {
		panic(newViolation(2, "unexpected error: %v", err))
	}
}

// Fail panics unconditionally. Use it for branches that cannot be reached.
func Fail(format string, args ...any) {
	panic(newViolation(2, format, args...))
}

func newViolation(skip int, format string, args ...any) *Violation {
	_, file, line, ok := runtime.Caller(skip)
	v := &Violation{Message: fmt.Sprintf(format, args...)}
	if ok {
		v.File = filepath.Base(file)
		v.Line = line
	}
	return v
}

func isNil(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	}
	return false
}
