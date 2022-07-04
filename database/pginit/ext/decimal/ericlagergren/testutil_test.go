package ericlagergren_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/lib/pq"
)

func MustConnectDatabaseSQL(tb testing.TB, driverName string) *sql.DB {
	tb.Helper()

	var sqlDriverName string

	switch driverName {
	case "github.com/lib/pq":
		sqlDriverName = "postgres"
	case "github.com/jackc/pgx/stdlib":
		sqlDriverName = "pgx"
	default:
		tb.Fatalf("Unknown driver %v", driverName)
	}

	db, err := sql.Open(sqlDriverName, os.Getenv("PGX_TEST_DATABASE"))
	if err != nil {
		tb.Fatal(err)
	}

	return db
}

func MustConnectPgx(tb testing.TB) *pgx.Conn {
	tb.Helper()

	conn, err := pgx.Connect(context.Background(), os.Getenv("PGX_TEST_DATABASE"))
	if err != nil {
		tb.Fatal(err)
	}

	return conn
}

func MustClose(tb testing.TB, conn interface {
	Close() error
},
) {
	tb.Helper()

	if err := conn.Close(); err != nil {
		tb.Fatal(err)
	}
}

func MustCloseContext(tb testing.TB, conn interface {
	Close(context.Context) error
},
) {
	tb.Helper()

	if err := conn.Close(context.Background()); err != nil {
		tb.Fatal(err)
	}
}

type forceTextEncoder struct {
	e pgtype.TextEncoder
}

func (f forceTextEncoder) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	b, err := f.e.EncodeText(ci, buf)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return b, nil
}

type forceBinaryEncoder struct {
	e pgtype.BinaryEncoder
}

func (f forceBinaryEncoder) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	b, err := f.e.EncodeBinary(ci, buf)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return b, nil
}

func ForceEncoder(encoder interface{}, formatCode int16) interface{} {
	switch formatCode {
	case pgx.TextFormatCode:
		if e, ok := encoder.(pgtype.TextEncoder); ok {
			return forceTextEncoder{e: e}
		}
	case pgx.BinaryFormatCode:
		if e, ok := encoder.(pgtype.BinaryEncoder); ok {
			return forceBinaryEncoder{e: e}
		}
	}

	return nil
}

func SuccessfulTranscode(tb testing.TB, pgTypeName string, values []interface{}) {
	tb.Helper()
	SuccessfulTranscodeEqFunc(tb, pgTypeName, values, reflect.DeepEqual)
}

func SuccessfulTranscodeEqFunc(
	tb testing.TB,
	pgTypeName string,
	values []interface{},
	eqFunc func(a, b interface{}) bool,
) {
	tb.Helper()

	PgxSuccessfulTranscodeEqFunc(tb, pgTypeName, values, eqFunc)

	for _, driverName := range []string{"github.com/lib/pq", "github.com/jackc/pgx/stdlib"} {
		DatabaseSQLSuccessfulTranscodeEqFunc(tb, driverName, pgTypeName, values, eqFunc)
	}
}

func PgxSuccessfulTranscodeEqFunc(tb testing.TB, pgTypeName string, values []interface{}, eqFunc func(a, b interface{}) bool) {
	tb.Helper()

	conn := MustConnectPgx(tb)
	defer MustCloseContext(tb, conn)

	_, err := conn.Prepare(context.Background(), "test", fmt.Sprintf("select $1::%s", pgTypeName))
	if err != nil {
		tb.Fatal(err)
	}

	formats := []struct {
		name       string
		formatCode int16
	}{
		{name: "TextFormat", formatCode: pgx.TextFormatCode},
		{name: "BinaryFormat", formatCode: pgx.BinaryFormatCode},
	}

	for index, val := range values {
		for _, paramFormat := range formats {
			for _, resultFormat := range formats {
				vEncoder := ForceEncoder(val, paramFormat.formatCode)
				if vEncoder == nil {
					tb.Logf(
						"Skipping Param %s Result %s: %#v does not implement %v for encoding",
						paramFormat.name, resultFormat.name, val, paramFormat.name,
					)

					continue
				}

				switch resultFormat.formatCode {
				case pgx.TextFormatCode:
					if _, ok := val.(pgtype.TextEncoder); !ok {
						tb.Logf(
							"Skipping Param %s Result %s: %#v does not implement %v for decoding",
							paramFormat.name, resultFormat.name, val, resultFormat.name,
						)

						continue
					}
				case pgx.BinaryFormatCode:
					if _, ok := val.(pgtype.BinaryEncoder); !ok {
						tb.Logf(
							"Skipping Param %s Result %s: %#v does not implement %v for decoding",
							paramFormat.name, resultFormat.name, val, resultFormat.name,
						)

						continue
					}
				}

				// Derefence value if it is a pointer
				derefV := val

				refVal := reflect.ValueOf(val)
				if refVal.Kind() == reflect.Ptr {
					derefV = refVal.Elem().Interface()
				}

				result := reflect.New(reflect.TypeOf(derefV))

				if err := conn.QueryRow(
					context.Background(),
					"test",
					pgx.QueryResultFormats{resultFormat.formatCode},
					vEncoder,
				).Scan(result.Interface()); err != nil {
					tb.Errorf("Param %s Result %s %d: %v", paramFormat.name, resultFormat.name, index, err)
				}

				if !eqFunc(result.Elem().Interface(), derefV) {
					tb.Errorf(
						"Param %s Result %s %d: expected %v, got %v",
						paramFormat.name, resultFormat.name, index, derefV, result.Elem().Interface(),
					)
				}
			}
		}
	}
}

func DatabaseSQLSuccessfulTranscodeEqFunc(
	tb testing.TB,
	driverName,
	pgTypeName string,
	values []interface{},
	eqFunc func(a, b interface{}) bool,
) {
	tb.Helper()

	conn := MustConnectDatabaseSQL(tb, driverName)
	defer MustClose(tb, conn)

	preStmt, err := conn.Prepare(fmt.Sprintf("select $1::%s", pgTypeName))
	if err != nil {
		tb.Fatal(err)
	}

	for index, val := range values {
		// Derefence value if it is a pointer
		derefV := val

		refVal := reflect.ValueOf(val)
		if refVal.Kind() == reflect.Ptr {
			derefV = refVal.Elem().Interface()
		}

		result := reflect.New(reflect.TypeOf(derefV))
		if err := preStmt.QueryRow(val).Scan(result.Interface()); err != nil {
			tb.Errorf("%v %d: %v", driverName, index, err)
		}

		if !eqFunc(result.Elem().Interface(), derefV) {
			tb.Errorf("%v %d: expected %v, got %v", driverName, index, derefV, result.Elem().Interface())
		}
	}
}

type NormalizeTest struct {
	SQL   string
	Value interface{}
}

func SuccessfulNormalize(tb testing.TB, tests []NormalizeTest) {
	tb.Helper()
	SuccessfulNormalizeEqFunc(tb, tests, reflect.DeepEqual)
}

func SuccessfulNormalizeEqFunc(tb testing.TB, tests []NormalizeTest, eqFunc func(a, b interface{}) bool) {
	tb.Helper()

	PgxSuccessfulNormalizeEqFunc(tb, tests, eqFunc)

	for _, driverName := range []string{"github.com/lib/pq", "github.com/jackc/pgx/stdlib"} {
		DatabaseSQLSuccessfulNormalizeEqFunc(tb, driverName, tests, eqFunc)
	}
}

func PgxSuccessfulNormalizeEqFunc(tb testing.TB, tests []NormalizeTest, eqFunc func(a, b interface{}) bool) {
	tb.Helper()

	conn := MustConnectPgx(tb)
	defer MustCloseContext(tb, conn)

	formats := []struct {
		name       string
		formatCode int16
	}{
		{name: "TextFormat", formatCode: pgx.TextFormatCode},
		{name: "BinaryFormat", formatCode: pgx.BinaryFormatCode},
	}

	for index, test := range tests {
		for _, format := range formats {
			psName := fmt.Sprintf("test%d", index)
			if _, err := conn.Prepare(context.Background(), psName, test.SQL); err != nil {
				tb.Fatal(err)
			}

			queryResultFormats := pgx.QueryResultFormats{format.formatCode}

			if ForceEncoder(test.Value, format.formatCode) == nil {
				tb.Logf("Skipping: %#v does not implement %v", test.Value, format.name)

				continue
			}
			// Derefence value if it is a pointer
			derefV := test.Value

			refVal := reflect.ValueOf(test.Value)
			if refVal.Kind() == reflect.Ptr {
				derefV = refVal.Elem().Interface()
			}

			result := reflect.New(reflect.TypeOf(derefV))
			if err := conn.QueryRow(context.Background(), psName, queryResultFormats).Scan(result.Interface()); err != nil {
				tb.Errorf("%v %d: %v", format.name, index, err)
			}

			if !eqFunc(result.Elem().Interface(), derefV) {
				tb.Errorf("%v %d: expected %v, got %v", format.name, index, derefV, result.Elem().Interface())
			}
		}
	}
}

func DatabaseSQLSuccessfulNormalizeEqFunc(
	tb testing.TB,
	driverName string,
	tests []NormalizeTest,
	eqFunc func(a, b interface{}) bool,
) {
	tb.Helper()

	conn := MustConnectDatabaseSQL(tb, driverName)
	defer MustClose(tb, conn)

	for index, test := range tests {
		preStmt, err := conn.Prepare(test.SQL)
		if err != nil {
			tb.Errorf("%d. %v", index, err)

			continue
		}

		// Derefence value if it is a pointer
		derefV := test.Value

		refVal := reflect.ValueOf(test.Value)
		if refVal.Kind() == reflect.Ptr {
			derefV = refVal.Elem().Interface()
		}

		result := reflect.New(reflect.TypeOf(derefV))
		if err := preStmt.QueryRow().Scan(result.Interface()); err != nil {
			tb.Errorf("%v %d: %v", driverName, index, err)
		}

		if !eqFunc(result.Elem().Interface(), derefV) {
			tb.Errorf("%v %d: expected %v, got %v", driverName, index, derefV, result.Elem().Interface())
		}
	}
}

func GoZeroToNullConversion(tb testing.TB, pgTypeName string, zero interface{}) {
	tb.Helper()

	PgxGoZeroToNullConversion(tb, pgTypeName, zero)

	for _, driverName := range []string{"github.com/lib/pq", "github.com/jackc/pgx/stdlib"} {
		DatabaseSQLGoZeroToNullConversion(tb, driverName, pgTypeName, zero)
	}
}

func NullToGoZeroConversion(tb testing.TB, pgTypeName string, zero interface{}) {
	tb.Helper()

	PgxNullToGoZeroConversion(tb, pgTypeName, zero)

	for _, driverName := range []string{"github.com/lib/pq", "github.com/jackc/pgx/stdlib"} {
		DatabaseSQLNullToGoZeroConversion(tb, driverName, pgTypeName, zero)
	}
}

func PgxGoZeroToNullConversion(tb testing.TB, pgTypeName string, zero interface{}) {
	tb.Helper()

	conn := MustConnectPgx(tb)
	defer MustCloseContext(tb, conn)

	_, err := conn.Prepare(context.Background(), "test", fmt.Sprintf("select $1::%s is null", pgTypeName))
	if err != nil {
		tb.Fatal(err)
	}

	formats := []struct {
		name       string
		formatCode int16
	}{
		{name: "TextFormat", formatCode: pgx.TextFormatCode},
		{name: "BinaryFormat", formatCode: pgx.BinaryFormatCode},
	}

	for _, paramFormat := range formats {
		vEncoder := ForceEncoder(zero, paramFormat.formatCode)
		if vEncoder == nil {
			tb.Logf("Skipping Param %s: %#v does not implement %v for encoding", paramFormat.name, zero, paramFormat.name)

			continue
		}

		var result bool
		if err := conn.QueryRow(context.Background(), "test", vEncoder).Scan(&result); err != nil {
			tb.Errorf("Param %s: %v", paramFormat.name, err)
		}

		if !result {
			tb.Errorf("Param %s: did not convert zero to null", paramFormat.name)
		}
	}
}

func PgxNullToGoZeroConversion(tb testing.TB, pgTypeName string, zero interface{}) {
	tb.Helper()

	conn := MustConnectPgx(tb)
	defer MustCloseContext(tb, conn)

	_, err := conn.Prepare(context.Background(), "test", fmt.Sprintf("select null::%s", pgTypeName))
	if err != nil {
		tb.Fatal(err)
	}

	formats := []struct {
		name       string
		formatCode int16
	}{
		{name: "TextFormat", formatCode: pgx.TextFormatCode},
		{name: "BinaryFormat", formatCode: pgx.BinaryFormatCode},
	}

	for _, resultFormat := range formats {
		switch resultFormat.formatCode {
		case pgx.TextFormatCode:
			if _, ok := zero.(pgtype.TextEncoder); !ok {
				tb.Logf("Skipping Result %s: %#v does not implement %v for decoding", resultFormat.name, zero, resultFormat.name)

				continue
			}
		case pgx.BinaryFormatCode:
			if _, ok := zero.(pgtype.BinaryEncoder); !ok {
				tb.Logf("Skipping Result %s: %#v does not implement %v for decoding", resultFormat.name, zero, resultFormat.name)

				continue
			}
		}

		// Derefence value if it is a pointer
		derefZero := zero

		refVal := reflect.ValueOf(zero)
		if refVal.Kind() == reflect.Ptr {
			derefZero = refVal.Elem().Interface()
		}

		result := reflect.New(reflect.TypeOf(derefZero))

		if err := conn.QueryRow(context.Background(), "test").Scan(result.Interface()); err != nil {
			tb.Errorf("Result %s: %v", resultFormat.name, err)
		}

		if !reflect.DeepEqual(result.Elem().Interface(), derefZero) {
			tb.Errorf("Result %s: did not convert null to zero", resultFormat.name)
		}
	}
}

func DatabaseSQLGoZeroToNullConversion(tb testing.TB, driverName, pgTypeName string, zero interface{}) {
	tb.Helper()

	conn := MustConnectDatabaseSQL(tb, driverName)
	defer MustClose(tb, conn)

	preStmt, err := conn.Prepare(fmt.Sprintf("select $1::%s is null", pgTypeName))
	if err != nil {
		tb.Fatal(err)
	}

	var result bool
	if err := preStmt.QueryRow(zero).Scan(&result); err != nil {
		tb.Errorf("%v %v", driverName, err)
	}

	if !result {
		tb.Errorf("%v: did not convert zero to null", driverName)
	}
}

func DatabaseSQLNullToGoZeroConversion(tb testing.TB, driverName, pgTypeName string, zero interface{}) {
	tb.Helper()

	conn := MustConnectDatabaseSQL(tb, driverName)
	defer MustClose(tb, conn)

	preStmt, err := conn.Prepare(fmt.Sprintf("select null::%s", pgTypeName))
	if err != nil {
		tb.Fatal(err)
	}

	// Derefence value if it is a pointer
	derefZero := zero

	refVal := reflect.ValueOf(zero)
	if refVal.Kind() == reflect.Ptr {
		derefZero = refVal.Elem().Interface()
	}

	result := reflect.New(reflect.TypeOf(derefZero))

	if err := preStmt.QueryRow().Scan(result.Interface()); err != nil {
		tb.Errorf("%v %v", driverName, err)
	}

	if !reflect.DeepEqual(result.Elem().Interface(), derefZero) {
		tb.Errorf("%s: did not convert null to zero", driverName)
	}
}
