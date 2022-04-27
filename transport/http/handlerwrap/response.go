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

func (hr *Response) render(log zerolog.Logger, w http.ResponseWriter, r *http.Request) {
	render(
		log,
		r.Header.Get("Accept"),
		hr.HTTPStatusCode,
		hr.Body,
		w,
	)
}
