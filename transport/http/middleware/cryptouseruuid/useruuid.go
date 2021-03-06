package cryptouseruuid

import (
	"context"
	"fmt"
	"net/http"

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

// UserIDNotFoundError is the error that is returned when there is no user uuid found in the ctx
type UserIDNotFoundError struct{}

func (m UserIDNotFoundError) Error() string {
	return "user id not found"
}

// UserUUIDInvalidError is the error that is returned when the user id is not a valid uuid
type UserUUIDInvalidError struct {
	submittedUUID string
}

func (m UserUUIDInvalidError) Error() string {
	return fmt.Sprintf("user id %s not a valid uuid", m.submittedUUID)
}

// UserUUID is a middleware to get the user uuid from HTTP header and set into ctx,
// otherwise abort with a 401 HTTP status code
func UserUUID(log *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
			uuidVal := req.Header.Get(HTTPHeaderKeyUserUUID)
			if uuidVal == "" {
				respWriter.WriteHeader(http.StatusUnauthorized)

				log.Error().
					Str("path", req.URL.Path).
					Str("method", req.Method).
					Err(UserIDNotFoundError{}).
					Msg(fmt.Sprintf("header %s not found in request context", HTTPHeaderKeyUserUUID))

				return
			}

			userUUID, err := uuid.FromString(uuidVal)
			if err != nil {
				log.Error().
					Err(UserUUIDInvalidError{submittedUUID: uuidVal}).
					Msg("invalid user id")

				respWriter.WriteHeader(http.StatusUnauthorized)

				return
			}

			next.ServeHTTP(respWriter, req.WithContext(SetUserUUID(req.Context(), &userUUID)))
		})
	}
}

// SetUserUUID returns a context in which the value associated with the key contextValKeyUserUUID is the provided uuid
func SetUserUUID(ctx context.Context, userUUID *uuid.UUID) context.Context {
	return context.WithValue(ctx, contextValKeyUserUUID, userUUID)
}

// GetUserUUID returns the uuid value associated with the context for key contextValKeyUserUUID
func GetUserUUID(ctx context.Context) (*uuid.UUID, error) {
	userUUID, ok := ctx.Value(contextValKeyUserUUID).(*uuid.UUID)
	if ok {
		return userUUID, nil
	}

	return nil, UserIDNotFoundError{}
}
