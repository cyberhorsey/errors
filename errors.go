// Package errors wraps github.com/pkg/errors with some custom error functionality.
// Originally based on https://hackernoon.com/golang-handling-errors-gracefully-8e27f1db729f with
// some heavy modifications and bug fixes.
package errors

import (
	stderrors "errors"
	"fmt"
	"strings"

	pkgerrors "github.com/pkg/errors"
)

// ErrorType is the type of an error
type ErrorType uint

const (
	// NoType error
	NoType ErrorType = iota
	// NotFound error
	NotFound
	// InvalidParameter error
	InvalidParameter
	// MissingParameter error
	MissingParameter
	// Validation error
	Validation
	// Forbidden error
	Forbidden
	// Public error
	Public
	// BadRequest error
	BadRequest
	// Unauthorized error
	Unauthorized
)

type customError struct {
	errorType ErrorType
	// originalError may be a customError or other error. Storing this is necessary for our Unwrap
	// method to work properly.
	originalError error
	// pkgError is always the github.com/pkg/errors error, used primarily for stack traces
	// error cause, and Error() messages.
	pkgError error
	context  errorContext
}

// Is provides a custom implementation of stderrors.Is(), which reports whether any error in
// customError's chain matches err.
//
// An error is considered to match if customError.originalError or customError.pkgError is
// equal to error.
//
// This is necessary to handle the special cases of stderrors.Is() with our pkg/errors wrapping for
// stack traces and Error() messages.
func (e *customError) Is(err error) bool {
	target := e.originalError

	for {
		if target == err {
			return true
		}

		if pkgerrors.Is(e.pkgError, err) {
			return true
		}

		if target = stderrors.Unwrap(target); target == nil {
			return false
		}
	}
}

// Unwrap returns the next error in the chain
func (e *customError) Unwrap() error {
	return e.originalError
}

// Format delegates the fmt.Formatter support to the pkgError for Stack Trace support from
// github.com/pkg/errors.
func (e *customError) Format(s fmt.State, verb rune) {
	if ferr, ok := e.pkgError.(fmt.Formatter); ok {
		ferr.Format(s, verb)
	}
}

// Error returns the message of a customError
func (e customError) Error() string {
	errKey := Key(&e)
	errDetail := Detail(&e)
	origErr := e.pkgError.Error()

	// Avoid duplication of detail/original error message e.g. from NewWithDetail()
	components := []string{}
	if !strings.Contains(origErr, errKey) {
		components = append(components, errKey)
	}

	if !strings.Contains(origErr, errDetail) {
		components = append(components, errDetail)
	}

	components = append(components, origErr)

	nonEmptyComps := make([]string, 0)

	for _, c := range components {
		if c != "" {
			nonEmptyComps = append(nonEmptyComps, c)
		}
	}

	return strings.Join(nonEmptyComps, ": ")
}

type errorContext map[string]string

// New creates a new customError
func (errorType ErrorType) New(msg string) error {
	origErr := pkgerrors.New(msg)

	return &customError{
		errorType:     errorType,
		originalError: origErr,
		pkgError:      origErr,
	}
}

// NewWithDetail creates a new customError with detail
func (errorType ErrorType) NewWithDetail(msg string) error {
	return WithDetail(errorType.New(msg), msg)
}

// NewWithKeyAndDetail creates a new customError with a name
func (errorType ErrorType) NewWithKeyAndDetail(key string, msg string) error {
	return WithKeyAndDetail(errorType.New(msg), key, msg)
}

// Newf creates a new customError with formatted message
func (errorType ErrorType) Newf(msg string, args ...interface{}) error {
	origErr := pkgerrors.New(fmt.Sprintf(msg, args...))

	return &customError{
		errorType:     errorType,
		originalError: origErr,
		pkgError:      origErr,
	}
}

// NewWithDetailf creates a new customError with formatted detail
func (errorType ErrorType) NewWithDetailf(msg string, args ...interface{}) error {
	return WithDetail(errorType.Newf(msg, args...), fmt.Sprintf(msg, args...))
}

// Wrap creates a new wrapped error
func (errorType ErrorType) Wrap(err error, msg string) error {
	return errorType.Wrapf(err, msg)
}

// Wrapf creates a new wrapped error with formatted message
func (errorType ErrorType) Wrapf(err error, msg string, args ...interface{}) error {
	if customErr, ok := err.(*customError); ok {
		return &customError{
			errorType:     errorType,
			originalError: err,
			pkgError:      pkgerrors.Wrapf(customErr.pkgError, msg, args...),
			context:       customErr.context,
		}
	}

	return &customError{
		errorType:     errorType,
		originalError: err,
		pkgError:      pkgerrors.Wrapf(err, msg, args...),
	}
}

// New creates a NoType error
func New(msg string) error {
	origErr := pkgerrors.New(msg)

	return &customError{
		errorType:     NoType,
		originalError: origErr,
		pkgError:      origErr,
	}
}

// Newf creates a no type error with formatted message
func Newf(msg string, args ...interface{}) error {
	origErr := pkgerrors.New(fmt.Sprintf(msg, args...))

	return &customError{
		errorType:     NoType,
		originalError: origErr,
		pkgError:      origErr,
	}
}

// Wrap an error with a string
func Wrap(err error, msg string) error {
	return Wrapf(err, msg)
}

// Wrapf an error with format string
func Wrapf(err error, msg string, args ...interface{}) error {
	if customErr, ok := err.(*customError); ok {
		return customErr.errorType.Wrapf(err, msg, args...)
	}

	return NoType.Wrapf(err, msg, args...)
}

// WithCause wraps causeErr with err. This is useful for wrapping an internal error
// with a sentinel error, for example.
func WithCause(err, causeErr error) error {
	mergedErrContext := make(errorContext)

	var errorType ErrorType

	if customErr, ok := causeErr.(*customError); ok {
		for k, v := range customErr.context {
			mergedErrContext[k] = v
		}

		errorType = customErr.errorType
	}

	if customErr, ok := err.(*customError); ok {
		for k, v := range customErr.context {
			mergedErrContext[k] = v
		}

		if customErr.errorType != NoType {
			errorType = customErr.errorType
		}
	}

	return &customError{
		errorType:     errorType,
		originalError: err,
		pkgError:      pkgerrors.Wrap(causeErr, err.Error()),
		context:       mergedErrContext,
	}
}

// Cause gives the original error
func Cause(err error) error {
	if customErr, ok := err.(*customError); ok {
		return Cause(pkgerrors.Cause(customErr.pkgError))
	}

	return pkgerrors.Cause(err)
}

// AddErrorContext adds a context to an error
func AddErrorContext(err error, key, message string) error {
	var context errorContext
	if customErr, ok := err.(*customError); ok {
		context = customErr.context
		if context == nil {
			context = make(errorContext)
		}

		context[key] = message

		return &customError{
			errorType:     customErr.errorType,
			originalError: customErr.originalError,
			pkgError:      customErr.pkgError,
			context:       context,
		}
	}

	context = errorContext{key: message}

	return &customError{
		errorType:     NoType,
		originalError: err,
		pkgError:      err,
		context:       context,
	}
}

// GetErrorContext returns the error context
func GetErrorContext(err error) map[string]string {
	if customErr, ok := err.(*customError); ok {
		return customErr.context
	}

	return nil
}

// GetErrorContextValue returns an error context value
func GetErrorContextValue(err error, key string) string {
	if errContext := GetErrorContext(err); errContext != nil {
		return errContext[key]
	}

	return ""
}

// GetType returns the error type
func GetType(err error) ErrorType {
	if customErr, ok := err.(*customError); ok {
		return customErr.errorType
	}

	return NoType
}

// WithPointer adds a pointer to the error
func WithPointer(err error, pointer string) error {
	return AddErrorContext(err, "pointer", pointer)
}

// WithDetail adds detail to the error
func WithDetail(err error, detail string) error {
	return AddErrorContext(err, "detail", detail)
}

// WithKey adds key to the error
func WithKey(err error, key string) error {
	return AddErrorContext(err, "key", key)
}

// WithKeyAndDetail adds key and detail to a message
func WithKeyAndDetail(err error, key string, detail string) error {
	err = WithKey(err, key)
	return WithDetail(err, detail)
}

// Pointer returns the error pointer
func Pointer(err error) string {
	return GetErrorContextValue(err, "pointer")
}

// Detail returns the error detail
func Detail(err error) string {
	return GetErrorContextValue(err, "detail")
}

// Key returns the error key
func Key(err error) string {
	return GetErrorContextValue(err, "key")
}

// WithFailFast signifies that err is fail fast (i.e. not resolvable by retries)
func WithFailFast(err error) error {
	return AddErrorContext(err, "failfast", "true")
}

// IsFailFast returns whether the error is fail fast (i.e. not resolvable by retries)
func IsFailFast(err error) bool {
	return GetErrorContextValue(err, "failfast") == "true"
}
