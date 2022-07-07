package handlerwrap

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestOKStyle(t *testing.T) {
	t.Parallel()

	type arg struct {
		name             string
		bodyName         string
		handler          TypedHandler
		expectedResponse *Response
		expectedBody     string
	}

	testcases := []arg{
		{
			name:     "successful response",
			bodyName: "data",
			handler: func(r *http.Request) (*Response, *ErrorResponse) {
				return &Response{
					Body: map[string]any{
						"hello": "world",
					},
					Headers:    map[string]string{},
					StatusCode: http.StatusOK,
				}, nil
			},
			expectedResponse: &Response{
				Body: map[string]any{
					"data": map[string]any{
						"hello": "world",
					},
					"ok": true,
				},
				Headers:    map[string]string{},
				StatusCode: http.StatusOK,
			},
			expectedBody: `{"data":{"hello":"world"},"ok":true}`,
		},
		{
			name:     "failed response",
			bodyName: "data",
			handler: func(r *http.Request) (*Response, *ErrorResponse) {
				return &Response{
						Body: map[string]any{
							"hello": "world",
						},
						Headers:    map[string]string{},
						StatusCode: http.StatusNotFound,
					}, &ErrorResponse{
						Error:      NotFoundError{Designation: "v"},
						StatusCode: http.StatusNotFound,
						ErrorCode:  "10000",
						ErrorMsg:   "not found",
					}
			},
			expectedResponse: &Response{
				Body: map[string]any{
					"error":         "10000",
					"error_message": "not found",
					"ok":            false,
				},
				Headers:    map[string]string{},
				StatusCode: http.StatusOK,
			},
			expectedBody: `{"error":"10000","error_message":"not found","ok":false}`,
		},
		{
			name:     "response body non json format",
			bodyName: "data",
			handler: func(r *http.Request) (*Response, *ErrorResponse) {
				return &Response{
					Body:       "hello",
					Headers:    map[string]string{},
					StatusCode: http.StatusOK,
				}, nil
			},
			expectedResponse: &Response{
				Body: map[string]any{
					"data": "hello",
					"ok":   true,
				},
				Headers:    map[string]string{},
				StatusCode: http.StatusOK,
			},
			expectedBody: `{"data":"hello","ok":true}`,
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			nr := httptest.NewRecorder()

			logger := zerolog.New(io.Discard)

			req := httptest.NewRequest("GET", "/", nil)

			resp, _ := OKStyle(tc.bodyName, tc.handler)(req)

			if !reflect.DeepEqual(resp.Body, tc.expectedResponse.Body) {
				t.Errorf("handler return wrong body: got %v, want %v", resp.Body, tc.expectedResponse.Body)
			}

			if resp.StatusCode != tc.expectedResponse.StatusCode {
				t.Errorf("handler return wrong status code: got %v, want %v", resp.StatusCode, tc.expectedResponse.StatusCode)
			}

			if !reflect.DeepEqual(resp.Headers, tc.expectedResponse.Headers) {
				t.Errorf("handler return wrong headers: got %v, want %v", resp.Headers, tc.expectedResponse.Headers)
			}

			resp.render(&logger, nr, req)

			rr := nr.Result()
			defer rr.Body.Close()

			body, _ := io.ReadAll(rr.Body)
			trimmedBody := strings.TrimSpace(string(body))
			if trimmedBody != tc.expectedBody {
				t.Errorf("wrong json body: got\n--%s--, want\n--%s--", tc.expectedBody, trimmedBody)
			}
		})
	}
}

func BenchmarkOKStyle(b *testing.B) {
	f := func(r *http.Request) (*Response, *ErrorResponse) {
		return &Response{
			Body: map[string]any{
				"hello": "world",
			},
			Headers:    make(map[string]string),
			StatusCode: http.StatusOK,
		}, nil
	}

	req := httptest.NewRequest("GET", "/", nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		OKStyle("data", f)(req)
	}
}
