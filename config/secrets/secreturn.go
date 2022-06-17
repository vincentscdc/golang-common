package secrets

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/mitchellh/mapstructure"
	"github.com/monacohq/golang-common/config/secrets/common"

	"github.com/monacohq/golang-common/config/secrets/internal/provider/awssm"
	"github.com/monacohq/golang-common/config/secrets/internal/provider/local"
)

// SecretUrn will retrieve secrets from a secrets provider
type SecretUrn map[string]any

// Bind unmarshalls the secret items into a user-defined structure
func (sm SecretUrn) Bind(v any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:    "secret_key",
		DecodeHook: decodeHook,
		Result:     v,
	})
	if err != nil {
		return fmt.Errorf("new decoder error: %w", err)
	}

	if err := decoder.Decode(sm); err != nil {
		return fmt.Errorf("bind error: %w", err)
	}

	return nil
}

// decodeHook uses reflection to automatically use helpers to transform values when mapping
func decodeHook(fromType, toType reflect.Type, value interface{}) (interface{}, error) {
	if fromType.Kind() == reflect.String {
		stringValue, ok := value.(string)
		if !ok {
			return 0, common.SecretValueTypeError{}
		}

		switch {
		case toType.Kind() == reflect.Int:
			return castStringToInt(stringValue)
		case toType.Kind() == reflect.Bool:
			return castStringToBool(stringValue)
		case toType.Kind() == reflect.Float64:
			return castStringToFloat64(stringValue)
		default:
			return value, nil
		}
	}

	return value, nil
}

// NewSecretUrnFromConfig returns SecretUrn from a SecreteConfig provided by the caller
// It is used for internal providers from this library core.
func NewSecretUrnFromConfig(config common.SecretsConfig) (SecretUrn, error) {
	var provider Provider
	switch config.(type) {
	case *common.SecretsConfigLocal:
		provider = local.NewFromConfig(config)
	case *common.SecretsConfigAWS:
		provider = awssm.NewFromConfig(config, awssm.ClientImpl{})
	default:
		return nil, common.SecretProviderUnknownError(config.Name())
	}

	return NewSecretUrnFromProvider(provider)
}

// NewSecretUrnFromProvider returns SecretUrn from a customized provider by the caller
// It is used for external providers which can be a customized one from the caller.
// The external provider must be enforced to implement the Provider interface.
func NewSecretUrnFromProvider(provider Provider) (SecretUrn, error) {
	secretValue, err := provider.GetSecret()
	if err != nil {
		return nil, fmt.Errorf("retrieve secrets from provider error: %w", err)
	}

	return SecretUrn(secretValue), nil
}

// getSecretItemValue retrieves the item value from the secret value
// ErrSecretItemNotSet returns if the item key is not found
func (sm SecretUrn) getSecretItemValue(key string) (any, error) {
	if item, ok := sm[key]; ok {
		return item, nil
	}

	return nil, common.SecretItemNotFoundError(key)
}

func (sm SecretUrn) GetSecretBool(key string) (bool, error) {
	item, err := sm.getSecretItemValue(key)
	if err != nil {
		return false, err
	}

	return castToBool(item)
}

func castToBool(v any) (bool, error) {
	if b, ok := v.(bool); ok {
		return b, nil
	}

	return false, common.SecretValueTypeError{}
}

func (sm SecretUrn) GetSecretFloat64(key string) (float64, error) {
	secret, err := sm.getSecretItemValue(key)
	if err != nil {
		return 0.0, err
	}

	return castToFloat64(secret)
}

func castToFloat64(v any) (float64, error) {
	if f, ok := v.(float64); ok {
		return f, nil
	}

	return 0.0, common.SecretValueTypeError{}
}

func (sm SecretUrn) GetSecretInt(key string) (int, error) {
	secret, err := sm.getSecretItemValue(key)
	if err != nil {
		return 0.0, err
	}

	return castToInt(secret)
}

func castToInt(v any) (int, error) {
	if i, ok := v.(int); ok {
		return i, nil
	}

	return 0, common.SecretValueTypeError{}
}

func (sm SecretUrn) GetSecretIntSlice(key string) ([]int, error) {
	secret, err := sm.getSecretItemValue(key)
	if err != nil {
		return nil, err
	}

	return castToIntSlice(secret)
}

func castToIntSlice(v any) ([]int, error) {
	sif, ok := v.([]any)
	if !ok {
		return nil, common.SecretItemValueTypeError{} // not slice
	}

	sliceInt := make([]int, 0, len(sif))

	for _, e := range sif {
		e, ok := e.(int)
		if !ok { // not int
			return nil, common.SecretItemValueTypeError{}
		}

		sliceInt = append(sliceInt, e)
	}

	return sliceInt, nil
}

func (sm SecretUrn) GetSecretString(key string) (string, error) {
	secret, err := sm.getSecretItemValue(key)
	if err != nil {
		return "", err
	}

	return castToString(secret)
}

func castToString(v any) (string, error) {
	if s, ok := v.(string); ok {
		return s, nil
	}

	return "", common.SecretValueTypeError{}
}

func (sm SecretUrn) GetSecretStringSlice(key string) ([]string, error) {
	secret, err := sm.getSecretItemValue(key)
	if err != nil {
		return nil, err
	}

	return castToStringSlice(secret)
}

func castToStringSlice(v any) ([]string, error) {
	sif, ok := v.([]any)
	if !ok { // not slice
		return nil, common.SecretItemValueTypeError{}
	}

	sliceString := make([]string, 0, len(sif))

	for _, e := range sif {
		e, ok := e.(string)
		if !ok { // not string
			return nil, common.SecretItemValueTypeError{}
		}

		sliceString = append(sliceString, e)
	}

	return sliceString, nil
}

func castStringToInt(v string) (int, error) {
	result, err := strconv.Atoi(v)
	if err != nil {
		return 0, common.SecretValueCastError{Val: v, Err: err}
	}

	return result, nil
}

func castStringToBool(v string) (bool, error) {
	result, err := strconv.ParseBool(v)
	if err != nil {
		return false, common.SecretValueCastError{Val: v, Err: err}
	}

	return result, nil
}

func castStringToFloat64(v string) (float64, error) {
	const bitsInt64 = 64

	result, err := strconv.ParseFloat(v, bitsInt64)
	if err != nil {
		return 0, common.SecretValueCastError{Val: v, Err: err}
	}

	return result, nil
}

func (sm SecretUrn) IsSecretSet(key string) bool {
	_, err := sm.getSecretItemValue(key)

	return !errors.Is(err, error(common.SecretItemNotFoundError(key)))
}
