package pginit

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v4/stdlib"
)

// StdConn returns a std lib *sql.DB.
func (pgi *PGInit) StdConn(ctx context.Context) (*sql.DB, error) {
	connStr := stdlib.RegisterConnConfig(pgi.pgxConf.ConnConfig)

	dbConn, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("open stdlib connection with config: %w", err)
	}

	dbConn.SetConnMaxIdleTime(pgi.pgxConf.MaxConnIdleTime)
	dbConn.SetConnMaxLifetime(pgi.pgxConf.MaxConnLifetime)
	dbConn.SetMaxOpenConns(int(pgi.pgxConf.MaxConns))
	dbConn.SetMaxIdleConns(int(pgi.pgxConf.MinConns))

	if err := dbConn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping std connection with config: %w", err)
	}

	return dbConn, nil
}
