package pginit_test

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/monacohq/golang-common/database/pginit"
	"github.com/rs/zerolog"
)

func Example_connPool() {
	pgi, err := pginit.New(&pginit.Config{
		Host:         "localhost",
		Port:         "5432",
		User:         "postgres",
		Password:     "postgres",
		Database:     "datawarehouse",
		MaxConns:     10,
		MaxIdleConns: 10,
		MaxLifeTime:  1 * time.Minute,
	})
	if err != nil {
		log.Fatalf("init pgi config: %v", err)
	}

	ctx := context.Background()

	pool, err := pgi.ConnPool(ctx)
	if err != nil {
		log.Fatalf("init pgi config: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping: %v", err)
	}
}

func Example_connPoolWithLogger() {
	logger := zerolog.New(os.Stderr)

	pgi, err := pginit.New(
		&pginit.Config{
			Host:         "localhost",
			Port:         "5432",
			User:         "postgres",
			Password:     "postgres",
			Database:     "datawarehouse",
			MaxConns:     10,
			MaxIdleConns: 10,
			MaxLifeTime:  1 * time.Minute,
		},
		pginit.WithLogLevel(zerolog.WarnLevel),
		pginit.WithLogger(&logger, "request-id"),
		pginit.WithDecimalType(),
		pginit.WithUUIDType(),
	)
	if err != nil {
		log.Fatalf("init pgi config: %v", err)
	}

	ctx := context.Background()

	pool, err := pgi.ConnPool(ctx)
	if err != nil {
		log.Fatalf("init pgi config: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping: %v", err)
	}
}
