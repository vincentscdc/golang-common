package handlerwrap

import (
	"net/http"

	"github.com/goccy/go-json"

	"github.com/rs/zerolog"
)

// TypedHandler is the handler that you are actually handling the response.
type TypedHandler func(r *http.Request) (*Response, *ErrorResponse)

// Wrapper will actually do the boring work of logging an error and render the response.
func Wrapper(
	log zerolog.Logger,
	f TypedHandler,
) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			res, err := f(r)
			if err != nil {
				log.Error().
					Err(err.Error).
					Str("ErrorCode", err.ErrorCode).
					Int("HTTPStatusCode", err.HTTPStatusCode).
					Msg(err.ErrorMsg)

				err.render(log, w, r)

				return
			}

			res.render(log, w, r)
		},
	)
}

func render(
	log zerolog.Logger,
	acceptHeader string, // nolint: unparam
	httpStatusCode int,
	responseBody interface{},
	w http.ResponseWriter,
) {
	// nolint: gocritic
	// LATER: add more encodings
	switch acceptHeader {
	default:
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(httpStatusCode)

		if err := json.NewEncoder(w).Encode(responseBody); err != nil {
			log.Error().Err(err).Msg("http render")

			w.WriteHeader(http.StatusInternalServerError)

			return
		}
	}
}
