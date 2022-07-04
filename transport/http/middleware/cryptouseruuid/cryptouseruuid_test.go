package cryptouseruuid

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/monacohq/golang-common/transport/http/handlerwrap"
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
		userID, err := getUserUUID(r.Context())
		if err != nil {
			t.Fatalf("failed to get user uuid %v", err.Error)
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
		expectedLogMsg         string
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
			expectedLogMsg:         `{"level":"error","error":"invalid UUID length: 3","UserID":"xxx","message":"invalid user id"}`,
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

func Test_setUserUUID(t *testing.T) {
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

			ctx = setUserUUID(ctx, tt.newUserUUID)

			if ctx.Value(contextValKeyUserUUID) != tt.want {
				t.Errorf("wrong userUUID in context = %v, want %v", ctx.Value(contextValKeyUserUUID), tt.want)
			}
		})
	}
}

func Test_getUserUUID(t *testing.T) {
	t.Parallel()

	userUUID, _ := uuid.NewV4()

	tests := []struct {
		name                  string
		userUUID              any
		expectedUserUUID      *uuid.UUID
		expectedErrorResponse *handlerwrap.ErrorResponse
	}{
		{
			name:                  "happy path",
			userUUID:              &userUUID,
			expectedUserUUID:      &userUUID,
			expectedErrorResponse: nil,
		},
		{
			name:             "value is not uuid",
			userUUID:         "test",
			expectedUserUUID: nil,
			expectedErrorResponse: &handlerwrap.ErrorResponse{
				Error:      ErrUserIDNotFound,
				StatusCode: http.StatusUnauthorized,
				ErrorCode:  "",
				ErrorMsg:   "user id not found",
			},
		},
		{
			name:             "user id not found",
			userUUID:         nil,
			expectedUserUUID: nil,
			expectedErrorResponse: &handlerwrap.ErrorResponse{
				Error:      ErrUserIDNotFound,
				StatusCode: http.StatusUnauthorized,
				ErrorCode:  "",
				ErrorMsg:   "user id not found",
			},
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

			actualUserUUID, actualErr := getUserUUID(ctx)

			if actualUserUUID != tt.expectedUserUUID {
				t.Errorf("expected user uuid: %v, got %v", tt.expectedUserUUID, actualUserUUID)
			}

			if actualErr != nil && !actualErr.IsEqual(tt.expectedErrorResponse) {
				t.Errorf("expected error: %v, got %v", tt.expectedErrorResponse, actualErr)
			}
		})
	}
}
