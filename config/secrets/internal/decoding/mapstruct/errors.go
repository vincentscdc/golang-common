package mapstruct

import "fmt"

type TagMismatchError struct {
	TagName string
}

func (e TagMismatchError) Error() string {
	return fmt.Sprintf("tag %s is not found in the source map", e.TagName)
}

type ValueTypeMismatchError struct {
	FieldName string
	FieldType string
	ValueType string
}

func (e ValueTypeMismatchError) Error() string {
	return fmt.Sprintf("can not use value of type %s as value of field %s (type %s)",
		e.ValueType, e.FieldName, e.FieldType)
}

type OverflowError struct {
	FieldName string
	FieldType string
	ValueType string
}

func (e OverflowError) Error() string {
	return fmt.Sprintf("can not decode %s, %s overflows %s", e.FieldName, e.ValueType, e.FieldType)
}

type RepeatedTagError string

func (e RepeatedTagError) Error() string {
	return fmt.Sprintf("repeated tag %s", string(e))
}

type DecodeError string

func (e DecodeError) Error() string {
	return string(e)
}
