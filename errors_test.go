package errors

import (
	stderrors "errors"
	"fmt"
	"testing"

	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestIs(t *testing.T) {
	originalErr := stderrors.New("original")                // go/errors
	wrappedErr1 := pkgerrors.Wrap(originalErr, "wrapped 1") // pkg/errors
	wrappedErr2 := Wrap(wrappedErr1, "wrapped 2")
	wrappedErr3 := fmt.Errorf("wrapped 3: %w", wrappedErr2)
	wrappedErr4 := Wrap(wrappedErr3, "wrapped 4")
	sentinel1 := Validation.New("sentinel1")
	sentinel2 := WithCause(sentinel1, wrappedErr4)
	wrappedSentinel := Wrap(sentinel2, "wrapped sentinel")

	assert.True(t, stderrors.Is(wrappedSentinel, sentinel2))
	assert.True(t, stderrors.Is(wrappedSentinel, sentinel1))
	assert.True(t, stderrors.Is(wrappedSentinel, wrappedErr4))
	assert.True(t, stderrors.Is(wrappedSentinel, wrappedErr3))
	assert.True(t, stderrors.Is(wrappedSentinel, wrappedErr2))
	assert.True(t, stderrors.Is(wrappedSentinel, wrappedErr1))
	assert.True(t, stderrors.Is(wrappedSentinel, originalErr))

	assert.False(t, stderrors.Is(sentinel1, nil))
	assert.False(t, stderrors.Is(sentinel1, wrappedErr4))

	assert.True(t, stderrors.Is(wrappedErr4, wrappedErr3))
	assert.True(t, stderrors.Is(wrappedErr4, wrappedErr2))
	assert.True(t, stderrors.Is(wrappedErr4, wrappedErr1))
	assert.True(t, stderrors.Is(wrappedErr4, originalErr))

	assert.True(t, stderrors.Is(wrappedErr3, wrappedErr2))
	assert.True(t, stderrors.Is(wrappedErr3, wrappedErr1))
	assert.True(t, stderrors.Is(wrappedErr3, originalErr))

	assert.True(t, stderrors.Is(wrappedErr2, wrappedErr1))
	assert.True(t, stderrors.Is(wrappedErr2, originalErr))

	assert.True(t, stderrors.Is(wrappedErr1, originalErr))
}

func TestUnwrap(t *testing.T) {
	originalErr := stderrors.New("original") // go/errors
	wrappedErr1 := Wrap(originalErr, "wrapped 1")
	wrappedErr2 := fmt.Errorf("wrapped 2: %w", wrappedErr1)
	wrappedErr3 := Wrap(wrappedErr2, "wrapped 3")

	// Unwrap must be tested with Is
	assert.True(t, stderrors.Is(wrappedErr3, wrappedErr2))
	assert.True(t, stderrors.Is(wrappedErr2, wrappedErr1))
	assert.True(t, stderrors.Is(wrappedErr1, originalErr))
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			"standardError",
			New("standard error"),
			"standard error",
		},
		{
			"wrappedError",
			Wrap(fmt.Errorf("standard error"), "wrapped error"),
			"wrapped error: standard error",
		},
		{
			"wrappedErrorWithDetail",
			WithDetail(
				Wrap(fmt.Errorf("standard error"), "wrapped error"),
				"Error Details",
			),
			"Error Details: wrapped error: standard error",
		},
		{
			"wrappedErrorWithNestedDetail",
			WithDetail(
				Wrap(
					BadRequest.NewWithDetail("standard error"),
					"wrapped error",
				),
				"Error Details",
			),
			"Error Details: wrapped error: standard error",
		},
		{
			"wrappedErrorWithKeyAndDetail",
			WithKey(
				WithDetail(
					Wrap(
						BadRequest.NewWithDetail("standard error"),
						"wrapped error",
					),
					"Error Details",
				),
				"ERR_KEY",
			),
			"ERR_KEY: Error Details: wrapped error: standard error",
		},
		{
			"causeErrorWithKeyAndDetail",
			Wrap(
				WithCause(
					BadRequest.NewWithKeyAndDetail("ERR_SENTINEL", "Sentinel error detail"),
					fmt.Errorf("standard error cause"),
				),
				"wrapped error",
			),
			"wrapped error: ERR_SENTINEL: Sentinel error detail: standard error cause",
		},
		{
			"newWithKeyAndDetail",
			BadRequest.NewWithKeyAndDetail("ERR_KEY", "Error Details"),
			"ERR_KEY: Error Details",
		},
		{
			"random",
			WithKeyAndDetail(
				Wrapf(
					pkgerrors.Wrap(fmt.Errorf("fmt error"), "wrapped pkgerror"),
					"wrapped error",
				),
				"ERR_KEY",
				"Error Details",
			),
			"ERR_KEY: Error Details: wrapped error: wrapped pkgerror: fmt error",
		},
	}

	for _, tt := range tests {
		got := tt.err.Error()
		assert.Equal(t, tt.want, got)
	}
}

func TestStackTrace(t *testing.T) {
	err := New("Inner error")
	wrappedErr := Wrap(err, "Outer error")
	trace := fmt.Sprintf("%+v", wrappedErr)
	assert.Contains(t, trace, "gitlab.com/gamestopcorp/platform/blockchain/nft-lib-errors%2egit.New")
	assert.Contains(t, trace, "gitlab.com/gamestopcorp/platform/blockchain/nft-lib-errors%2egit.Wrapf")
}

func TestErrorType_Wrapf(t *testing.T) {
	origErr := New("an_error")
	origErr = AddErrorContext(origErr, "field", "value")
	wrappedErr := BadRequest.Wrapf(origErr, "error %s", "1")
	wrappedErr = Wrapf(wrappedErr, "outer wrapped err %s", "1")

	wantContext := map[string]string{
		"field": "value",
	}
	assert.Equal(t, wantContext, GetErrorContext(wrappedErr))
	assert.Equal(t, BadRequest, GetType(wrappedErr))
	assert.EqualError(t, wrappedErr, "outer wrapped err 1: error 1: an_error")
}

func TestErrorType_NewWithDetail(t *testing.T) {
	err := Public.NewWithDetail("This is public error detail")
	wrappedErr := Wrap(err, "wrapped error")

	assert.Equal(t, Public, GetType(wrappedErr))
	assert.Equal(t, "This is public error detail", Detail(wrappedErr))
	assert.EqualError(t, wrappedErr, "wrapped error: This is public error detail")
}

func TestErrorType_NewWithKeyAndDetail(t *testing.T) {
	err := Public.NewWithKeyAndDetail("ERROR_KEY", "This is public error message")
	assert.Equal(t, "This is public error message", Detail(err))
	assert.Equal(t, "ERROR_KEY", Key(err))
}

func TestErrorType_NewWithDetailf(t *testing.T) {
	err := Public.NewWithDetailf("This is public %s detail", "error")
	wrappedErr := Wrap(err, "wrapped error")

	assert.Equal(t, Public, GetType(wrappedErr))
	assert.Equal(t, "This is public error detail", Detail(wrappedErr))
	assert.EqualError(t, wrappedErr, "wrapped error: This is public error detail")
}

func TestWrapf(t *testing.T) {
	err := New("an_error")
	err = AddErrorContext(err, "field", "value")
	wrappedErr := BadRequest.Wrapf(err, "error %s", "1")

	wantContext := map[string]string{
		"field": "value",
	}

	assert.Equal(t, BadRequest, GetType(wrappedErr))
	assert.Equal(t, wantContext, GetErrorContext(wrappedErr))
	assert.EqualError(t, wrappedErr, "error 1: an_error")
}

func TestWrapf_NoTypeError(t *testing.T) {
	err := Newf("an_error %s", "2")
	wrappedErr := Wrapf(err, "error %s", "1")

	assert.Equal(t, NoType, GetType(wrappedErr))
	assert.EqualError(t, wrappedErr, "error 1: an_error 2")
}

func TestGetErrorContext(t *testing.T) {
	err := BadRequest.New("an_error")
	errWithContext := AddErrorContext(err, "field1", "the field is empty")
	errWithContext = AddErrorContext(errWithContext, "field2", "the field is empty")

	wantContext := map[string]string{
		"field1": "the field is empty",
		"field2": "the field is empty",
	}

	assert.Equal(t, BadRequest, GetType(errWithContext))
	assert.Equal(t, wantContext, GetErrorContext(errWithContext))
	assert.Equal(t, err.Error(), errWithContext.Error())
	assert.Equal(t, "the field is empty", GetErrorContextValue(errWithContext, "field2"))
}

func TestGetErrorContext_standardError(t *testing.T) {
	err := fmt.Errorf("this is a standard error")
	assert.Equal(t, map[string]string(nil), GetErrorContext(err))
}

func TestGetErrorContext_NoTypeError(t *testing.T) {
	err := fmt.Errorf("this is a standard error")
	errWithContext := AddErrorContext(err, "field1", "the field is empty")
	errWithContext = AddErrorContext(errWithContext, "field2", "the field is empty")

	wantContext := map[string]string{
		"field1": "the field is empty",
		"field2": "the field is empty",
	}

	assert.Equal(t, NoType, GetType(errWithContext))
	assert.Equal(t, wantContext, GetErrorContext(errWithContext))
	assert.Equal(t, err.Error(), errWithContext.Error())
}

type causeTestError struct{}

func (e causeTestError) Error() string {
	return "Cause test error"
}

func TestCause(t *testing.T) {
	originalErr := causeTestError{}
	wrappedErr := Wrap(originalErr, "wrapped causeTestError")
	wrappedErr = Wrap(wrappedErr, "outer wrapped causeTestError")
	wrappedErr = Wrap(wrappedErr, "outer outer wrapped causeTestError")
	causeErr := Cause(wrappedErr)
	originalErr, ok := causeErr.(causeTestError)

	assert.Equal(t, true, ok)
	assert.Equal(t, causeErr, originalErr)
}

func TestGetType(t *testing.T) {
	originalErr := New("original error with no type")
	wrappedErr := Validation.Wrap(originalErr, "validation wrapped err")
	outerWrappedErr := Wrap(wrappedErr, "outer wrapped error")

	errType := GetType(outerWrappedErr)
	assert.Equal(t, Validation, errType)
	errType = GetType(stderrors.New("hi"))
	assert.Equal(t, NoType, errType)
}

func Test_WithCause(t *testing.T) {
	causeErr := New("an_error")
	causeErr = AddErrorContext(causeErr, "field", "value")
	causeErr = BadRequest.Wrapf(causeErr, "error %s", "1")
	causeErr = Wrapf(causeErr, "outer wrapped err %s", "1")

	sentinelErr := Validation.NewWithDetail("sentinelErr with detail")
	gotErr := WithCause(sentinelErr, causeErr)
	assert.Equal(
		t,
		map[string]string{
			"field":  "value",
			"detail": "sentinelErr with detail",
		},
		GetErrorContext(gotErr),
	)
	assert.Equal(t, Validation, GetType(gotErr))
	assert.EqualError(t, gotErr, "sentinelErr with detail: outer wrapped err 1: error 1: an_error")
	assert.True(t, stderrors.Is(gotErr, sentinelErr))

	standardSentinelErr := fmt.Errorf("sentinelErr")
	gotErr = WithCause(standardSentinelErr, causeErr)
	assert.Equal(t, map[string]string{"field": "value"}, GetErrorContext(gotErr))
	assert.Equal(t, BadRequest, GetType(gotErr))
	assert.EqualError(t, gotErr, "sentinelErr: outer wrapped err 1: error 1: an_error")
	assert.True(t, stderrors.Is(gotErr, standardSentinelErr))

	standardCauseErr := fmt.Errorf("an_error")
	gotErr = WithCause(sentinelErr, standardCauseErr)
	assert.Equal(t, map[string]string{"detail": "sentinelErr with detail"}, GetErrorContext(gotErr))
	assert.Equal(t, Validation, GetType(gotErr))
	assert.EqualError(t, gotErr, "sentinelErr with detail: an_error")
	assert.True(t, stderrors.Is(gotErr, sentinelErr))

	sentinel1 := fmt.Errorf("sentinel1")
	sentinel2 := pkgerrors.Wrap(sentinel1, "sentinel2")
	sentinel3 := fmt.Errorf("sentinel2")
	gotErr = WithCause(sentinel3, sentinel2)
	assert.True(t, stderrors.Is(gotErr, sentinel1))
	assert.True(t, stderrors.Is(gotErr, sentinel2))
	assert.True(t, stderrors.Is(gotErr, sentinel3))
}

func TestPointer(t *testing.T) {
	err := fmt.Errorf("this is an error")
	assert.Equal(t, "", Pointer(err))
	err = WithPointer(err, "thefield")
	assert.Equal(t, "thefield", Pointer(err))
}

func TestDetail(t *testing.T) {
	err := fmt.Errorf("this is an error")
	assert.Equal(t, "", Detail(err))
	err = WithDetail(err, "the detail")
	assert.Equal(t, "the detail", Detail(err))
}

func TestKey(t *testing.T) {
	err := fmt.Errorf("this is an error")
	assert.Equal(t, "", Key(err))
	err = WithKey(err, "ERROR_NAME")
	assert.Equal(t, "ERROR_NAME", Key(err))
}

func TestFailFast(t *testing.T) {
	err := fmt.Errorf("this is an error")
	assert.False(t, IsFailFast(err))
	err = WithFailFast(err)
	assert.True(t, IsFailFast(err))
}
