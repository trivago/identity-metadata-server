package shared

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

type ErrorWithStatus struct {
	Message string
	Code    int
}

// Error returns the error message.
func (e ErrorWithStatus) Error() string {
	return e.Message
}

// NewErrorWithStatus creates a new error with a status code.
// It formats the message using fmt.Sprintf.
func NewErrorWithStatus(code int, format string, args ...any) error {
	return ErrorWithStatus{
		Message: fmt.Sprintf(format, args...),
		Code:    code,
	}
}

// WrapHTTPError wraps an error in an httpReturnableError.
func WrapErrorWithStatus(err error, code int) error {
	msg := "nil"
	if err != nil {
		msg = err.Error()
	}

	return ErrorWithStatus{
		Message: msg,
		Code:    code,
	}
}

// HttpError renders a given error to gin, using the status code from the
// error if there is one.
func HttpError(c *gin.Context, defaultStatus int, err error) {
	status := defaultStatus
	if httpErr, ok := err.(ErrorWithStatus); ok {
		status = httpErr.Code
	}
	HttpErrorString(c, status, err.Error())
}

// HttpErrorString renders a given error message to gin with the given status.
// This function has a similar signature to HttpError, so it can be used as a
// drop-in replacement.
func HttpErrorString(c *gin.Context, status int, errMsg string) {
	c.String(status, "%s\n", errMsg)
}

// WrapErrorf combines an existing error with a new formatted error message.
// It returns a new error that contains both the original error and the new
// formatted error message. If the original error is nil, it returns only the
// new formatted error message.
func WrapErrorf(err error, format string, args ...any) error {
	newError := fmt.Errorf(format, args...)
	if err == nil {
		return newError
	}
	return errors.Join(newError, err)
}
