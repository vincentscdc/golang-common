<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# pginit

```go
import "github.com/monacohq/golang-common/database/pginit"
```

This package allows you to init a connection pool to postgres database via pgx below are default value in pginit:

default MaxConns = 25

default MaxIdleConns = 25

default MaxLifeTime = 5 minute

default LogLevel = Warn

<details><summary>Example (Conn Pool)</summary>
<p>

```go
package main

import (
	"context"
	"github.com/monacohq/golang-common/database/pginit"
	"log"
	"time"
)

func main() {
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
```

</p>
</details>

<details><summary>Example (Conn Pool With Logger)</summary>
<p>

```go
package main

import (
	"context"
	"github.com/monacohq/golang-common/database/pginit"
	"github.com/rs/zerolog"
	"log"
	"os"
	"time"
)

func main() {
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
```

</p>
</details>

## Index

- [type Config](<#type-config>)
- [type Option](<#type-option>)
  - [func WithDecimalType() Option](<#func-withdecimaltype>)
  - [func WithLogLevel(zLvl zerolog.Level) Option](<#func-withloglevel>)
  - [func WithLogger(logger *zerolog.Logger, reqIDKeyFromCtx string) Option](<#func-withlogger>)
  - [func WithUUIDType() Option](<#func-withuuidtype>)
- [type PGInit](<#type-pginit>)
  - [func New(conf *Config, opts ...Option) (*PGInit, error)](<#func-new>)
  - [func (pgi *PGInit) ConnPool(ctx context.Context) (*pgxpool.Pool, error)](<#func-pginit-connpool>)


## type [Config](<https://github.com/monacohq/golang-common/blob/main/database/pginit/config.go#L6-L15>)

Config allow you to set database credential to connect to database

```go
type Config struct {
    User         string
    Password     string
    Host         string
    Port         string
    Database     string
    MaxConns     int32
    MaxIdleConns int32
    MaxLifeTime  time.Duration
}
```

## type [Option](<https://github.com/monacohq/golang-common/blob/main/database/pginit/pool.go#L25>)

Option configures PGInit behaviour\.

```go
type Option func(*PGInit)
```

### func [WithDecimalType](<https://github.com/monacohq/golang-common/blob/main/database/pginit/pool.go#L133>)

```go
func WithDecimalType() Option
```

WithDecimalType set pgx decimal type to ericlagergren/decimal\.

### func [WithLogLevel](<https://github.com/monacohq/golang-common/blob/main/database/pginit/pool.go#L113>)

```go
func WithLogLevel(zLvl zerolog.Level) Option
```

WithLogLevel set pgx log level\.

### func [WithLogger](<https://github.com/monacohq/golang-common/blob/main/database/pginit/pool.go#L97>)

```go
func WithLogger(logger *zerolog.Logger, reqIDKeyFromCtx string) Option
```

WithLogger Add logger to pgx\. if the request context contains request id\, can pass in the request id context key to reqIDKeyFromCtx and logger will log with the request id\. Only will log if the log level is equal and above pgx\.LogLevelWarn\.

### func [WithUUIDType](<https://github.com/monacohq/golang-common/blob/main/database/pginit/pool.go#L144>)

```go
func WithUUIDType() Option
```

WithUUIDType set pgx uuid type to gofrs/uuid\.

## type [PGInit](<https://github.com/monacohq/golang-common/blob/main/database/pginit/pool.go#L28-L32>)

PGInit provides capabilities for connect to postgres with pgx\.pool\.

```go
type PGInit struct {
    // contains filtered or unexported fields
}
```

### func [New](<https://github.com/monacohq/golang-common/blob/main/database/pginit/pool.go#L36>)

```go
func New(conf *Config, opts ...Option) (*PGInit, error)
```

New initializes a PGInit using the provided Config and options\. If opts is not provided it will initializes PGInit with default configuration\.

### func \(\*PGInit\) [ConnPool](<https://github.com/monacohq/golang-common/blob/main/database/pginit/pool.go#L85>)

```go
func (pgi *PGInit) ConnPool(ctx context.Context) (*pgxpool.Pool, error)
```

ConnPool initiates connection to database and return a pgxpool\.Pool\.



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
