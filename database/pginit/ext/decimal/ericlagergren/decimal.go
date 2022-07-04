package ericlagergren

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"

	"github.com/ericlagergren/decimal"
	"github.com/jackc/pgtype"
)

var (
	ErrUndefined        = errors.New("cannot encode status undefined")
	ErrBadStatus        = errors.New("invalid status")
	errConversionFailed = errors.New("failed to convert")
	errAssignFailed     = errors.New("failed to assign")
	errScanFailed       = errors.New("failed to scan")
)

const (
	base      = 10
	bitSize8  = 8
	bitSize16 = 16
	bitSize32 = 32
	bitSize64 = 64
)

type Numeric struct {
	Decimal decimal.Big
	Status  pgtype.Status
}

// nolint: cyclop // need to check each type to make sure all case is covered
func (dst *Numeric) Set(src interface{}) error {
	if src == nil {
		*dst = Numeric{Status: pgtype.Null}

		return nil
	}

	if value, ok := src.(interface{ Get() interface{} }); ok {
		value2 := value.Get()
		if value2 != value {
			return dst.Set(value2)
		}
	}

	switch value := src.(type) {
	case decimal.Big:
		*dst = Numeric{Decimal: value, Status: pgtype.Present}
	case float32:
		*dst = Numeric{Decimal: *new(decimal.Big).SetFloat64(float64(value)), Status: pgtype.Present}
	case float64:
		*dst = Numeric{Decimal: *new(decimal.Big).SetFloat64(value), Status: pgtype.Present}
	case int8:
		*dst = Numeric{Decimal: *decimal.New(int64(value), 0), Status: pgtype.Present}
	case uint8:
		*dst = Numeric{Decimal: *decimal.New(int64(value), 0), Status: pgtype.Present}
	case int16:
		*dst = Numeric{Decimal: *decimal.New(int64(value), 0), Status: pgtype.Present}
	case uint16:
		*dst = Numeric{Decimal: *decimal.New(int64(value), 0), Status: pgtype.Present}
	case int32:
		*dst = Numeric{Decimal: *decimal.New(int64(value), 0), Status: pgtype.Present}
	case uint32:
		*dst = Numeric{Decimal: *decimal.New(int64(value), 0), Status: pgtype.Present}
	case int64:
		*dst = Numeric{Decimal: *decimal.New(value, 0), Status: pgtype.Present}
	case uint64:
		// uint64 could be greater than int64 so convert to string then to decimal
		dec, ok := new(decimal.Big).SetString(strconv.FormatUint(value, base))
		if !ok {
			return fmt.Errorf("set: %w", errConversionFailed)
		}

		*dst = Numeric{Decimal: *dec, Status: pgtype.Present}
	case int:
		*dst = Numeric{Decimal: *decimal.New(int64(value), 0), Status: pgtype.Present}
	case uint:
		// uint could be greater than int64 so convert to string then to decimal
		dec, ok := new(decimal.Big).SetString(strconv.FormatUint(uint64(value), base))
		if !ok {
			return fmt.Errorf("set: %w", errConversionFailed)
		}

		*dst = Numeric{Decimal: *dec, Status: pgtype.Present}
	case string:
		dec, ok := new(decimal.Big).SetString(value)
		if !ok {
			return fmt.Errorf("set: %w", errConversionFailed)
		}

		*dst = Numeric{Decimal: *dec, Status: pgtype.Present}
	default:
		// If all else fails see if pgtype.Numeric can handle it. If so, translate through that.
		num := &pgtype.Numeric{}
		if err := num.Set(value); err != nil {
			return fmt.Errorf("cannot convert %v to Numeric: %w", value, err)
		}

		buf, err := num.EncodeText(nil, nil)
		if err != nil {
			return fmt.Errorf("cannot convert %v to Numeric: %w", value, err)
		}

		dec, ok := new(decimal.Big).SetString(string(buf))
		if !ok {
			return fmt.Errorf("set: %w", errConversionFailed)
		}

		*dst = Numeric{Decimal: *dec, Status: pgtype.Present}
	}

	return nil
}

func (dst *Numeric) Get() interface{} {
	switch {
	case dst.Status == pgtype.Present:
		return dst.Decimal
	case dst.Status == pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

// nolint: gocognit, gocyclo, revive, cyclop, stylecheck // need to check each type to make sure all case is covered
func (src *Numeric) AssignTo(dst interface{}) error {
	switch {
	case src.Status == pgtype.Present:
		switch val := dst.(type) {
		case *decimal.Big:
			*val = src.Decimal
		case *float32:
			f, _ := src.Decimal.Float64()
			*val = float32(f)
		case *float64:
			f, _ := src.Decimal.Float64()
			*val = f
		case *int:
			if src.Decimal.Scale() < 0 {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			n, err := strconv.ParseInt(src.Decimal.String(), base, strconv.IntSize)
			if err != nil {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			*val = int(n)
		case *int8:
			if src.Decimal.Scale() < 0 {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			n, err := strconv.ParseInt(src.Decimal.String(), base, bitSize8)
			if err != nil {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			*val = int8(n)
		case *int16:
			if src.Decimal.Scale() < 0 {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			n, err := strconv.ParseInt(src.Decimal.String(), base, bitSize16)
			if err != nil {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			*val = int16(n)
		case *int32:
			if src.Decimal.Scale() < 0 {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			n, err := strconv.ParseInt(src.Decimal.String(), base, bitSize32)
			if err != nil {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			*val = int32(n)
		case *int64:
			if src.Decimal.Scale() < 0 {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			n, err := strconv.ParseInt(src.Decimal.String(), base, bitSize64)
			if err != nil {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			*val = n
		case *uint:
			if src.Decimal.Scale() < 0 || src.Decimal.Sign() < 0 {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			n, err := strconv.ParseUint(src.Decimal.String(), base, strconv.IntSize)
			if err != nil {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			*val = uint(n)
		case *uint8:
			if src.Decimal.Scale() < 0 || src.Decimal.Sign() < 0 {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			n, err := strconv.ParseUint(src.Decimal.String(), base, bitSize8)
			if err != nil {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			*val = uint8(n)
		case *uint16:
			if src.Decimal.Scale() < 0 || src.Decimal.Sign() < 0 {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			n, err := strconv.ParseUint(src.Decimal.String(), base, bitSize16)
			if err != nil {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			*val = uint16(n)
		case *uint32:
			if src.Decimal.Scale() < 0 || src.Decimal.Sign() < 0 {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			n, err := strconv.ParseUint(src.Decimal.String(), base, bitSize32)
			if err != nil {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			*val = uint32(n)
		case *uint64:
			if src.Decimal.Scale() < 0 || src.Decimal.Sign() < 0 {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			n, err := strconv.ParseUint(src.Decimal.String(), base, bitSize64)
			if err != nil {
				return fmt.Errorf("%w: %v to %T", errConversionFailed, dst, *val)
			}

			*val = n
		default:
			if nextDst, retry := pgtype.GetAssignToDstType(dst); retry {
				return src.AssignTo(nextDst)
			}

			return fmt.Errorf("%w: %T", errAssignFailed, dst)
		}
	case src.Status == pgtype.Null:
		if err := pgtype.NullAssignTo(dst); err != nil {
			return fmt.Errorf("%w: %T", errAssignFailed, dst)
		}

		return nil
	case src.Status == pgtype.Undefined:
		return fmt.Errorf("AssignTo: %w", ErrUndefined)
	}

	return nil
}

func (dst *Numeric) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = Numeric{Status: pgtype.Null}

		return nil
	}

	dec, ok := new(decimal.Big).SetString(string(src))
	if !ok {
		return fmt.Errorf("set: %w", errConversionFailed)
	}

	*dst = Numeric{Decimal: *dec, Status: pgtype.Present}

	return nil
}

func (dst *Numeric) DecodeBinary(connInfo *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = Numeric{Status: pgtype.Null}

		return nil
	}

	// For now at least, implement this in terms of pgtype.Numeric

	num := &pgtype.Numeric{}
	if err := num.DecodeBinary(connInfo, src); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	*dst = Numeric{Decimal: *new(decimal.Big).SetBigMantScale(num.Int, -int(num.Exp)), Status: pgtype.Present}

	return nil
}

// nolint: revive // different naming is to diffrentiate source and destination
func (src *Numeric) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch {
	case src.Status == pgtype.Null:
		return nil, nil
	case src.Status == pgtype.Undefined:
		return nil, ErrUndefined
	}

	return append(buf, src.Decimal.String()...), nil
}

// nolint: revive // different naming is to diffrentiate source and destination
func (src *Numeric) EncodeBinary(connInfo *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch {
	case src.Status == pgtype.Null:
		return nil, nil
	case src.Status == pgtype.Undefined:
		return nil, ErrUndefined
	}

	// For now at least, implement this in terms of pgtype.Numeric
	num := &pgtype.Numeric{}
	if err := num.DecodeText(connInfo, []byte(fmt.Sprintf("%f", &src.Decimal))); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	bytes, err := num.EncodeBinary(connInfo, buf)
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	return bytes, nil
}

// Scan implements the database/sql Scanner interface.
func (dst *Numeric) Scan(src interface{}) error {
	if src == nil {
		*dst = Numeric{Status: pgtype.Null}

		return nil
	}

	switch src := src.(type) {
	case float64:
		*dst = Numeric{Decimal: *new(decimal.Big).SetFloat64(src), Status: pgtype.Present}

		return nil
	case string:
		return dst.DecodeText(nil, []byte(src))
	case []byte:
		return dst.DecodeText(nil, src)
	}

	return fmt.Errorf("%w %T", errScanFailed, src)
}

// Value implements the database/sql/driver Valuer interface.
// nolint: revive // different naming is to diffrentiate source and destination
func (src *Numeric) Value() (driver.Value, error) {
	switch {
	case src.Status == pgtype.Present:
		return src.Decimal.String(), nil
	case src.Status == pgtype.Null:
		return nil, nil
	default:
		return nil, ErrUndefined
	}
}

// nolint: revive // different naming is to diffrentiate source and destination
func (src *Numeric) MarshalJSON() ([]byte, error) {
	switch src.Status {
	case pgtype.Present:
		bytes, err := src.Decimal.MarshalText()
		if err != nil {
			return nil, fmt.Errorf("marshal: %w", err)
		}

		return bytes, nil
	case pgtype.Null:
		return []byte("null"), nil
	case pgtype.Undefined:
		return nil, ErrUndefined
	}

	return nil, ErrBadStatus
}

func (dst *Numeric) UnmarshalJSON(bytes []byte) error {
	dec := new(decimal.Big)

	if err := dec.UnmarshalJSON(bytes); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	status := pgtype.Null
	if string(bytes) != "null" {
		status = pgtype.Present
	}

	*dst = Numeric{Decimal: *dec, Status: status}

	return nil
}
