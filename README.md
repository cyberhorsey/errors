# Errors Utility

This package provides utility methods for error handling in Go, such as error types, context, pointer, and detail. The utility methods build on the functionality in [github.com/pkg/errors](http://github.com/pkg/errors).

**Examples**

Create a new error of a specific type, with detail:

```go
doc, err := jsonapi.DecodeRequest(r, "account", req)
if err != nil {
  return errors.BadRequest.NewWithDetail("Invalid request body")
}
```

Wrap an error:

```go
err := errors.New("an_error")
err = errors.AddErrorContext(err, "field", "value")
wrappedErr := errors.BadRequest.Wrapf(err, "error %s", "1")
```

Handle errors of specific types with pointer/detail:

```go
switch errors.GetType(err) {
case errors.NotFound:
  status = http.StatusNotFound
  title = "Not Found"
case errors.InvalidParameter:
  status = http.StatusBadRequest
  title = "Invalid Parameter"
case errors.MissingParameter:
  status = http.StatusBadRequest
  title = "Missing Parameter"
case errors.Validation, errors.Public:
  title = "Resource Invalid"
  status = http.StatusUnprocessableEntity
case errors.BadRequest:
  title = "Bad Request"
  status = http.StatusBadRequest
case errors.Forbidden:
  title = "Forbidden"
  status = http.StatusForbidden
default:
  status = http.StatusInternalServerError
  title = "Internal Server Error"
}
pointer := errors.Pointer(err)
detail := errors.Detail(err)
```

Add/retrieve error context:

```go
err := errors.BadRequest.New("an_error")
errWithContext := errors.AddErrorContext(err, "field1", "the field is empty")
fmt.Println(errors.GetErrorContextValue(errWithContext, "field1"))
```

Get original error cause from wrapped error:

```go
cusErr := customError{}
wrappedErr := errors.Wrap(cusErr, "wrapped customError")
wrappedErr = errors.Wrap(wrappedErr, "outer wrapped customError")
wrappedErr = errors.Wrap(wrappedErr, "outer outer wrapped customError")
causeErr := errors.Cause(wrappedErr)
originalErr, ok := causeErr.(customError)
# originalErr == cusErr
```

See more examples in errors_test.go.
