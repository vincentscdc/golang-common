package cryptouseruuid

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"
)

func Test_UserUUID(t *testing.T) {
	t.Parallel()

	type args struct {
		userUUID string
		handler  http.HandlerFunc
	}

	echoHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		userID, err := GetUserUUID(r.Context())
		if err != nil {
			t.Fatalf("failed to get user uuid %v", err.Error())
		}
		if _, err := w.Write([]byte(userID.String())); err != nil {
			t.Fatalf("unexpected write into response error: %v", err)
		}
	})

	tests := []struct {
		name                   string
		args                   args
		expectedHTTPStatusCode int
		expectedBody           string
	}{
		{
			name: "happy path",
			args: args{
				userUUID: "b7202eb0-5bf0-475d-8ee2-d3d2c168a5d5",
				handler:  echoHandler,
			},
			expectedHTTPStatusCode: http.StatusOK,
			expectedBody:           "b7202eb0-5bf0-475d-8ee2-d3d2c168a5d5",
		},
		{
			name: "response 401 when header is not set",
			args: args{
				handler: echoHandler,
			},
			expectedHTTPStatusCode: http.StatusUnauthorized,
		},
		{
			name: "response 401 and log error when user id format is invalid",
			args: args{
				userUUID: "xxx",
				handler:  echoHandler,
			},
			expectedHTTPStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var logMsg bytes.Buffer
			log := zerolog.New(&logMsg)

			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set(HTTPHeaderKeyUserUUID, tt.args.userUUID)
			rr := httptest.NewRecorder()

			userUUID := UserUUID(&log)

			handler := userUUID(tt.args.handler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedHTTPStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedHTTPStatusCode)
			}

			if strings.TrimSpace(rr.Body.String()) != tt.expectedBody {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), tt.expectedBody)
			}
		})
	}
}

func Test_SetUserUUID(t *testing.T) {
	t.Parallel()

	userUUID, _ := uuid.NewV4()
	newUserUUID, _ := uuid.NewV4()

	tests := []struct {
		name        string
		userUUID    *uuid.UUID
		newUserUUID *uuid.UUID
		want        *uuid.UUID
	}{
		{
			name:        "happy path",
			userUUID:    nil,
			newUserUUID: &userUUID,
			want:        &userUUID,
		},
		{
			name:        "overwrite existing key",
			userUUID:    &userUUID,
			newUserUUID: &newUserUUID,
			want:        &newUserUUID,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			if tt.userUUID != nil {
				ctx = context.WithValue(context.Background(), contextValKeyUserUUID, tt.userUUID)
			}

			ctx = SetUserUUID(ctx, tt.newUserUUID)

			if ctx.Value(contextValKeyUserUUID) != tt.want {
				t.Errorf("wrong userUUID in context = %v, want %v", ctx.Value(contextValKeyUserUUID), tt.want)
			}
		})
	}
}

func Test_GetUserUUID(t *testing.T) {
	t.Parallel()

	userUUID, _ := uuid.NewV4()

	tests := []struct {
		name                  string
		userUUID              any
		expectedUserUUID      *uuid.UUID
		expectedErrorResponse error
	}{
		{
			name:                  "happy path",
			userUUID:              &userUUID,
			expectedUserUUID:      &userUUID,
			expectedErrorResponse: nil,
		},
		{
			name:                  "value is not uuid",
			userUUID:              "test",
			expectedUserUUID:      nil,
			expectedErrorResponse: UserIDNotFoundError{},
		},
		{
			name:                  "user id not found",
			userUUID:              nil,
			expectedUserUUID:      nil,
			expectedErrorResponse: UserIDNotFoundError{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			if tt.userUUID != nil {
				ctx = context.WithValue(context.Background(), contextValKeyUserUUID, tt.userUUID)
			}

			actualUserUUID, actualErr := GetUserUUID(ctx)

			if actualUserUUID != tt.expectedUserUUID {
				t.Errorf("expected user uuid: %v, got %v", tt.expectedUserUUID, actualUserUUID)
			}

			if actualErr != nil && !errors.Is(actualErr, tt.expectedErrorResponse) {
				t.Errorf("expected error: %v, got %v", tt.expectedErrorResponse, actualErr)
			}
		})
	}
}

func Test_UUIDMiddlewareNotSet(t *testing.T) {
	t.Parallel()

	echoHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := GetUserUUID(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}
		if _, err := w.Write([]byte(userID.String())); err != nil {
			t.Fatalf("unexpected write into response error: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()

	req := httptest.NewRequest("GET", "/", nil)
	setRequestHeaderUserID(req, uuid.Must(uuid.NewV4()).String())

	echoHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}
}

func setRequestHeaderUserID(r *http.Request, uuid string) {
	r.Header.Set(HTTPHeaderKeyUserUUID, uuid)
}

func BenchmarkUserUUID(b *testing.B) {
	uuid := "b7202eb0-5bf0-475d-8ee2-d3d2c168a5d5"

	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	var logMsg bytes.Buffer
	log := zerolog.New(&logMsg)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	req.Header.Set(HTTPHeaderKeyUserUUID, uuid)

	userUUID := UserUUID(&log)

	handler := userUUID(f)

	b.ResetTimer()

	for i := 0; i <= b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkSetUserUUID(b *testing.B) {
	ctx := context.Background()
	userUUID, _ := uuid.NewV4()

	b.ResetTimer()

	for i := 0; i <= b.N; i++ {
		ctx = SetUserUUID(ctx, &userUUID)
	}
}

func BenchmarkGetUserUUID(b *testing.B) {
	userUUID, _ := uuid.NewV4()

	ctx := context.WithValue(context.Background(), contextValKeyUserUUID, userUUID)

	b.ResetTimer()

	for i := 0; i <= b.N; i++ {
		GetUserUUID(ctx)
	}
}
