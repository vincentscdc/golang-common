package pginit_test

import (
	"context"
	"testing"
	"time"

	"github.com/monacohq/golang-common/database/pginit"
	"github.com/rs/zerolog"
)

func TestPGInit_StdConn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *pginit.Config
		wantErr bool
	}{
		{
			name: "expecting error if wrong user",
			config: &pginit.Config{
				Host:     testHost,
				Port:     testPort,
				User:     "err",
				Password: "postgres",
				Database: "datawarehouse",
			},
			wantErr: true,
		},
		{
			name: "expecting no error with custom connection setting",
			config: &pginit.Config{
				Host:         testHost,
				Port:         testPort,
				User:         "postgres",
				Password:     "postgres",
				Database:     "datawarehouse",
				MaxConns:     15,
				MaxIdleConns: 10,
				MaxLifeTime:  10 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.TODO()

			pgi, err := pginit.New(
				tt.config,
				pginit.WithLogLevel(zerolog.DebugLevel),
				pginit.WithLogger(&zerolog.Logger{}, "request-id"),
				pginit.WithDecimalType(),
				pginit.WithUUIDType(),
			)
			if err != nil {
				t.Errorf("unexpected error in test (%v)", err)
			}

			db, err := pgi.StdConn(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected err (%v) but got (%v)", tt.wantErr, err)
			}

			if !tt.wantErr && db.Stats().MaxOpenConnections != int(tt.config.MaxConns) {
				t.Errorf("expected (%v) but got (%v)", tt.config.MaxConns, db.Stats().MaxOpenConnections)
			}
		})
	}
}
