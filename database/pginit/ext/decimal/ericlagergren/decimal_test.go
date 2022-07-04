package ericlagergren_test

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/monacohq/golang-common/database/pginit/ext/decimal/ericlagergren"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.uber.org/goleak"
)

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
		os.Setenv("PGX_TEST_DATABASE", databaseURL)

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
	var testHost, testPort string

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

	return testHost + ":" + testPort
}

func mustParseDecimal(t *testing.T, src string) decimal.Big {
	t.Helper()

	dec, ok := new(decimal.Big).SetString(src)
	if !ok {
		t.Fatalf("cannot set %v to decimal", src)
	}

	return *dec
}

func TestNumericNormalize(t *testing.T) {
	t.Parallel()

	SuccessfulNormalizeEqFunc(t, []NormalizeTest{
		{
			SQL:   "select '0'::numeric",
			Value: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "0"), Status: pgtype.Present},
		},
		{
			SQL:   "select '1'::numeric",
			Value: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present},
		},
		{
			SQL:   "select '10.00'::numeric",
			Value: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "10.00"), Status: pgtype.Present},
		},
		{
			SQL:   "select '1e-3'::numeric",
			Value: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "0.001"), Status: pgtype.Present},
		},
		{
			SQL:   "select '-1'::numeric",
			Value: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "-1"), Status: pgtype.Present},
		},
		{
			SQL:   "select '10000'::numeric",
			Value: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "10000"), Status: pgtype.Present},
		},
		{
			SQL:   "select '3.14'::numeric",
			Value: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "3.14"), Status: pgtype.Present},
		},
		{
			SQL:   "select '1.1'::numeric",
			Value: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1.1"), Status: pgtype.Present},
		},
		{
			SQL:   "select '100010001'::numeric",
			Value: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "100010001"), Status: pgtype.Present},
		},
		{
			SQL:   "select '100010001.0001'::numeric",
			Value: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "100010001.0001"), Status: pgtype.Present},
		},
		{
			SQL: "select '4237234789234789289347892374324872138321894178943189043890124832108934.43219085471578891547854892438945012347981'::numeric",
			Value: &ericlagergren.Numeric{
				Decimal: mustParseDecimal(t, "4237234789234789289347892374324872138321894178943189043890124832108934.43219085471578891547854892438945012347981"),
				Status:  pgtype.Present,
			},
		},
		{
			SQL: "select '0.8925092023480223478923478978978937897879595901237890234789243679037419057877231734823098432903527585734549035904590854890345905434578345789347890402348952348905890489054234237489234987723894789234'::numeric",
			Value: &ericlagergren.Numeric{
				Decimal: mustParseDecimal(t, "0.8925092023480223478923478978978937897879595901237890234789243679037419057877231734823098432903527585734549035904590854890345905434578345789347890402348952348905890489054234237489234987723894789234"),
				Status:  pgtype.Present,
			},
		},
		{
			SQL: "select '0.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000123'::numeric",
			Value: &ericlagergren.Numeric{
				Decimal: mustParseDecimal(t, "0.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000123"),
				Status:  pgtype.Present,
			},
		},
	}, func(aa, bb interface{}) bool {
		a, ok := aa.(ericlagergren.Numeric)
		if !ok {
			t.Errorf("cannot convert %T", aa)
		}
		b, ok := bb.(ericlagergren.Numeric)
		if !ok {
			t.Errorf("cannot convert %T", bb)
		}

		equal := false
		if res := a.Decimal.Cmp(&b.Decimal); res == 0 {
			equal = true
		}

		return a.Status == b.Status && equal
	})
}

func TestNumericTranscode(t *testing.T) {
	t.Parallel()

	SuccessfulTranscodeEqFunc(t, "numeric", []interface{}{
		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "0"), Status: pgtype.Present},
		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present},
		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "-1"), Status: pgtype.Present},
		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "100000"), Status: pgtype.Present},

		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "0.1"), Status: pgtype.Present},
		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "0.01"), Status: pgtype.Present},
		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "0.001"), Status: pgtype.Present},
		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "0.0001"), Status: pgtype.Present},
		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "0.00001"), Status: pgtype.Present},
		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "0.000001"), Status: pgtype.Present},

		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "3.14"), Status: pgtype.Present},
		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "0.00000123"), Status: pgtype.Present},
		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "0.000000123"), Status: pgtype.Present},
		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "0.0000000123"), Status: pgtype.Present},
		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "0.00000000123"), Status: pgtype.Present},
		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "0.00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001234567890123456789"), Status: pgtype.Present},
		&ericlagergren.Numeric{Decimal: mustParseDecimal(t, "4309132809320932980457137401234890237489238912983572189348951289375283573984571892758234678903467889512893489128589347891272139.8489235871258912789347891235879148795891238915678189467128957812395781238579189025891238901583915890128973578957912385798125789012378905238905471598123758923478294374327894237892234"), Status: pgtype.Present},
		&ericlagergren.Numeric{Status: pgtype.Null},
	}, func(aa, bb interface{}) bool {
		a, ok := aa.(ericlagergren.Numeric)
		if !ok {
			t.Errorf("cannot convert %T", aa)
		}
		b, ok := bb.(ericlagergren.Numeric)
		if !ok {
			t.Errorf("cannot convert %T", bb)
		}

		equal := false
		if res := a.Decimal.Cmp(&b.Decimal); res == 0 {
			equal = true
		}

		return a.Status == b.Status && equal
	})
}

func TestNumericTranscodeFuzz(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewSource(0)) // nolint: gosec // good enough for unit test
	max := &big.Int{}
	max.SetString("9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 10)

	values := make([]interface{}, 0, 2000)

	for i := 0; i < 500; i++ {
		num := fmt.Sprintf("%s.%s", (&big.Int{}).Rand(r, max).String(), (&big.Int{}).Rand(r, max).String())
		negNum := "-" + num

		values = append(
			values,
			&ericlagergren.Numeric{Decimal: mustParseDecimal(t, num), Status: pgtype.Present},
			&ericlagergren.Numeric{Decimal: mustParseDecimal(t, negNum), Status: pgtype.Present},
		)
	}

	SuccessfulTranscodeEqFunc(t, "numeric", values,
		func(aa, bb interface{}) bool {
			a, ok := aa.(ericlagergren.Numeric)
			if !ok {
				t.Errorf("cannot convert %T", aa)
			}
			b, ok := bb.(ericlagergren.Numeric)
			if !ok {
				t.Errorf("cannot convert %T", bb)
			}

			equal := false
			if res := a.Decimal.Cmp(&b.Decimal); res == 0 {
				equal = true
			}

			return a.Status == b.Status && equal
		})
}

func TestNumericSet(t *testing.T) {
	t.Parallel()

	type _int8 int8

	successfulTests := []struct {
		source interface{}
		result *ericlagergren.Numeric
	}{
		{source: nil, result: nil},
		{source: *decimal.New(1, 0), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present}},
		// {source: decimal.NullDecimal{Valid: true, Decimal: decimal.New(1, 0)}, result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present}},
		// {source: decimal.NullDecimal{Valid: false}, result: &ericlagergren.Numeric{Status: pgtype.Null}},
		{source: float32(1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present}},
		{source: float64(1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present}},
		{source: int8(1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present}},
		{source: int16(1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present}},
		{source: int32(1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present}},
		{source: int64(1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present}},
		{source: int8(-1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "-1"), Status: pgtype.Present}},
		{source: int16(-1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "-1"), Status: pgtype.Present}},
		{source: int32(-1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "-1"), Status: pgtype.Present}},
		{source: int64(-1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "-1"), Status: pgtype.Present}},
		{source: uint8(1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present}},
		{source: uint16(1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present}},
		{source: uint32(1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present}},
		{source: uint64(1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present}},
		{source: int(-1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "-1"), Status: pgtype.Present}},
		{source: uint(1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present}},
		{source: "1", result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present}},
		{source: _int8(1), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1"), Status: pgtype.Present}},
		{source: float64(1000), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1000"), Status: pgtype.Present}},
		{source: float64(1234), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1234"), Status: pgtype.Present}},
		{source: float64(12345678900), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "12345678900"), Status: pgtype.Present}},
		{source: float64(1.25), result: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "1.25"), Status: pgtype.Present}},
	}

	for i, tt := range successfulTests {
		r := &ericlagergren.Numeric{}
		if err := r.Set(tt.source); err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if tt.result != nil && !(r.Status == tt.result.Status && r.Decimal.Cmp(&tt.result.Decimal) == 0) {
			t.Errorf("%d: expected %v to convert to %v, but it was %v", i, tt.source, tt.result, r)
		}
	}
}

func TestNumericAssignTo(t *testing.T) {
	t.Parallel()

	type _int8 int8

	var (
		i8   int8
		i16  int16
		i32  int32
		i64  int64
		i    int
		ui8  uint8
		ui16 uint16
		ui32 uint32
		ui64 uint64
		ui   uint
		pi8  *int8
		_i8  _int8
		_pi8 *_int8
		f32  float32
		f64  float64
		pf32 *float32
		pf64 *float64
		d    decimal.Big
	)

	simpleTests := []struct {
		src      *ericlagergren.Numeric
		dst      interface{}
		expected interface{}
	}{
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42"), Status: pgtype.Present}, dst: &f32, expected: float32(42)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42"), Status: pgtype.Present}, dst: &f64, expected: float64(42)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "4.2"), Status: pgtype.Present}, dst: &f32, expected: float32(4.2)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "4.2"), Status: pgtype.Present}, dst: &f64, expected: float64(4.2)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42"), Status: pgtype.Present}, dst: &i16, expected: int16(42)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42"), Status: pgtype.Present}, dst: &i32, expected: int32(42)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42"), Status: pgtype.Present}, dst: &i64, expected: int64(42)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42000"), Status: pgtype.Present}, dst: &i64, expected: int64(42000)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42"), Status: pgtype.Present}, dst: &i, expected: int(42)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42"), Status: pgtype.Present}, dst: &ui8, expected: uint8(42)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42"), Status: pgtype.Present}, dst: &ui16, expected: uint16(42)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42"), Status: pgtype.Present}, dst: &ui32, expected: uint32(42)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42"), Status: pgtype.Present}, dst: &ui64, expected: uint64(42)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42"), Status: pgtype.Present}, dst: &ui, expected: uint(42)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42"), Status: pgtype.Present}, dst: &_i8, expected: _int8(42)},
		{src: &ericlagergren.Numeric{Status: pgtype.Null}, dst: &pi8, expected: ((*int8)(nil))},
		{src: &ericlagergren.Numeric{Status: pgtype.Null}, dst: &_pi8, expected: ((*_int8)(nil))},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42"), Status: pgtype.Present}, dst: &d, expected: decimal.New(42, 0)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42000"), Status: pgtype.Present}, dst: &d, expected: decimal.New(42, -3)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "0.042"), Status: pgtype.Present}, dst: &d, expected: decimal.New(42, 3)},
		// {src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42"), Status: pgtype.Present}, dst: &nd, expected: decimal.NullDecimal{Valid: true, Decimal: decimal.New(42, 0)}},
		// {src: &ericlagergren.Numeric{Status: pgtype.Null}, dst: &nd, expected: decimal.NullDecimal{Valid: false}},
	}

	for i, tt := range simpleTests {
		// Zero out the destination variable
		reflect.ValueOf(tt.dst).Elem().Set(reflect.Zero(reflect.TypeOf(tt.dst).Elem()))

		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		// Need to specially handle Decimal or NullDecimal methods so we can use their Equal method. Without this
		// we end up checking reference equality on the *big.Int they contain.
		switch dst := tt.dst.(type) {
		case *decimal.Big:
			expected, ok := tt.expected.(*decimal.Big)
			if !ok {
				t.Errorf("cannot convert %T", tt.expected)
			}

			if dst.Cmp(expected) != 0 {
				t.Errorf("%d: expected %+v to assign %+v, but result was %+v", i, tt.src, tt.expected, d)
			}
		default:
			if dst := reflect.ValueOf(tt.dst).Elem().Interface(); dst != tt.expected {
				t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, dst)
			}
		}
	}

	pointerAllocTests := []struct {
		src      *ericlagergren.Numeric
		dst      interface{}
		expected interface{}
	}{
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42"), Status: pgtype.Present}, dst: &pf32, expected: float32(42)},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "42"), Status: pgtype.Present}, dst: &pf64, expected: float64(42)},
	}

	for i, tt := range pointerAllocTests {
		err := tt.src.AssignTo(tt.dst)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if dst := reflect.ValueOf(tt.dst).Elem().Elem().Interface(); dst != tt.expected {
			t.Errorf("%d: expected %v to assign %v, but result was %v", i, tt.src, tt.expected, dst)
		}
	}

	errorTests := []struct {
		src *ericlagergren.Numeric
		dst interface{}
	}{
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "150"), Status: pgtype.Present}, dst: &i8},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "40000"), Status: pgtype.Present}, dst: &i16},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "-1"), Status: pgtype.Present}, dst: &ui8},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "-1"), Status: pgtype.Present}, dst: &ui16},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "-1"), Status: pgtype.Present}, dst: &ui32},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "-1"), Status: pgtype.Present}, dst: &ui64},
		{src: &ericlagergren.Numeric{Decimal: mustParseDecimal(t, "-1"), Status: pgtype.Present}, dst: &ui},
		{src: &ericlagergren.Numeric{Status: pgtype.Null}, dst: &i32},
	}

	for i, tt := range errorTests {
		err := tt.src.AssignTo(tt.dst)
		if err == nil {
			t.Errorf("%d: expected error but none was returned (%v -> %v)", i, tt.src, tt.dst)
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	benchmarks := []struct {
		name      string
		numberStr string
	}{
		{"Zero", "0"},
		{"Small", "12345"},
		{"Medium", "12345.12345"},
		{"Large", "123457890.1234567890"},
		{"Huge", "123457890123457890123457890.1234567890123457890123457890"},
	}

	for _, bm := range benchmarks {
		src := &ericlagergren.Numeric{}
		if err := src.Set(bm.numberStr); err != nil {
			b.Errorf("expected no error but got %v", err)
		}

		textFormat, err := src.EncodeText(nil, nil)
		if err != nil {
			b.Errorf("expected no error but got %v", err)
		}

		binaryFormat, err := src.EncodeBinary(nil, nil)
		if err != nil {
			b.Errorf("expected no error but got %v", err)
		}

		b.Run(fmt.Sprintf("%s-Text", bm.name), func(b *testing.B) {
			dst := &ericlagergren.Numeric{}
			for i := 0; i < b.N; i++ {
				err := dst.DecodeText(nil, textFormat)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("%s-Binary", bm.name), func(b *testing.B) {
			dst := &ericlagergren.Numeric{}
			for i := 0; i < b.N; i++ {
				err := dst.DecodeBinary(nil, binaryFormat)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestNumericMarshalJSON(t *testing.T) {
	t.Parallel()

	simpleTests := []struct {
		src      *ericlagergren.Numeric
		expected []byte
		err      error
	}{
		{
			src: &ericlagergren.Numeric{
				Decimal: *decimal.New(1, 0),
				Status:  pgtype.Present,
			},
			expected: []byte("1"),
			err:      nil,
		},
		{
			src: &ericlagergren.Numeric{
				Decimal: *decimal.New(123, -3),
				Status:  pgtype.Present,
			},
			expected: []byte("1.23E+5"),
			err:      nil,
		},
		{
			src: &ericlagergren.Numeric{
				Decimal: *decimal.New(123, 9),
				Status:  pgtype.Present,
			},
			expected: []byte("1.23E-7"),
			err:      nil,
		},
		{
			src: &ericlagergren.Numeric{
				Decimal: *decimal.New(123, 5),
				Status:  pgtype.Present,
			},
			expected: []byte("0.00123"),
			err:      nil,
		},
		{
			src: &ericlagergren.Numeric{
				Decimal: *decimal.New(123, 5),
				Status:  pgtype.Null,
			},
			expected: []byte("null"),
			err:      nil,
		},
		{
			src: &ericlagergren.Numeric{
				Decimal: *decimal.New(123, 5),
				Status:  pgtype.Undefined,
			},
			expected: nil,
			err:      ericlagergren.ErrUndefined,
		},
		{
			src: &ericlagergren.Numeric{
				Decimal: *decimal.New(123, 5),
				Status:  5,
			},
			expected: nil,
			err:      ericlagergren.ErrBadStatus,
		},
	}

	for i, tt := range simpleTests {
		got, err := tt.src.MarshalJSON()
		if !errors.Is(err, tt.err) {
			t.Errorf("%d: expected error %v but got %v", i, tt.err, err)
		}

		if !bytes.Equal(got, tt.expected) {
			t.Errorf("%d: expected %s but got %s", i, tt.expected, got)
		}
	}
}

func TestNumericUnmarshalJSON(t *testing.T) {
	t.Parallel()

	simpleTests := []struct {
		dst      *ericlagergren.Numeric
		bytes    []byte
		expected *ericlagergren.Numeric
	}{
		{
			dst: &ericlagergren.Numeric{
				Decimal: *decimal.New(3, 0),
				Status:  pgtype.Present,
			},
			bytes: []byte(`2`),
			expected: &ericlagergren.Numeric{
				Decimal: *decimal.New(2, 0),
				Status:  pgtype.Present,
			},
		},
		{
			dst: &ericlagergren.Numeric{
				Decimal: *decimal.New(3, -2),
				Status:  pgtype.Present,
			},
			bytes: []byte(`200`),
			expected: &ericlagergren.Numeric{
				Decimal: *decimal.New(2, -2),
				Status:  pgtype.Present,
			},
		},
	}

	for i, tt := range simpleTests {
		if err := tt.dst.UnmarshalJSON(tt.bytes); err != nil {
			t.Errorf("%d: expected no error but got %v", i, err)
		}

		if tt.dst.Decimal.Cmp(&tt.expected.Decimal) != 0 {
			t.Errorf("%d: expected %+v but got %+v", i, tt.expected, tt.dst)
		}
	}
}

func TestNumericGet(t *testing.T) {
	t.Parallel()

	simpleTests := []struct {
		dst      *ericlagergren.Numeric
		expected interface{}
	}{
		{
			dst: &ericlagergren.Numeric{
				Decimal: *decimal.New(3, 0),
				Status:  pgtype.Present,
			},
			expected: *decimal.New(3, 0),
		},
		{
			dst: &ericlagergren.Numeric{
				Decimal: *decimal.New(3, 0),
				Status:  pgtype.Null,
			},
			expected: nil,
		},
		{
			dst: &ericlagergren.Numeric{
				Decimal: *decimal.New(3, 0),
				Status:  pgtype.Undefined,
			},
			expected: pgtype.Undefined,
		},
	}

	for i, tt := range simpleTests {
		got := tt.dst.Get()

		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("%d: expected %+v but got %+v", i, tt.expected, got)
		}
	}
}
