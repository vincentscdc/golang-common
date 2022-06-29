package mapstruct

import (
	"reflect"
)

type Decoder struct {
	TagName string

	// wl is a waiting list of unscanned struct
	wl []reflect.Value

	tags map[string]int // used to detect repeated tags
}

const DefaultTagName = "secret_key"

func (d *Decoder) Decode(src map[string]any, dst any) error {
	dstv := reflect.ValueOf(dst)

	if dstv.Kind() != reflect.Ptr {
		return DecodeError("destination must be a pointer")
	}

	dste := dstv.Elem()

	if dste.Kind() != reflect.Struct {
		return DecodeError("destination must be of type struct")
	}

	if d.TagName == "" {
		d.TagName = DefaultTagName
	}

	d.wl = []reflect.Value{dste}
	d.tags = make(map[string]int)

	return d.loop(src)
}

func (d *Decoder) loop(src map[string]any) error {
	for {
		if len(d.wl) == 0 {
			return nil
		}

		dst := d.wl[0]
		d.wl = d.wl[1:]

		if err := d.decode(src, dst); err != nil {
			return err
		}
	}
}

func (d *Decoder) decode(src map[string]any, dst reflect.Value) error {
	dstTyp := dst.Type()

	for idx := 0; idx < dstTyp.NumField(); idx++ {
		field := dstTyp.Field(idx)

		// ignore not exported
		if !field.IsExported() {
			continue
		}

		if field.Type.Kind() == reflect.Struct {
			d.wl = append(d.wl, dst.FieldByName(field.Name))

			continue
		}

		key := lookupTagValueByName(&field, d.TagName)
		if key == "" {
			continue
		}

		d.tags[key]++
		if d.tags[key] > 1 {
			return RepeatedTagError(key)
		}

		if val, ok := src[key]; ok {
			if err := d.setValue(field.Name, dst.Field(idx), val); err != nil {
				return err
			}
		} else {
			return TagMismatchError{TagName: key}
		}
	}

	return nil
}

func (d *Decoder) setValue(name string, field reflect.Value, val any) error {
	var err error

	// nolint:exhaustive // not listed types are not being supported
	switch kind := field.Kind(); kind {
	case reflect.Bool:
		err = d.decodeBool(name, field, val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		err = d.decodeInt(name, field, val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = d.decodeUint(name, field, val)
	case reflect.Float32, reflect.Float64:
		err = d.decodeFloat(name, field, val)
	case reflect.String:
		err = d.decodeString(name, field, val)
	default:
		err = DecodeError("not supported type " + kind.String())
	}

	return err
}

func (d *Decoder) decodeBool(name string, field reflect.Value, val any) error {
	rvl := reflect.Indirect(reflect.ValueOf(val))

	if rvl.Kind() != reflect.Bool {
		return ValueTypeMismatchError{
			FieldName: name,
			FieldType: field.Type().Name(),
			ValueType: rvl.Type().Name(),
		}
	}

	field.SetBool(rvl.Bool())

	return nil
}

func (d *Decoder) decodeInt(name string, field reflect.Value, val any) error {
	rvl := reflect.Indirect(reflect.ValueOf(val))

	// unmarshalling JSON number into an interface will be stored as float64
	if rvl.Kind() != reflect.Float64 {
		return ValueTypeMismatchError{
			FieldName: name,
			FieldType: field.Type().Name(),
			ValueType: rvl.Type().Name(),
		}
	}

	intval := int64(rvl.Float())

	if field.OverflowInt(intval) {
		return OverflowError{
			FieldName: name,
			FieldType: field.Type().Name(),
			ValueType: rvl.Type().Name(),
		}
	}

	field.SetInt(intval)

	return nil
}

func (d *Decoder) decodeUint(name string, field reflect.Value, val any) error {
	rvl := reflect.Indirect(reflect.ValueOf(val))

	// unmarshalling JSON number into an interface will be stored as float64
	if rvl.Kind() != reflect.Float64 {
		return ValueTypeMismatchError{
			FieldName: name,
			FieldType: field.Type().Name(),
			ValueType: rvl.Type().Name(),
		}
	}

	fval := rvl.Float()
	uintval := uint64(fval)

	if fval < 0 || field.OverflowUint(uintval) {
		return OverflowError{
			FieldName: name,
			FieldType: field.Type().Name(),
			ValueType: rvl.Type().Name(),
		}
	}

	field.SetUint(uintval)

	return nil
}

func (d *Decoder) decodeFloat(name string, field reflect.Value, val any) error {
	rvl := reflect.Indirect(reflect.ValueOf(val))

	// unmarshalling JSON number into an interface will be stored as float64
	if rvl.Kind() != reflect.Float64 {
		return ValueTypeMismatchError{
			FieldName: name,
			FieldType: field.Type().Name(),
			ValueType: rvl.Type().Name(),
		}
	}

	fval := rvl.Float()

	if field.OverflowFloat(fval) {
		return OverflowError{
			FieldName: name,
			FieldType: field.Type().Name(),
			ValueType: rvl.Type().Name(),
		}
	}

	field.SetFloat(fval)

	return nil
}

func (d *Decoder) decodeString(name string, field reflect.Value, val any) error {
	rvl := reflect.Indirect(reflect.ValueOf(val))

	if rvl.Kind() != reflect.String {
		return ValueTypeMismatchError{
			FieldName: name,
			FieldType: field.Type().Name(),
			ValueType: rvl.Type().Name(),
		}
	}

	field.SetString(rvl.String())

	return nil
}

func lookupTagValueByName(f *reflect.StructField, name string) string {
	if val, ok := f.Tag.Lookup(name); ok {
		return val
	}

	return ""
}
