package handlerwrap

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
)

// ErrorResponse is a wrapper for the error response body to have a clean way of displaying errors.
type ErrorResponse struct {
	Error      error             `json:"-"`
	Headers    map[string]string `json:"-"`
	StatusCode int               `json:"-"`
	ErrorCode  string            `json:"error_code"`
	ErrorMsg   string            `json:"error_msg"`
}

// NewErrorResponse creates a new ErrorResponse.
func NewErrorResponse(
	err error,
	headers map[string]string,
	statusCode int,
	errCode string,
	msg string,
) *ErrorResponse {
	return &ErrorResponse{
		Error:      err,
		Headers:    headers,
		StatusCode: statusCode,
		ErrorCode:  errCode,
		ErrorMsg:   msg,
	}
}

// AddHeaders add the headers to the error response
// it will overwrite a header if it already present, but will leave others in place
func (her *ErrorResponse) AddHeaders(headers map[string]string) {
	for k, v := range headers {
		her.Headers[k] = v
	}
}

func (her *ErrorResponse) render(log *zerolog.Logger, respW http.ResponseWriter, req *http.Request) {
	render(
		log,
		her.Headers,
		her.StatusCode,
		her,
		respW,
	)
}

// IsEqual checks if an error response is equal to another.
func (her *ErrorResponse) IsEqual(errR1 *ErrorResponse) bool {
	if !errors.Is(errR1.Error, her.Error) {
		return false
	}

	if errR1.StatusCode != her.StatusCode {
		return false
	}

	if errR1.ErrorCode != her.ErrorCode {
		return false
	}

	if errR1.ErrorMsg != her.ErrorMsg {
		return false
	}

	return true
}

// InternalServerError is an error that is returned when an internal server error occurs.
type InternalServerError struct {
	Err error
}

func (e InternalServerError) Error() string {
	return e.Err.Error()
}

func (e InternalServerError) ToErrorResponse() *ErrorResponse {
	return NewErrorResponse(e, make(map[string]string), http.StatusInternalServerError, "internal_error", "internal error")
}

// NotFoundError is an error that is returned when a resource is not found.
type NotFoundError struct {
	Designation string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("no corresponding `%s` has been found", e.Designation)
}

func (e NotFoundError) ToErrorResponse() *ErrorResponse {
	return NewErrorResponse(e, make(map[string]string), http.StatusNotFound, "not_found", e.Error())
}
