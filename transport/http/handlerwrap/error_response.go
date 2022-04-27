package handlerwrap

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
)

// ErrorResponse is a wrapper for the error response body to have a clean way of displaying errors.
type ErrorResponse struct {
	Error          error  `json:"-"`
	HTTPStatusCode int    `json:"-"`
	ErrorCode      string `json:"error_code"`
	ErrorMsg       string `json:"error_msg"`
}

// NewErrorResponse creates a new ErrorResponse.
func NewErrorResponse(
	e error,
	hsc int,
	ec string,
	msg string,
) *ErrorResponse {
	return &ErrorResponse{
		Error:          e,
		HTTPStatusCode: hsc,
		ErrorCode:      ec,
		ErrorMsg:       msg,
	}
}

func (her *ErrorResponse) render(log zerolog.Logger, w http.ResponseWriter, r *http.Request) {
	render(
		log,
		r.Header.Get("Accept"),
		her.HTTPStatusCode,
		her,
		w,
	)
}

// IsEqual checks if an error response is equal to another.
func (her *ErrorResponse) IsEqual(e1 *ErrorResponse) bool {
	if !errors.Is(e1.Error, her.Error) {
		return false
	}

	if e1.HTTPStatusCode != her.HTTPStatusCode {
		return false
	}

	if e1.ErrorCode != her.ErrorCode {
		return false
	}

	if e1.ErrorMsg != her.ErrorMsg {
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
	return NewErrorResponse(e, http.StatusInternalServerError, "internal_error", "internal error")
}

// NotFoundError is an error that is returned when a resource is not found.
type NotFoundError struct {
	Designation string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("no corresponding `%s` has been found", e.Designation)
}

func (e NotFoundError) ToErrorResponse() *ErrorResponse {
	return NewErrorResponse(e, http.StatusNotFound, "not_found", e.Error())
}
