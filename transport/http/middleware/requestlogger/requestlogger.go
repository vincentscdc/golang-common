package requestlogger

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type ResponseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

func NewResponseWriterWrapper(w http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		bytes:          0,
	}
}

func (w *ResponseWriterWrapper) Status() int {
	return w.statusCode
}

func (w *ResponseWriterWrapper) BytesWritten() int {
	return w.bytes
}

func (w *ResponseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *ResponseWriterWrapper) Write(buf []byte) (int, error) {
	bytes, err := w.ResponseWriter.Write(buf)
	if err != nil {
		return 0, fmt.Errorf("failed to write: %w", err)
	}

	w.bytes += bytes

	return bytes, nil
}

// RequestLogger is a middleware that logs the start and end of each request, along
// with some useful data about what was requested, and some useful response data.
func RequestLogger(log *zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
			respWriterWrapper := NewResponseWriterWrapper(respWriter)

			scheme := "http"
			if req.TLS != nil {
				scheme = "https"
			}

			requestURL := fmt.Sprintf("%s://%s%s", scheme, req.Host, req.RequestURI)
			if strings.HasPrefix(req.RequestURI, scheme) {
				requestURL = req.RequestURI
			}

			start := time.Now()
			defer func() {
				log.Info().
					Str("method", req.Method).
					Str("url", requestURL).
					Str("path", req.URL.Path).
					Str("remote", req.RemoteAddr).
					Str("proto", req.Proto).
					Int("bytes", respWriterWrapper.BytesWritten()).
					Int("status", respWriterWrapper.Status()).
					Dur("duration", time.Since(start)).
					Send()
			}()

			next.ServeHTTP(respWriterWrapper, req)
		})
	}
}
