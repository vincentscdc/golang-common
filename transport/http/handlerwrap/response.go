package handlerwrap

import (
	"net/http"

	"github.com/rs/zerolog"
)

// Response is a wrapper for the response body.
type Response struct {
	Body           any
	HTTPStatusCode int
}

func (hr *Response) render(log zerolog.Logger, respW http.ResponseWriter, req *http.Request) {
	render(
		log,
		req.Header.Get("Accept"),
		hr.HTTPStatusCode,
		hr.Body,
		respW,
	)
}
