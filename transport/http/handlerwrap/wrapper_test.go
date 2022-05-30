package handlerwrap

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func TestWrapper(t *testing.T) {
	t.Parallel()

	zl := zerolog.New(io.Discard).With()

	tests := []struct {
		name              string
		httpResponse      *Response
		httpErrorResponse *ErrorResponse
		expectedStatus    int
	}{
		{
			name: "normal path",
			httpResponse: &Response{
				Body: struct {
					Test int `json:"test"`
				}{Test: 123},
				StatusCode: http.StatusCreated,
			},
			httpErrorResponse: nil,
			expectedStatus:    http.StatusCreated,
		},
		{
			name:         "error path",
			httpResponse: nil,
			httpErrorResponse: &ErrorResponse{
				Error:      fmt.Errorf("test render"),
				StatusCode: http.StatusBadRequest,
				ErrorCode:  "test_render",
				ErrorMsg:   "test error user",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", &bytes.Reader{})

			nr := httptest.NewRecorder()

			f := func(r *http.Request) (*Response, *ErrorResponse) {
				if tt.httpResponse != nil {
					return tt.httpResponse, nil
				}

				return nil, tt.httpErrorResponse
			}

			logger := zl.Logger()

			handler := Wrapper(&logger, f)

			handler.ServeHTTP(nr, req)

			resp := nr.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)

				return
			}
		})
	}
}

func BenchmarkHTTPWrapper(b *testing.B) {
	zl := zerolog.New(io.Discard).With()
	req := httptest.NewRequest(http.MethodGet, "/", &bytes.Reader{})
	nr := httptest.NewRecorder()

	f := func(r *http.Request) (*Response, *ErrorResponse) {
		return &Response{
			Body: struct {
				Test int
			}{
				Test: 123,
			},
			StatusCode: 200,
		}, nil
	}

	logger := zl.Logger()

	handler := Wrapper(&logger, f)

	b.ResetTimer()

	for i := 0; i <= b.N; i++ {
		handler.ServeHTTP(nr, req)
	}
}
