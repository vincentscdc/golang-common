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
	log *zerolog.Logger,
	f TypedHandler,
) http.HandlerFunc {
	return http.HandlerFunc(
		func(respW http.ResponseWriter, req *http.Request) {
			res, err := f(req)
			if err != nil {
				log.Error().
					Err(err.Err).
					Str("ErrorCode", err.Error).
					Int("HTTPStatusCode", err.StatusCode).
					Msg(err.ErrorMessage)

				err.render(log, respW, req)

				return
			}

			res.render(log, respW, req)
		},
	)
}

func render(
	log *zerolog.Logger,
	headers map[string]string,
	statusCode int,
	responseBody interface{},
	respW http.ResponseWriter,
) {
	acceptHeader, ok := headers["Accept"]
	if !ok {
		acceptHeader = ""
	}

	// nolint: gocritic // LATER: add more encodings to fix this
	switch acceptHeader {
	default:
		for header, headerValue := range headers {
			respW.Header().Add(header, headerValue)
		}

		respW.Header().Add("Content-Type", "application/json")
		respW.WriteHeader(statusCode)

		if err := json.NewEncoder(respW).Encode(responseBody); err != nil {
			log.Error().Err(err).Msg("http render")

			respW.WriteHeader(http.StatusInternalServerError)

			return
		}
	}
}
