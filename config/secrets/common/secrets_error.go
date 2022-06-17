package common

import (
	"fmt"
)

type SecretValueCastError struct {
	Val string
	Err error
}

func (s SecretValueCastError) Error() string {
	return fmt.Sprintf("secret value casting (%s) error: %v", s.Val, s.Err)
}

type SecretValueTypeError struct{}

func (SecretValueTypeError) Error() string {
	return "unexpected secret value type"
}

type SecretItemValueTypeError struct{}

func (SecretItemValueTypeError) Error() string {
	return "unexpected secret item value type"
}

type SecretNotFoundError string

func (e SecretNotFoundError) Error() string {
	return fmt.Sprintf("secret `%v' not found", string(e))
}

type SecretItemNotFoundError string

func (e SecretItemNotFoundError) Error() string {
	return fmt.Sprintf("secret item `%v' not found", string(e))
}

type SecretProviderError struct{}

func (SecretProviderError) Error() string {
	return "secret provider error"
}

type SecretFileFormatError string

func (e SecretFileFormatError) Error() string {
	return fmt.Sprintf("unsupported secret file format: %v", string(e))
}

type SecretProviderUnknownError string

func (e SecretProviderUnknownError) Error() string {
	return fmt.Sprintf("unknown secret provider: %v", string(e))
}

type DecoderNotSetError string

func (e DecoderNotSetError) Error() string {
	return fmt.Sprintf("mapstructure decoder not set: %v", string(e))
}
