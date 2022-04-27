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

	tests := []struct {
		name           string
		requestBody    any
		accept         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "happy path",
			requestBody: struct {
				Test int `json:"test"`
			}{Test: 123},
			accept:         "application/json",
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"test":123}`,
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
				Body:           tt.requestBody,
				HTTPStatusCode: tt.expectedStatus,
			}

			hr.render(zl.Logger(), nr, req)

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
