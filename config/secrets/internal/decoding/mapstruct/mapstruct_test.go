package mapstruct

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestDecodeNotPointer(t *testing.T) {
	t.Parallel()

	if err := new(Decoder).Decode(make(map[string]any), struct{}{}); err == nil {
		t.Error("should return error")
	}
}

func TestDecodeNotStructType(t *testing.T) {
	t.Parallel()

	dst := ""
	if err := new(Decoder).Decode(make(map[string]any), &dst); err == nil {
		t.Error("should return error")
	}
}

func TestDecodeSimpleStruct(t *testing.T) {
	t.Parallel()

	var testdata map[string]any
	_ = json.Unmarshal([]byte(`
{
	"String_Key": "helloworld",
	"Bool_Key": true,
	"Integer_Key": 123,
	"Float_Key": 0.5
}`), &testdata)

	dst := struct {
		Field1 string  `secret_key:"String_Key"`
		Field2 bool    `secret_key:"Bool_Key"`
		Field3 int     `secret_key:"Integer_Key"`
		Field4 float64 `secret_key:"Float_Key"`
		Field5 string
	}{Field5: "existing_value"}

	err := new(Decoder).Decode(testdata, &dst)
	if err != nil {
		t.Error(err)
	}

	if expected, _ := testdata["String_Key"].(string); dst.Field1 != expected {
		t.Errorf("expected %v, but got %v", expected, dst.Field1)
	}

	if expected, _ := testdata["Bool_Key"].(bool); dst.Field2 != expected {
		t.Errorf("expected %v, but got %v", expected, dst.Field2)
	}

	if expected, _ := testdata["Integer_Key"].(float64); dst.Field3 != int(expected) {
		t.Errorf("expected %v, but got %v", expected, dst.Field3)
	}

	if expected, _ := testdata["Float_Key"].(float64); dst.Field4 != expected {
		t.Errorf("expected %v, but got %v", expected, dst.Field4)
	}

	if dst.Field5 != "existing_value" {
		t.Errorf("expected \"existing_value\", but got %s", dst.Field5)
	}
}

func TestDecodeNestedStruct(t *testing.T) {
	t.Parallel()

	var testdata map[string]any
	_ = json.Unmarshal([]byte(`
{
	"String_Key": "helloworld",
	"Bool_Key": true,
	"Integer_Key": 123,
	"Float_Key": 0.5
}`), &testdata)

	dst := struct {
		Field       string `secret_key:"String_Key"`
		InnerStruct struct {
			Field       bool `secret_key:"Bool_Key"`
			InnerStruct struct {
				Field       int `secret_key:"Integer_Key"`
				InnerStruct struct {
					Field float64 `secret_key:"Float_Key"`
				}
			}
		}
	}{}

	err := new(Decoder).Decode(testdata, &dst)
	if err != nil {
		t.Error(err)
	}

	if expected, _ := testdata["String_Key"].(string); dst.Field != expected {
		t.Errorf("expected %v, but got %v", expected, dst.Field)
	}

	if expected, _ := testdata["Bool_Key"].(bool); dst.InnerStruct.Field != expected {
		t.Errorf("expected %v, but got %v", expected, dst.InnerStruct.Field)
	}

	if expected, _ := testdata["Integer_Key"].(float64); dst.InnerStruct.InnerStruct.Field != int(expected) {
		t.Errorf("expected %v, but got %v", expected, dst.InnerStruct.InnerStruct.Field)
	}

	if expected, _ := testdata["Float_Key"].(float64); dst.InnerStruct.InnerStruct.InnerStruct.Field != expected {
		t.Errorf("expected %v, but got %v", expected, dst.InnerStruct.InnerStruct.InnerStruct.Field)
	}
}

func TestDecodeDuplicateTag(t *testing.T) {
	t.Parallel()

	var testdata map[string]any
	_ = json.Unmarshal([]byte(`
{
	"String_Key": "helloworld"
}`), &testdata)

	dst := struct {
		Field       string `secret_key:"String_Key"`
		InnerStruct struct {
			Field string `secret_key:"String_Key"`
		}
	}{}

	err := new(Decoder).Decode(testdata, &dst)
	if err == nil {
		t.Error("expect non nil")
	}

	terr := RepeatedTagError("String_Key")
	if !errors.As(err, &terr) {
		t.Errorf("expect %v got %v", terr, err)
	}
}

func TestDecodePrivateField(t *testing.T) {
	t.Parallel()

	var testdata map[string]any
	_ = json.Unmarshal([]byte(`
{
	"String_Key": "helloworld"
}`), &testdata)

	dst := struct {
		field string `secret_key:"String_Key"`
	}{}

	err := new(Decoder).Decode(testdata, &dst)
	if err != nil {
		t.Error(err)
	}

	if dst.field != "" {
		t.Errorf("expected empty, but got %v", dst.field)
	}
}

func TestDecodeCustomTagName(t *testing.T) {
	t.Parallel()

	var testdata map[string]any
	_ = json.Unmarshal([]byte(`
{
	"String_Key": "helloworld"
}`), &testdata)

	dst := struct {
		Field string `custom_tag:"String_Key"`
	}{}

	err := (&Decoder{
		TagName: "custom_tag",
	}).Decode(testdata, &dst)
	if err != nil {
		t.Error(err)
	}

	if expected, _ := testdata["String_Key"].(string); dst.Field != expected {
		t.Errorf("expected %v, but got %v", expected, dst.Field)
	}
}

func TestDecodeEmbeddedField(t *testing.T) {
	t.Parallel()

	var testdata map[string]any
	_ = json.Unmarshal([]byte(`
{
	"String_Key": "helloworld"
}`), &testdata)

	type EmbeddedStruct struct {
		EmbeddedField string `secret_key:"String_Key"`
	}

	dst := struct {
		Field string

		EmbeddedStruct
	}{}

	err := new(Decoder).Decode(testdata, &dst)
	if err != nil {
		t.Error(err)
	}

	expected, _ := testdata["String_Key"].(string)

	if dst.EmbeddedField != expected {
		t.Errorf("expected %v, but got %v", expected, dst.EmbeddedField)
	}

	if dst.EmbeddedStruct.EmbeddedField != expected {
		t.Errorf("expected %v, but got %v", expected, dst.EmbeddedStruct.EmbeddedField)
	}
}

func TestTypeDecode(t *testing.T) {
	t.Parallel()

	s := struct {
		Bool bool

		String string

		Int   int
		Int8  int8
		Int16 int16
		Int32 int32
		Int64 int64

		Uint   uint
		Uint8  uint8
		Uint16 uint16
		Uint32 uint32
		Uint64 uint64

		Float32 float32
		Float64 float64
	}{}

	refv := reflect.ValueOf(&s).Elem()

	testCases := []struct {
		name  string
		field reflect.Value
		value any
		exp   any
	}{
		{"bool", refv.FieldByName("Bool"), true, true},
		{"string", refv.FieldByName("String"), "hello world", "hello world"},
		{"int->int", refv.FieldByName("Int"), 123, 123},
		{"uint->int", refv.FieldByName("Int"), uint(123), int(123)},
		{"float->int", refv.FieldByName("Int"), float64(123), int(float64(123))},
		{"int->int8", refv.FieldByName("Int8"), 123, int8(123)},
		{"uint->int8", refv.FieldByName("Int8"), uint(123), int8(123)},
		{"float->int8", refv.FieldByName("Int8"), float64(123), int8(123)},
		{"int->int16", refv.FieldByName("Int16"), 123, int16(123)},
		{"uint->int16", refv.FieldByName("Int16"), uint(123), int16(123)},
		{"float->int16", refv.FieldByName("Int16"), float64(123), int16(123)},
		{"int->int32", refv.FieldByName("Int32"), 123, int32(123)},
		{"uint->int32", refv.FieldByName("Int32"), uint(123), int32(123)},
		{"float->int32", refv.FieldByName("Int32"), float64(123), int32(123)},
		{"int->int64", refv.FieldByName("Int64"), 123, int64(123)},
		{"uint->int64", refv.FieldByName("Int64"), uint(123), int64(123)},
		{"float->int64", refv.FieldByName("Int64"), float64(123), int64(123)},
		{"uint->uint", refv.FieldByName("Uint"), uint(123), uint(123)},
		{"int->uint", refv.FieldByName("Uint"), int(123), uint(123)},
		{"float->uint", refv.FieldByName("Uint"), float64(123), uint(123)},
		{"uint->uint8", refv.FieldByName("Uint8"), uint(123), uint8(123)},
		{"int->uint8", refv.FieldByName("Uint8"), int(123), uint8(123)},
		{"float->uint8", refv.FieldByName("Uint8"), float64(123), uint8(123)},
		{"uint->uint16", refv.FieldByName("Uint16"), uint(123), uint16(123)},
		{"int->uint16", refv.FieldByName("Uint16"), int(123), uint16(123)},
		{"float64->uint16", refv.FieldByName("Uint16"), float64(123), uint16(123)},
		{"uint->uint32", refv.FieldByName("Uint32"), uint(123), uint32(123)},
		{"int->uint32", refv.FieldByName("Uint32"), int(123), uint32(123)},
		{"float->uint32", refv.FieldByName("Uint32"), float64(123), uint32(123)},
		{"uint->uint64", refv.FieldByName("Uint64"), uint(123), uint64(123)},
		{"int->uint64", refv.FieldByName("Uint64"), 123, uint64(123)},
		{"float->uint64", refv.FieldByName("Uint64"), float64(123), uint64(123)},
		{"int->float32", refv.FieldByName("Float32"), 123, float32(123)},
		{"uint->float32", refv.FieldByName("Float32"), uint(123), float32(123)},
		{"float->float32", refv.FieldByName("Float32"), float64(123), float32(123)},
		{"int->float64", refv.FieldByName("Float64"), 123, float64(123)},
		{"uint->float64", refv.FieldByName("Float64"), uint(123), float64(123)},
		{"float->float64", refv.FieldByName("Float64"), float64(123), float64(123)},
	}

	decoder := &Decoder{}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if err := decoder.setValue("", tc.field, tc.value); err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(tc.field.Interface(), tc.exp) {
				t.Errorf("expect %v, but get %v", tc.exp, tc.field.Interface())
			}
		})
	}
}

func TestTypeDecodeOverflow(t *testing.T) {
	t.Parallel()

	s := struct {
		Int8  int8
		Int16 int16
		Int32 int32
		Int64 int64

		Uint   uint
		Uint8  uint8
		Uint16 uint16
		Uint32 uint32

		Float32 float32
	}{}

	refv := reflect.ValueOf(&s).Elem()

	testCases := []struct {
		name  string
		field reflect.Value
		value any
	}{
		{"int8", refv.FieldByName("Int8"), 1 << 7},
		{"int16", refv.FieldByName("Int16"), 1 << 15},
		{"int32", refv.FieldByName("Int32"), 1 << 31},
		{"uint64->int64", refv.FieldByName("Int64"), uint64(1) << 63},
		{"float->int64", refv.FieldByName("Int64"), math.MaxFloat64},
		{"uint8", refv.FieldByName("Uint8"), 1 << 8},
		{"uint16", refv.FieldByName("Uint16"), 1 << 16},
		{"uint32", refv.FieldByName("Uint32"), 1 << 32},
		{"negative->uint", refv.FieldByName("Uint"), -1},
		{"float->uint", refv.FieldByName("Uint"), math.MaxFloat64},
		{"float32", refv.FieldByName("Float32"), math.MaxFloat64},
	}

	decoder := &Decoder{}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if err := decoder.setValue("", tc.field, tc.value); err == nil {
				t.Errorf("%v should be overflow %s", tc.value, tc.field.Kind().String())
			}
		})
	}
}

func TestTypeDecodeNonStringValueToStringType(t *testing.T) {
	t.Parallel()

	s := struct {
		String string
	}{}

	dec := &Decoder{}

	val := 123

	err := dec.decodeString("", reflect.ValueOf(&s).Elem().FieldByName("String"), val)
	if err == nil {
		t.Error("should be returned error")
	}
}

func TestTypeDecodeNotSupportedType(t *testing.T) {
	t.Parallel()

	s := struct {
		UnsupportedType []int
	}{}

	dec := &Decoder{}

	err := dec.setValue("", reflect.ValueOf(&s).Elem().FieldByName("UnsupportedType"), []int{1})
	if err == nil {
		t.Error("should be returned error")
	}
}

func TestTyepDecodeTypeMismatch(t *testing.T) {
	t.Parallel()

	s := struct {
		Bool bool

		Int int

		Uint uint

		Float32 float32
	}{}

	refv := reflect.ValueOf(&s).Elem()

	testCases := []struct {
		name  string
		field reflect.Value
		val   any
		exp   error
	}{
		{"int->bool", refv.FieldByName("Bool"), 1, ValueTypeMismatchError{"", refv.FieldByName("Bool").Type().Name(), "int"}},
		{"bool->int", refv.FieldByName("Int"), true, ValueTypeMismatchError{"", refv.FieldByName("Int").Type().Name(), "bool"}},
		{"string->int", refv.FieldByName("Int"), "string", ValueTypeMismatchError{"", refv.FieldByName("Int").Type().Name(), "string"}},
		{"bool->uint", refv.FieldByName("Uint"), true, ValueTypeMismatchError{"", refv.FieldByName("Uint").Type().Name(), "bool"}},
		{"string->uint", refv.FieldByName("Uint"), "string", ValueTypeMismatchError{"", refv.FieldByName("Uint").Type().Name(), "string"}},
		{"bool->float", refv.FieldByName("Float32"), true, ValueTypeMismatchError{"", refv.FieldByName("Float32").Type().Name(), "bool"}},
		{"string->float", refv.FieldByName("Float32"), "string", ValueTypeMismatchError{"", refv.FieldByName("Float32").Type().Name(), "string"}},
	}

	dec := &Decoder{}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := dec.setValue("", tc.field, tc.val)
			if !errors.Is(err, tc.exp) {
				t.Error("expect to ValueTypeMismatchError, but not")
			}
		})
	}
}

func TestDecodeTagMismatch(t *testing.T) {
	t.Parallel()

	s := struct {
		Field string `secret_key:"non_exists_key"`
	}{}

	dec := &Decoder{}
	emptyMap := make(map[string]any)
	expErr := TagMismatchError{TagName: "non_exists_key"}

	err := dec.Decode(emptyMap, &s)

	if !errors.Is(err, expErr) {
		t.Error("expect to TagMismatchError, but not")
	}
}
