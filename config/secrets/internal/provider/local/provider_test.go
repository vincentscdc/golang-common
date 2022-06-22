package local

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/monacohq/golang-common/config/secrets/common"
)

func TestNewFromConfig(t *testing.T) {
	t.Parallel()

	if provider := NewFromConfig(nil); provider != nil {
		t.Fatalf("expect nil")
	}

	if provider := NewFromConfig(common.SecretsConfig(nil)); provider != nil {
		t.Fatalf("expect nil")
	}

	if provider := NewFromConfig(&common.SecretsConfigLocal{}); provider == nil {
		t.Fatalf("expect non nil")
	}
}

func TestGetSecret(t *testing.T) {
	t.Parallel()

	provider := SecretsProvider{
		config: &common.SecretsConfigLocal{
			Path: "invalid local secrets file",
		},
	}

	secretValue, err := provider.GetSecret(context.TODO())
	if err == nil {
		t.Fatalf("expect non-nil")
	}

	if secretValue != nil {
		t.Fatalf("expect nil")
	}
}

type mockDecoder func(v any) error

func (d mockDecoder) Decode(v any) error {
	return d(v)
}

func Test_decodeConfig(t *testing.T) {
	t.Parallel()

	if _, err := decodeConfig(mockDecoder(func(v any) error {
		return fmt.Errorf("unsupport file type")
	})); err == nil {
		t.Fatalf("expect parse file error")
	}
}

func TestGetSecretFromUnknownFormat(t *testing.T) {
	t.Parallel()

	sm := SecretsProvider{
		config: &common.SecretsConfigLocal{
			Path: "../../../example/local_secrets_example.unknown",
		},
	}

	e := common.SecretFileFormatError("unknown")

	if _, err := sm.GetSecret(context.TODO()); err == nil || !errors.Is(err, e) {
		t.Fatalf("expect %v got %v", e, err)
	}
}

func TestGetSecretFromYAML(t *testing.T) {
	t.Parallel()

	provider := SecretsProvider{
		config: &common.SecretsConfigLocal{
			Path: "../../../example/local_secrets_example.yaml",
		},
	}

	{
		secretValue, err := provider.GetSecret(context.TODO())
		if err != nil {
			t.Fatalf("expect nil got %v", err)
		}

		if len(secretValue) != 14 {
			t.Fatalf("expect 14 got %v", len(secretValue))
		}
	}
}

func TestGetSecretFromJSON(t *testing.T) {
	t.Parallel()

	provider := SecretsProvider{
		config: &common.SecretsConfigLocal{
			Path: "../../../example/local_secrets_example.json",
		},
	}

	{
		secretValue, err := provider.GetSecret(context.TODO())
		if err != nil {
			t.Fatalf("expect nil got %v", err)
		}

		if len(secretValue) != 14 {
			t.Fatalf("expect 14 got %v", len(secretValue))
		}
	}
}

func TestGetSecretFromTOML(t *testing.T) {
	t.Parallel()

	provider := SecretsProvider{
		config: &common.SecretsConfigLocal{
			Path: "../../../example/local_secrets_example.toml",
		},
	}

	{
		secretValue, err := provider.GetSecret(context.TODO())
		if err != nil {
			t.Fatalf("expect nil got %v", err)
		}

		if len(secretValue) != 14 {
			t.Fatalf("expect 14 got %v", len(secretValue))
		}
	}
}

func TestGetSecretFromFileWithoutExtension(t *testing.T) {
	t.Parallel()

	provider := SecretsProvider{
		config: &common.SecretsConfigLocal{
			Path: "../../../example/local_secrets_example_wo_ext",
		},
	}

	if _, err := provider.GetSecret(context.TODO()); err == nil {
		t.Fatalf("expect err")
	}
}
