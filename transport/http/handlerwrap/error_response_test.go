package handlerwrap

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestErrorResponse_render(t *testing.T) {
	t.Parallel()

	zl := zerolog.New(io.Discard).With()

	tests := []struct {
		name           string
		accept         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "happy path",
			accept:         "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error_code":"test_render","error_msg":"test error user"}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", &bytes.Reader{})
			req.Header.Add("Accept", tt.accept)

			nr := httptest.NewRecorder()

			her := &ErrorResponse{
				Error:          fmt.Errorf("test render"),
				HTTPStatusCode: http.StatusBadRequest,
				ErrorCode:      "test_render",
				ErrorMsg:       "test error user",
			}

			her.render(zl.Logger(), nr, req)

			resp := nr.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)

				return
			}

			if resp.Header.Get("Content-Type") != tt.accept {
				t.Errorf("expected response header %s, got %s", tt.accept, resp.Header.Get("Content-Type"))

				return
			}

			body, _ := io.ReadAll(resp.Body)
			trimmedBody := strings.TrimSpace(string(body))
			if trimmedBody != tt.expectedBody {
				t.Errorf("expected body\n--%s--\ngot\n--%s--", tt.expectedBody, trimmedBody)

				return
			}
		})
	}
}

func TestErrorResponse_IsEqual(t *testing.T) {
	t.Parallel()

	type fields struct {
		Error          error
		HTTPStatusCode int
		ErrorCode      string
		ErrorMsg       string
	}

	testErr := errors.New("test render")

	refE := &ErrorResponse{
		Error:          testErr,
		HTTPStatusCode: http.StatusBadRequest,
		ErrorCode:      "test_render",
		ErrorMsg:       "test error user",
	}

	type args struct {
		e1 *ErrorResponse
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "equal",
			args: args{
				e1: &ErrorResponse{
					Error:          testErr,
					HTTPStatusCode: http.StatusBadRequest,
					ErrorCode:      "test_render",
					ErrorMsg:       "test error user",
				},
			},
			want: true,
		},
		{
			name: "diff error",
			args: args{
				e1: &ErrorResponse{
					Error:          fmt.Errorf("diff"),
					HTTPStatusCode: http.StatusBadRequest,
					ErrorCode:      "test_render",
					ErrorMsg:       "test error user",
				},
			},
			want: false,
		},
		{
			name: "diff http status code",
			args: args{
				e1: &ErrorResponse{
					Error:          testErr,
					HTTPStatusCode: http.StatusInternalServerError,
					ErrorCode:      "test_render",
					ErrorMsg:       "test error user",
				},
			},
			want: false,
		},
		{
			name: "diff error code",
			args: args{
				e1: &ErrorResponse{
					Error:          testErr,
					HTTPStatusCode: http.StatusBadRequest,
					ErrorCode:      "diff",
					ErrorMsg:       "test error user",
				},
			},
			want: false,
		},
		{
			name: "diff error msg",
			args: args{
				e1: &ErrorResponse{
					Error:          testErr,
					HTTPStatusCode: http.StatusBadRequest,
					ErrorCode:      "test_render",
					ErrorMsg:       "diff",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := refE.IsEqual(tt.args.e1); got != tt.want {
				t.Errorf("ErrorResponse.IsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInternalServerError_Error(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test render")

	type fields struct {
		Err error
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "happy path",
			fields: fields{Err: testErr},
			want:   "test render",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := InternalServerError{
				Err: testErr,
			}

			if got := e.Error(); got != tt.want {
				t.Errorf("InternalServerError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInternalServerError_ToErrorResponse(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test render")

	type fields struct {
		Err error
	}

	tests := []struct {
		name   string
		fields fields
		want   *ErrorResponse
	}{
		{
			name:   "happy path",
			fields: fields{Err: testErr},
			want: &ErrorResponse{
				Error:          InternalServerError{Err: testErr},
				HTTPStatusCode: http.StatusInternalServerError,
				ErrorCode:      "internal_error",
				ErrorMsg:       "internal error",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := InternalServerError{
				Err: tt.fields.Err,
			}

			if got := e.ToErrorResponse(); !got.IsEqual(tt.want) {
				t.Errorf("InternalServerError.ToErrorResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotFoundError_Error(t *testing.T) {
	t.Parallel()

	type fields struct {
		Designation string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "happy path",
			fields: fields{Designation: "v"},
			want:   "no corresponding `v` has been found",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NotFoundError{
				Designation: tt.fields.Designation,
			}

			if got := e.Error(); got != tt.want {
				t.Errorf("NotFoundError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotFoundError_ToErrorResponse(t *testing.T) {
	t.Parallel()

	type fields struct {
		Designation string
	}

	tests := []struct {
		name   string
		fields fields
		want   *ErrorResponse
	}{
		{
			name:   "happy path",
			fields: fields{Designation: "v"},
			want: &ErrorResponse{
				Error:          NotFoundError{Designation: "v"},
				HTTPStatusCode: http.StatusNotFound,
				ErrorCode:      "not_found",
				ErrorMsg:       "no corresponding `v` has been found",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NotFoundError{
				Designation: tt.fields.Designation,
			}

			if got := e.ToErrorResponse(); !got.IsEqual(tt.want) {
				t.Errorf("NotFoundError.ToErrorResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
