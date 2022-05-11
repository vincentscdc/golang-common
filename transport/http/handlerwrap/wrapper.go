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
		func(respW http.ResponseWriter, req *http.Request) {
			res, err := f(req)
			if err != nil {
				log.Error().
					Err(err.Error).
					Str("ErrorCode", err.ErrorCode).
					Int("HTTPStatusCode", err.HTTPStatusCode).
					Msg(err.ErrorMsg)

				err.render(log, respW, req)

				return
			}

			res.render(log, respW, req)
		},
	)
}

func render(
	log zerolog.Logger,
	acceptHeader string,
	httpStatusCode int,
	responseBody interface{},
	respW http.ResponseWriter,
) {
	// nolint: gocritic
	// LATER: add more encodings
	switch acceptHeader {
	default:
		respW.Header().Add("Content-Type", "application/json")
		respW.WriteHeader(httpStatusCode)

		if err := json.NewEncoder(respW).Encode(responseBody); err != nil {
			log.Error().Err(err).Msg("http render")

			respW.WriteHeader(http.StatusInternalServerError)

			return
		}
	}
}
