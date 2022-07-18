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

	"github.com/monacohq/golang-common/transport/http/middleware/cryptouseruuid"
	"github.com/rs/zerolog"
)

func TestErrorResponse_render(t *testing.T) {
	t.Parallel()

	zl := zerolog.New(io.Discard).With()

	responseHeaders := map[string]string{"x-frame-options": "DENY", "x-content-type-options": "nosniff"}

	tests := []struct {
		name            string
		accept          string
		expectedStatus  int
		expectedBody    string
		expectedHeaders map[string]string
	}{
		{
			name:            "happy path",
			accept:          "application/json",
			expectedStatus:  http.StatusBadRequest,
			expectedBody:    `{"error":"test_render","error_message":"test error user"}`,
			expectedHeaders: responseHeaders,
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
				Err:          fmt.Errorf("test render"),
				Headers:      responseHeaders,
				StatusCode:   http.StatusBadRequest,
				Error:        "test_render",
				ErrorMessage: "test error user",
			}

			logger := zl.Logger()

			her.render(&logger, nr, req)

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

			for header, headerValue := range tt.expectedHeaders {
				if resp.Header.Get(header) != headerValue {
					t.Errorf("expected response header %s: %s, got %s: %s", header, headerValue, header, resp.Header.Get(header))

					return
				}
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
		Err            error
		HTTPStatusCode int
		Error          string
		ErrorMessage   string
	}

	testErr := errors.New("test render")

	refE := &ErrorResponse{
		Err:          testErr,
		StatusCode:   http.StatusBadRequest,
		Error:        "test_render",
		ErrorMessage: "test error user",
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
					Err:          testErr,
					StatusCode:   http.StatusBadRequest,
					Error:        "test_render",
					ErrorMessage: "test error user",
				},
			},
			want: true,
		},
		{
			name: "diff error",
			args: args{
				e1: &ErrorResponse{
					Err:          fmt.Errorf("diff"),
					StatusCode:   http.StatusBadRequest,
					Error:        "test_render",
					ErrorMessage: "test error user",
				},
			},
			want: false,
		},
		{
			name: "diff http status code",
			args: args{
				e1: &ErrorResponse{
					Err:          testErr,
					StatusCode:   http.StatusInternalServerError,
					Error:        "test_render",
					ErrorMessage: "test error user",
				},
			},
			want: false,
		},
		{
			name: "diff error code",
			args: args{
				e1: &ErrorResponse{
					Err:          testErr,
					StatusCode:   http.StatusBadRequest,
					Error:        "diff",
					ErrorMessage: "test error user",
				},
			},
			want: false,
		},
		{
			name: "diff error msg",
			args: args{
				e1: &ErrorResponse{
					Err:          testErr,
					StatusCode:   http.StatusBadRequest,
					Error:        "test_render",
					ErrorMessage: "diff",
				},
			},
			want: false,
		},
		{
			name: "diff L10NError",
			args: args{
				e1: &ErrorResponse{
					Err:          testErr,
					StatusCode:   http.StatusBadRequest,
					Error:        "test_render",
					ErrorMessage: "test error user",
					L10NError: &L10NError{
						TitleKey:   "title",
						MessageKey: "messge",
					},
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
				Err:          InternalServerError{Err: testErr},
				StatusCode:   http.StatusInternalServerError,
				Error:        "internal_error",
				ErrorMessage: "internal error",
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
				Err:          NotFoundError{Designation: "v"},
				StatusCode:   http.StatusNotFound,
				Error:        "not_found",
				ErrorMessage: "no corresponding `v` has been found",
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

func TestErrorResponse_AddHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		headers         map[string]string
		newHeaders      map[string]string
		expectedHeaders map[string]string
	}{
		{
			name:            "all new headers",
			headers:         map[string]string{"elle": "a"},
			newHeaders:      map[string]string{"il": "a"},
			expectedHeaders: map[string]string{"elle": "a", "il": "a"},
		},
		{
			name:            "overwrite headers",
			headers:         map[string]string{"elle": "a"},
			newHeaders:      map[string]string{"elle": "b"},
			expectedHeaders: map[string]string{"elle": "b"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			her := &ErrorResponse{
				Headers: tt.headers,
			}

			her.AddHeaders(tt.newHeaders)

			if len(her.Headers) != len(tt.expectedHeaders) {
				t.Errorf("wrong headers = %v, want %v", her.Headers, tt.expectedHeaders)
			}

			// check if all headers have the right value and are here
			for k, v := range tt.expectedHeaders {
				foundV, ok := her.Headers[k]
				if !ok {
					t.Errorf("header %s expected but not found", k)

					return
				}

				if foundV != v {
					t.Errorf("header %s has value %s, exected %s", k, foundV, v)
				}
			}
		})
	}
}

func TestNewErrorResponseFromCryptoUserUUIDError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		statusCode int
	}{
		{
			name:       "user id not found",
			err:        cryptouseruuid.UserIDNotFoundError{},
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "unknown error",
			err:        errors.New("unknown"),
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			errResp := NewErrorResponseFromCryptoUserUUIDError(tt.err)

			if errResp.StatusCode != tt.statusCode {
				t.Errorf("expected: %v, actual: %v", tt.statusCode, errResp.StatusCode)
			}
		})
	}
}
