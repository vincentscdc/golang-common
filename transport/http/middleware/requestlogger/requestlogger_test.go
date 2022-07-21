package requestlogger

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	leak := flag.Bool("leak", false, "use leak detector")
	flag.Parse()

	if *leak {
		goleak.VerifyTestMain(m)

		return
	}

	os.Exit(m.Run())
}

func TestRequestLogger(t *testing.T) {
	t.Parallel()

	type RequestLoggerField struct {
		Method   string  `json:"method"`
		URL      string  `json:"url"`
		Path     string  `json:"path"`
		Remote   string  `json:"remote"`
		Proto    string  `json:"proto"`
		Bytes    int     `json:"bytes"`
		Status   int     `json:"status"`
		Duration float32 `json:"duration"`
	}

	echoHandler := func(status, bytes int) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)

			data := make([]byte, bytes)
			w.Write(data)
		})
	}

	testcases := []struct {
		name           string
		expectedMethod string
		expectedURL    string
		expectedStatus int
		expectedBytes  int
	}{
		{
			name:           "happy response 200 when method is GET",
			expectedMethod: "GET",
			expectedURL:    "http://example.com",
			expectedStatus: http.StatusOK,
			expectedBytes:  4,
		},
		{
			name:           "happy response 201 when method is POST",
			expectedMethod: "POST",
			expectedURL:    "https://example.com",
			expectedStatus: http.StatusCreated,
			expectedBytes:  16,
		},
	}

	for _, tt := range testcases {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var logMsg bytes.Buffer
			log := zerolog.New(&logMsg)

			rr := httptest.NewRecorder()

			logger := RequestLogger(&log)

			handler := logger(echoHandler(tt.expectedStatus, tt.expectedBytes))
			handler.ServeHTTP(rr, httptest.NewRequest(tt.expectedMethod, tt.expectedURL, nil))

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			var logResult RequestLoggerField

			err := json.Unmarshal(logMsg.Bytes(), &logResult)
			if err != nil {
				t.Errorf("logger format unmarshal failed: %v", err)
			}

			if logResult.Method != tt.expectedMethod {
				t.Errorf("logger info wrong method: got %v want %v", logResult.Method, tt.expectedMethod)
			}

			if logResult.URL != tt.expectedURL {
				t.Errorf("logger info wrong method: got %v want %v", logResult.URL, tt.expectedURL)
			}

			if logResult.Status != tt.expectedStatus {
				t.Errorf("logger info wrong status code: got %v want %v", logResult.Status, tt.expectedStatus)
			}

			if logResult.Bytes != tt.expectedBytes {
				t.Errorf("logger info wrong bytes: got %v want %v", logResult.Status, tt.expectedStatus)
			}
		})
	}
}

func BenchmarkRequestLogger(b *testing.B) {
	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		data := make([]byte, 4)
		w.Write(data)
	})

	log := zerolog.New(io.Discard)

	req := httptest.NewRequest("GET", "https://example.com", nil)
	rr := httptest.NewRecorder()

	logger := RequestLogger(&log)

	handler := logger(f)

	b.ResetTimer()

	for i := 0; i <= b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}
