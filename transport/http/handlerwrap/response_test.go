package handlerwrap

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestResponse_render(t *testing.T) {
	t.Parallel()

	zl := zerolog.New(io.Discard).With()

	responseHeaders := map[string]string{"x-frame-options": "DENY", "x-content-type-options": "nosniff"}

	tests := []struct {
		name            string
		requestBody     any
		accept          string
		expectedStatus  int
		expectedBody    string
		expectedHeaders map[string]string
	}{
		{
			name: "happy path",
			requestBody: struct {
				Test int `json:"test"`
			}{Test: 123},
			accept:          "application/json",
			expectedStatus:  http.StatusCreated,
			expectedBody:    `{"test":123}`,
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

			hr := &Response{
				Body:       tt.requestBody,
				StatusCode: tt.expectedStatus,
				Headers:    responseHeaders,
			}

			logger := zl.Logger()

			hr.render(&logger, nr, req)

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
