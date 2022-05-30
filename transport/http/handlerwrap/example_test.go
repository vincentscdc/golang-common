package handlerwrap

import (
	"context"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/rs/zerolog"
)

// Wrapping a POST http handler.
func Example_post() {
	zl := zerolog.New(io.Discard).With()

	type postRequest struct {
		Name string `json:"name"`
	}

	createHandler := func() TypedHandler {
		return func(r *http.Request) (*Response, *ErrorResponse) {
			var pr postRequest

			if err := BindBody(r, &pr); err != nil {
				return nil, err
			}

			log.Println(pr)

			return &Response{
				Body:       pr,
				Headers:    make(map[string]string),
				StatusCode: http.StatusCreated,
			}, nil
		}
	}

	logger := zl.Logger()

	Wrapper(&logger, createHandler()).ServeHTTP(nil, nil)
}

// Wrapping a GET http handler.
func Example_get() {
	zl := zerolog.New(io.Discard).With()

	getter := func(ctx context.Context, key string) (string, *ErrorResponse) {
		if key == "id" {
			return "1", nil
		}

		return "", MissingParamError{Name: key}.ToErrorResponse()
	}

	getHandler := func(nupg NamedURLParamsGetter) TypedHandler {
		return func(r *http.Request) (*Response, *ErrorResponse) {
			idParam, errR := nupg(r.Context(), "id")
			if errR != nil {
				return nil, errR
			}

			id, err := strconv.ParseInt(idParam, 10, 64)
			if err != nil {
				return nil, ParsingParamError{
					Name:  "id",
					Value: idParam,
				}.ToErrorResponse()
			}

			return &Response{
				Body:       id,
				Headers:    make(map[string]string),
				StatusCode: http.StatusOK,
			}, nil
		}
	}

	logger := zl.Logger()

	Wrapper(&logger, getHandler(getter)).ServeHTTP(nil, nil)
}
