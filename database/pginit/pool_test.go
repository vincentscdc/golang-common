package pginit_test

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/monacohq/golang-common/database/pginit"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/rs/zerolog"
	"go.uber.org/goleak"
)

var testHost, testPort string // nolint: gochecknoglobals, nolintlint

func TestMain(m *testing.M) {
	leak := flag.Bool("leak", false, "use leak detector")
	flag.Parse()

	if *leak {
		goleak.VerifyTestMain(m)

		return
	}

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14",
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=datawarehouse",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	databaseURL := fmt.Sprintf("postgres://postgres:%s@%s/datawarehouse?sslmode=disable", "postgres", getHostPort(resource, "5432/tcp"))
	resource.Expire(180) // Tell docker to hard kill the container in 180 seconds
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	if err := pool.Retry(func() error {
		ctx := context.Background()
		db, err := pgx.Connect(ctx, databaseURL)
		if err != nil {
			return fmt.Errorf("pgx connect: %w", err)
		}
		if err := db.Ping(ctx); err != nil {
			return fmt.Errorf("ping: %w", err)
		}

		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker(%s): %s", databaseURL, err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func getHostPort(resource *dockertest.Resource, id string) string {
	dockerURL := os.Getenv("DOCKER_HOST")
	if dockerURL == "" {
		hostAndPort := resource.GetHostPort("5432/tcp")
		hp := strings.Split(hostAndPort, ":")
		testHost = hp[0]
		testPort = hp[1]

		return testHost + ":" + testPort
	}

	u, err := url.Parse(dockerURL)
	if err != nil {
		panic(err)
	}

	testHost = u.Hostname()
	testPort = resource.GetPort(id)

	return u.Hostname() + ":" + resource.GetPort(id)
}

func TestConnPool(t *testing.T) {
	t.Parallel()

	type want struct {
		Err    error
		Config pginit.Config
	}

	type args struct {
		Config pginit.Config
	}

	tests := []struct {
		name string
		args args
		want want
		err  error
	}{
		{
			name: "expecting no error with default connection setting",
			args: args{
				pginit.Config{
					Host:     testHost,
					Port:     testPort,
					User:     "postgres",
					Password: "postgres",
					Database: "datawarehouse",
				},
			},
			want: want{
				Err: nil,
				Config: pginit.Config{
					MaxConns:     25,
					MaxIdleConns: 25,
					MaxLifeTime:  5 * time.Minute,
				},
			},
		},
		{
			name: "expecting no error with custom connection setting",
			args: args{
				pginit.Config{
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
			want: want{
				Err: nil,
				Config: pginit.Config{
					MaxConns:     15,
					MaxIdleConns: 10,
					MaxLifeTime:  10 * time.Minute,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.TODO()

			pgi, err := pginit.New(&tt.args.Config, pginit.WithLogger(&zerolog.Logger{}, ""))
			if err != nil && !errors.Is(err, tt.want.Err) {
				t.Errorf("expected (%v) but got (%v)", tt.want.Err, err)
			}
			db, err := pgi.ConnPool(ctx)
			if err != nil && !errors.Is(err, tt.want.Err) {
				t.Errorf("expected (%v) but got (%v)", tt.want.Err, err)
			}

			if err := db.Ping(ctx); err != nil {
				t.Errorf("require no err but got (%v)", err)
			}

			if db.Config().MaxConns != tt.want.Config.MaxConns {
				t.Errorf("expected (%v) but got (%v)", tt.want.Config.MaxConns, db.Config().MaxConns)
			}
			if db.Config().MaxConnLifetime != tt.want.Config.MaxLifeTime {
				t.Errorf("expected (%v) but got (%v)", tt.want.Config.MaxLifeTime, db.Config().MaxConnLifetime)
			}
			if db.Config().MinConns != tt.want.Config.MaxIdleConns {
				t.Errorf("expected (%v) but got (%v)", tt.want.Config.MaxIdleConns, db.Config().MinConns)
			}
		})
	}
}

func TestConnPoolWithLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		lvl       zerolog.Level
		wantedLvl pgx.LogLevel
	}{
		{
			name:      "level debug",
			lvl:       zerolog.DebugLevel,
			wantedLvl: pgx.LogLevelDebug,
		},
		{
			name:      "level info",
			lvl:       zerolog.InfoLevel,
			wantedLvl: pgx.LogLevelInfo,
		},
		{
			name:      "level warn",
			lvl:       zerolog.WarnLevel,
			wantedLvl: pgx.LogLevelWarn,
		},
		{
			name:      "level error",
			lvl:       zerolog.ErrorLevel,
			wantedLvl: pgx.LogLevelError,
		},
		{
			name:      "level none",
			lvl:       zerolog.NoLevel,
			wantedLvl: pgx.LogLevelNone,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			pgi, err := pginit.New(
				&pginit.Config{
					Host:     testHost,
					Port:     testPort,
					User:     "postgres",
					Password: "postgres",
					Database: "datawarehouse",
					MaxConns: 2,
				},
				pginit.WithLogLevel(tt.lvl),
				pginit.WithLogger(&zerolog.Logger{}, "request-id"),
				pginit.WithDecimalType(),
				pginit.WithUUIDType(),
			)
			if err != nil {
				t.Error("expected no error")
			}

			db, err := pgi.ConnPool(ctx)
			if err != nil {
				t.Error("expected no error")
			}

			if err := db.Ping(ctx); err != nil {
				t.Error("expected no error")
			}

			if db.Config().ConnConfig.Logger == nil {
				t.Error("expected logger not nil")
			}

			if db.Config().ConnConfig.LogLevel != tt.wantedLvl {
				t.Errorf("expected log level %d got %d", tt.wantedLvl, db.Config().ConnConfig.LogLevel)
			}

			ctx = context.WithValue(ctx, "request-id", "12345") // nolint: revive, staticcheck, nolintlint
			if _, err := db.Exec(ctx, "SELECT * FROM ERROR"); err == nil {
				t.Error("expected return error")
			}
		})
	}
}

func TestConnPool_WithCustomDataTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		opts             []pginit.Option
		expectErrDecimal bool
		expectErrUUID    bool
	}{
		{
			name: "decimal + uuid",
			opts: []pginit.Option{
				pginit.WithLogLevel(zerolog.DebugLevel),
				pginit.WithLogger(&zerolog.Logger{}, "request-id"),
				pginit.WithDecimalType(),
				pginit.WithUUIDType(),
			},
			expectErrDecimal: false,
			expectErrUUID:    false,
		},
		{
			name: "uuid + decimal",
			opts: []pginit.Option{
				pginit.WithLogLevel(zerolog.DebugLevel),
				pginit.WithLogger(&zerolog.Logger{}, "request-id"),
				pginit.WithUUIDType(),
				pginit.WithDecimalType(),
			},
			expectErrDecimal: false,
			expectErrUUID:    false,
		},
		{
			name: "decimal",
			opts: []pginit.Option{
				pginit.WithLogLevel(zerolog.DebugLevel),
				pginit.WithLogger(&zerolog.Logger{}, "request-id"),
				pginit.WithDecimalType(),
			},
			expectErrDecimal: false,
			expectErrUUID:    true,
		},
		{
			name: "uuid",
			opts: []pginit.Option{
				pginit.WithLogLevel(zerolog.DebugLevel),
				pginit.WithLogger(&zerolog.Logger{}, "request-id"),
				pginit.WithUUIDType(),
			},
			expectErrDecimal: true,
			expectErrUUID:    false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			pgi, err := pginit.New(
				&pginit.Config{
					Host:     testHost,
					Port:     testPort,
					User:     "postgres",
					Password: "postgres",
					Database: "datawarehouse",
					MaxConns: 2,
				},
				tt.opts...,
			)
			if err != nil {
				t.Error("expected no error")
			}

			db, err := pgi.ConnPool(ctx)
			if err != nil {
				t.Error("expected no error")
			}

			err = db.Ping(ctx)
			if err != nil {
				t.Error("expected no error")
			}

			d := &decimal.Big{}
			err = db.QueryRow(context.Background(), "select 10.98").Scan(d)
			if err != nil && !tt.expectErrDecimal {
				t.Errorf("expected no err: %s", err)
			}

			u := &uuid.UUID{}
			err = db.QueryRow(context.Background(), "select 'b7202eb0-5bf0-475d-8ee2-d3d2c168a5d5'").Scan(u)
			if err != nil && !tt.expectErrUUID {
				t.Errorf("expected no err: %s", err)
			}
		})
	}
}

func TestConnPoolWithCustomTypes_CRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name string
	}{
		{
			name: "CRUD operation with custom type uuid and decimal",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pgi, err := pginit.New(&pginit.Config{
				Host:     testHost,
				Port:     testPort,
				User:     "postgres",
				Password: "postgres",
				Database: "datawarehouse",
			},
				pginit.WithLogger(&zerolog.Logger{}, "request-id"),
				pginit.WithDecimalType(),
				pginit.WithUUIDType(),
			)
			if err != nil {
				t.Errorf("expected no error but got: %v", err)
			}

			pool, err := pgi.ConnPool(ctx)
			if err != nil {
				t.Errorf("expected no error but got: %v", err)
			}

			conn, err := pool.Acquire(ctx)
			if err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
			defer conn.Release()

			tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
			defer tx.Rollback(ctx)

			_, err = tx.Exec(ctx, "CREATE TABLE IF NOT EXISTS uuid_decimal(uuid uuid, price numeric, PRIMARY KEY (uuid))")
			if err != nil {
				t.Errorf("expected no error but got: %v", err)
			}

			// create
			row := tx.QueryRow(ctx, "INSERT INTO uuid_decimal(uuid, price) VALUES('b7202eb0-5bf0-475d-8ee2-d3d2c168a5d5', 10.988888888889) RETURNING uuid, price")
			r := struct {
				uuid  uuid.UUID
				price decimal.Big
			}{}
			if err := row.Scan(&r.uuid, &r.price); err != nil { // nolint: govet // inline err is within scope
				t.Errorf("expected no error but got: %v, (%+v)", err, row)
			}
			if r.uuid.String() != "b7202eb0-5bf0-475d-8ee2-d3d2c168a5d5" || r.price.Cmp(decimal.New(10988888888889, 12)) != 0 {
				t.Error("inserted data doesn't match with input")
			}

			// read
			rows, err := tx.Query(ctx, "SELECT * FROM uuid_decimal")
			if err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
			defer rows.Close()
			var results []struct {
				uuid  uuid.UUID
				price decimal.Big
			}
			for rows.Next() {
				r := struct { // nolint: govet // r is within loop scope
					uuid  uuid.UUID
					price decimal.Big
				}{}
				if err := rows.Scan(&r.uuid, &r.price); err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
				if r.uuid.String() != "b7202eb0-5bf0-475d-8ee2-d3d2c168a5d5" || r.price.Cmp(decimal.New(10988888888889, 12)) != 0 {
					t.Error("inserted data doesn't match with input")
				}
				results = append(results, r)
			}
			if len(results) != 1 {
				t.Errorf("expected 1 result but got: %v", len(results))
			}
			// update
			row = tx.QueryRow(ctx, "UPDATE uuid_decimal SET price = 11.00 WHERE uuid = $1 RETURNING uuid, price", "b7202eb0-5bf0-475d-8ee2-d3d2c168a5d5")
			if err := row.Scan(&r.uuid, &r.price); err != nil {
				t.Errorf("expected no error but got: %v, (%+v)", err, row)
			}
			if r.price.Cmp(decimal.New(1100, 2)) != 0 {
				t.Errorf("expected 11.00 but got %+v", r)
			}

			// delete
			row = tx.QueryRow(ctx, "DELETE FROM uuid_decimal WHERE uuid = $1 RETURNING uuid, price", "b7202eb0-5bf0-475d-8ee2-d3d2c168a5d5")
			if err := row.Scan(&r.uuid, &r.price); err != nil {
				t.Errorf("expected no error but got: %v, (%+v)", err, row)
			}
			if r.uuid.String() != "b7202eb0-5bf0-475d-8ee2-d3d2c168a5d5" {
				t.Error("inserted data doesn't match with input")
			}
			row = tx.QueryRow(ctx, "SELECT * FROM uuid_decimal WHERE uuid = $1", "b7202eb0-5bf0-475d-8ee2-d3d2c168a5d5")
			if err := row.Scan(&r.uuid, &r.price); err != nil && !errors.Is(err, pgx.ErrNoRows) {
				t.Errorf("expected no error but got: %v, (%+v)", err, row)
			}
		})
	}
}

func BenchmarkConnPool(b *testing.B) {
	for i := 0; i <= b.N; i++ {
		ctx := context.Background()

		b.StartTimer()

		pgi, _ := pginit.New(
			&pginit.Config{
				Host:     testHost,
				Port:     testPort,
				User:     "postgres",
				Password: "postgres",
				Database: "datawarehouse",
			},
			pginit.WithLogger(&zerolog.Logger{}, "request-id"),
			pginit.WithDecimalType(),
			pginit.WithUUIDType(),
		)

		pgi.ConnPool(ctx)

		b.StopTimer()
	}
}
