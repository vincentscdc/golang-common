package handlerwrap

import (
	"net/http"

	"github.com/rs/zerolog"
)

// Response is a wrapper for the response body.
type Response struct {
	Headers    map[string]string
	Body       any
	StatusCode int
}

func (hr *Response) render(log *zerolog.Logger, respW http.ResponseWriter, _ *http.Request) {
	render(
		log,
		hr.Headers,
		hr.StatusCode,
		hr.Body,
		respW,
	)
}
