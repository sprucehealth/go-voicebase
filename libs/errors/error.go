// Package errors provides helper functions to attach context to errors that
// allow for easier debugging. The downside is that it masks the original error
// if a receiving function doesn't know to expect errors from this package (which
// wrap the original error). Therefore, this package is mainly useful for applications
// rather than packages.
package errors

import (
	"fmt"
	"strings"
)

type aerr struct {
	err         error // actual error
	trace       []string
	annotations []string
}

func wrap(err error) aerr {
	if e, ok := err.(aerr); ok {
		return e
	}
	return aerr{err: err}
}

// Error implements the error interface.
func (e aerr) Error() string {
	es := e.err.Error()
	if len(e.annotations) != 0 {
		es += " (" + strings.Join(e.annotations, ", ") + ")"
	}
	if len(e.trace) != 0 {
		es += " [" + strings.Join(e.trace, ", ") + "]"
	}
	return es
}

// Cause returns the original error ignoring any wrapped errors (traces).
func Cause(e error) error {
	if e, ok := e.(aerr); ok {
		return e.err
	}
	return e
}

// Errorf formats according to a format specifier and returns the string as a value that satisfies error.
func Errorf(format string, args ...interface{}) error {
	return trace(fmt.Errorf(format, args...), 1)
}
