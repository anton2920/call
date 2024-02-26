package main

import (
	"fmt"
	"net/http"
	"runtime"
)

type HTTPError struct {
	StatusCode int
	Message    string
}

var (
	ForbiddenError     = NewHTTPError(http.StatusForbidden, "whoops... Your permissions are insufficient")
	ReloadPageError    = NewHTTPError(http.StatusBadRequest, "whoops... Something went wrong. Please reload this page or try again")
	TryAgainLaterError = NewHTTPError(http.StatusInternalServerError, "whoops... Something went wrong. Please try again later")
)

func (e HTTPError) Error() string {
	return e.Message
}

func NewHTTPError(status int, message string) HTTPError {
	return HTTPError{StatusCode: status, Message: message}
}

func WrapErrorWithTrace(err error) error {
	return WrapErrorWithTraceSkip(err, 2)
}

func WrapErrorWithTraceSkip(err error, skip int) error {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return err
	}
	return fmt.Errorf("%s:%d: %w", file, line, err)
}
