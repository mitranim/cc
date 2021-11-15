package cc

import (
	"errors"
	"fmt"
	"strings"
)

/*
Combines multiple errors. Returned by `cc.Conc.All`.

While `cc.Errs` does implement the `error` interface, you should never cast it
to `error`. Instead, call the method `cc.Errs.Err`, which will correctly return
a nil interface when all errors are nil.
*/
type Errs []error

// Implement the `error` interface, combining the messages of all non-nil errors.
func (self Errs) Error() string {
	switch self.CountNonNil() {
	case 0:
		return ``

	case 1:
		return self.First().Error()

	default:
		return self.format()
	}
}

// Implement a hidden interface in the "errors" package. Alias for `cc.Errs.First`.
func (self Errs) Unwrap() error { return self.First() }

/*
Implement a hidden interface in the "errors" package. Tries `errors.Is` on every
non-nil error until one succeeds.
*/
func (self Errs) Is(err error) bool {
	for _, val := range self {
		if val != nil && errors.Is(val, err) {
			return true
		}
	}
	return false
}

/*
Implement a hidden interface in the "errors" package. Tries `errors.As` on every
non-nil error until one succeeds.
*/
func (self Errs) As(out interface{}) bool {
	for _, val := range self {
		if val != nil && errors.As(val, out) {
			return true
		}
	}
	return false
}

/*
If all errors are nil, returns nil. If there's exactly one non-nil error,
returns it as-is. Otherwise, returns self. This is the only correct way to
convert `cc.Errs` to `error`.
*/
func (self Errs) Err() error {
	switch self.CountNonNil() {
	case 0:
		return nil

	case 1:
		return self.First()

	default:
		return self
	}
}

// True if at least one error is non-nil.
func (self Errs) HasSome() bool {
	return self.CountNonNil() > 0
}

// Returns the amount of errors that satisfy the given function.
func (self Errs) Count(fun func(error) bool) (count int) {
	if fun != nil {
		for _, val := range self {
			if fun(val) {
				count++
			}
		}
	}
	return
}

// Returns the amount of nil errors.
func (self Errs) CountNil() int {
	return self.Count(isErrNil)
}

// Returns the amount of non-nil errors.
func (self Errs) CountNonNil() int {
	return self.Count(isErrNonNil)
}

// Finds the first error that satisfies the given test.
func (self Errs) Find(fun func(error) bool) error {
	if fun == nil {
		return nil
	}

	for _, val := range self {
		if val != nil && fun(val) {
			return val
		}
	}
	return nil
}

// Returns the first non-nil error.
func (self Errs) First() error {
	return self.Find(isErrNonNil)
}

func (self Errs) format() string {
	var buf strings.Builder
	buf.WriteString(`[cc] multiple errors`)

	for _, val := range self {
		if val == nil {
			continue
		}
		buf.WriteString(`; `)
		buf.WriteString(val.Error())
	}

	return buf.String()
}

/*
Treats an arbitrary non-`error` value as an `error`. Should be used for caught
non-`error` panics. Should not be used to wrap `error`.
*/
type nonErr [1]interface{}

// Implement `error`.
func (self nonErr) Error() string {
	if self[0] != nil {
		return fmt.Sprint(self[0])
	}
	return ``
}

// Implement hidden interface in "errors".
func (self nonErr) Unwrap() error {
	err, _ := self[0].(error)
	return err
}

func toErr(val interface{}) error {
	if val == nil {
		return nil
	}
	err, _ := val.(error)
	if err != nil {
		return err
	}
	return nonErr{val}
}

func isErrNil(err error) bool    { return err == nil }
func isErrNonNil(err error) bool { return err != nil }
