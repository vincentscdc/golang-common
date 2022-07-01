package cryptouseruuid

import (
	"context"
	"errors"
	"net/http"

	"github.com/monacohq/golang-common/transport/http/handlerwrap"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"
)

const (
	HTTPHeaderKeyUserUUID = "X-CRYPTO-USER-UUID"
)

type contextValKey string

const (
	contextValKeyUserUUID contextValKey = "authenticatedUserUUID"
)

var ErrUserIDNotFound = errors.New("user id not found")

// UserUUID is a middleware to get the user uuid from HTTP header.
// Set it into ctx otherwise abort with a 401 HTTP status code
func UserUUID(log *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
			uuidVal := req.Header.Get(HTTPHeaderKeyUserUUID)
			if uuidVal == "" {
				respWriter.WriteHeader(http.StatusUnauthorized)
				logWithField := log.With().Fields(map[string]string{
					"path":   req.URL.Path,
					"method": req.Method,
				}).Logger()

				logWithField.Error().Msg("uuid is empty")

				return
			}

			userUUID, err := uuid.FromString(uuidVal)
			if err != nil {
				log.Error().
					Err(err).
					Str("UserID", uuidVal).
					Msg("invalid user id")
				respWriter.WriteHeader(http.StatusUnauthorized)

				return
			}

			next.ServeHTTP(respWriter, req.WithContext(setUserUUID(req.Context(), &userUUID)))
		})
	}
}

func setUserUUID(ctx context.Context, userUUID *uuid.UUID) context.Context {
	return context.WithValue(ctx, contextValKeyUserUUID, userUUID)
}

func getUserUUID(ctx context.Context) (*uuid.UUID, *handlerwrap.ErrorResponse) {
	userUUID, ok := ctx.Value(contextValKeyUserUUID).(*uuid.UUID)
	if ok {
		return userUUID, nil
	}

	return nil,
		handlerwrap.NewErrorResponse(ErrUserIDNotFound, make(map[string]string),
			http.StatusUnauthorized, "", "user id not found")
}
