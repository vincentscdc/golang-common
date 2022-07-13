package pginit

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgtype"
	gofrs "github.com/jackc/pgtype/ext/gofrs-uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zerologadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/monacohq/golang-common/database/pginit/ext/decimal/ericlagergren"
	"github.com/rs/zerolog"
)

const (
	defaultMaxConns     = 25
	defaultMaxIdleConns = 25
	defaultMaxLifeTime  = 5 * time.Minute
)

// Option configures PGInit behaviour.
type Option func(*PGInit)

// PGInit provides capabilities for connect to postgres with pgx.pool.
type PGInit struct {
	pgxConf         *pgxpool.Config
	logLvl          pgx.LogLevel
	customDataTypes []pgtype.DataType
}

// New initializes a PGInit using the provided Config and options. If
// opts is not provided it will initializes PGInit with default configuration.
func New(conf *Config, opts ...Option) (*PGInit, error) {
	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		conf.User, conf.Password, net.JoinHostPort(conf.Host, conf.Port), conf.Database,
	)

	pgxConf, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	pgxConf.MaxConns = defaultMaxConns
	if conf.MaxConns != 0 {
		pgxConf.MaxConns = conf.MaxConns
	}

	pgxConf.MinConns = defaultMaxIdleConns
	if conf.MaxIdleConns != 0 && conf.MaxConns >= conf.MaxIdleConns {
		pgxConf.MinConns = conf.MaxIdleConns
	} else {
		pgxConf.MinConns = pgxConf.MaxConns
	}

	pgxConf.MaxConnLifetime = defaultMaxLifeTime
	if conf.MaxLifeTime != 0 {
		pgxConf.MaxConnLifetime = conf.MaxLifeTime
	}

	pgi := &PGInit{
		pgxConf: pgxConf,
		logLvl:  pgx.LogLevelWarn,
	}

	for _, opt := range opts {
		opt(pgi)
	}

	pgi.pgxConf.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		for _, dataType := range pgi.customDataTypes {
			c.ConnInfo().RegisterDataType(dataType)
		}

		return nil
	}

	return pgi, nil
}

// ConnPool initiates connection to database and return a pgxpool.Pool.
func (pgi *PGInit) ConnPool(ctx context.Context) (*pgxpool.Pool, error) {
	pool, err := pgxpool.ConnectConfig(ctx, pgi.pgxConf)
	if err != nil {
		return nil, fmt.Errorf("connect config: %w", err)
	}

	return pool, nil
}

// WithLogger Add logger to pgx. if the request context contains request id,
// can pass in the request id context key to reqIDKeyFromCtx and logger will
// log with the request id. Only will log if the log level is equal and above pgx.LogLevelWarn.
func WithLogger(logger *zerolog.Logger, reqIDKeyFromCtx string) Option {
	return func(pgi *PGInit) {
		pgi.pgxConf.ConnConfig.LogLevel = pgi.logLvl
		pgi.pgxConf.ConnConfig.Logger = zerologadapter.NewLogger(*logger, zerologadapter.WithContextFunc(
			func(ctx context.Context, logWith zerolog.Context) zerolog.Context {
				if ctxValue, ok := ctx.Value(reqIDKeyFromCtx).(string); ok {
					logWith = logWith.Str(reqIDKeyFromCtx, ctxValue)
				}

				return logWith
			},
		))
	}
}

// WithLogLevel set pgx log level.
func WithLogLevel(zLvl zerolog.Level) Option {
	return func(pgi *PGInit) {
		switch {
		case zLvl == zerolog.DebugLevel:
			pgi.logLvl = pgx.LogLevelDebug
		case zLvl == zerolog.InfoLevel:
			pgi.logLvl = pgx.LogLevelInfo
		case zLvl == zerolog.WarnLevel:
			pgi.logLvl = pgx.LogLevelWarn
		case zLvl == zerolog.ErrorLevel:
			pgi.logLvl = pgx.LogLevelError
		case zLvl == zerolog.NoLevel:
			pgi.logLvl = pgx.LogLevelNone
		}

		pgi.pgxConf.ConnConfig.LogLevel = pgi.logLvl
	}
}

func WithDecimalType() Option {
	return func(p *PGInit) {
		p.customDataTypes = append(p.customDataTypes, pgtype.DataType{
			Value: &ericlagergren.Numeric{},
			Name:  "numeric",
			OID:   pgtype.NumericOID,
		})
	}
}

func WithUUIDType() Option {
	return func(p *PGInit) {
		p.customDataTypes = append(p.customDataTypes, pgtype.DataType{
			Value: &gofrs.UUID{},
			Name:  "uuid",
			OID:   pgtype.UUIDOID,
		})
	}
}
