package secrets

import (
	"reflect"
	"testing"

	"github.com/monacohq/golang-common/config/secrets/common"
)

func BenchmarkGetLocalSecrets(b *testing.B) {
	localConfig := &common.SecretsConfigLocal{
		Path: "example/local_secrets_example.yaml",
	}

	provider, err := NewSecretUrnFromConfig(localConfig)
	if err != nil {
		return
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		provider.GetSecretString("item_string")
	}
}

func TestGetSecretType(t *testing.T) {
	t.Parallel()

	localConfig := &common.SecretsConfigLocal{
		Path: "example/local_secrets_example.yaml",
	}

	provider, err := NewSecretUrnFromConfig(localConfig)
	if err != nil {
		t.Fatalf("expect nil got %v", err)
	}

	if val, err := provider.GetSecretBool("item_bool"); err != nil {
		t.Fatalf("expect %v got %v, err %v", true, val, err)
	}

	if val, err := provider.GetSecretFloat64("item_float64"); err != nil {
		t.Fatalf("expect %v got %v, err %v", 12.34, val, err)
	}

	if val, err := provider.GetSecretInt("item_int"); err != nil {
		t.Fatalf("expect %v got %v, err %v", 1234, val, err)
	}

	if val, err := provider.GetSecretIntSlice("item_intslice"); err != nil {
		t.Fatalf("expect %v got %v, err %v", []int{1, 2, 3, 4}, val, err)
	}

	if val, err := provider.GetSecretString("item_string"); err != nil {
		t.Fatalf("expect %v got %v, err %v", "1234", val, err)
	}

	if val, err := provider.GetSecretStringSlice("item_stringslice"); err != nil {
		t.Fatalf("expect %v got %v, err %v", []string{"1", "2", "3", "4"}, val, err)
	}
}

func TestGetSecretItemNotFound(t *testing.T) {
	t.Parallel()

	localConfig := &common.SecretsConfigLocal{
		Path: "example/local_secrets_example.yaml",
	}

	provider, err := NewSecretUrnFromConfig(localConfig)
	if err != nil {
		t.Fatalf("expect nil got %v", err)
	}

	if val, err := provider.GetSecretBool("item_bool_not_found"); err == nil {
		t.Fatalf("expect type err got %v", val)
	}

	if val, err := provider.GetSecretFloat64("item_float64_not_found"); err == nil {
		t.Fatalf("expect type err got %v", val)
	}

	if val, err := provider.GetSecretInt("item_int_not_found"); err == nil {
		t.Fatalf("expect type err got %v", val)
	}

	if val, err := provider.GetSecretIntSlice("item_intslice_not_found"); err == nil {
		t.Fatalf("expect type err got %v", val)
	}

	if val, err := provider.GetSecretString("item_string_not_found"); err == nil {
		t.Fatalf("expect type err got %v", val)
	}

	if val, err := provider.GetSecretStringSlice("item_stringslice_not_found"); err == nil {
		t.Fatalf("expect type err got %v", val)
	}

	if provider.IsSecretSet("oracle key") {
		t.Fatalf("expect a not found item")
	}
}

func TestGetSecretItemErrType(t *testing.T) {
	t.Parallel()

	localConfig := &common.SecretsConfigLocal{
		Path: "example/local_secrets_example.yaml",
	}

	provider, err := NewSecretUrnFromConfig(localConfig)
	if err != nil {
		t.Fatalf("expect nil got %v", err)
	}

	if val, err := provider.GetSecretBool("item_bool_err_type"); err == nil {
		t.Fatalf("expect type err got %v", val)
	}

	if val, err := provider.GetSecretFloat64("item_float64_err_type"); err == nil {
		t.Fatalf("expect type err got %v", val)
	}

	if val, err := provider.GetSecretInt("item_int_err_type"); err == nil {
		t.Fatalf("expect type err got %v", val)
	}

	if val, err := provider.GetSecretIntSlice("item_intslice_err_type"); err == nil {
		t.Fatalf("expect type err got %v", val)
	}

	if val, err := provider.GetSecretIntSlice("item_intslice_err_type_1"); err == nil {
		t.Fatalf("expect type err got %v", val)
	}

	if val, err := provider.GetSecretString("item_string_err_type"); err == nil {
		t.Fatalf("expect type err got %v", val)
	}

	if val, err := provider.GetSecretStringSlice("item_stringslice_err_type"); err == nil {
		t.Fatalf("expect type err got %v", val)
	}

	if val, err := provider.GetSecretStringSlice("item_stringslice_err_type_1"); err == nil {
		t.Fatalf("expect type err got %v", val)
	}
}

type testSecretsConfig struct{}

func (testSecretsConfig) Name() string {
	return ""
}

func TestNewSecretUrnFromConfig(t *testing.T) {
	t.Parallel()

	testSecID := "1234"

	cases := []struct {
		name        string
		inputConfig common.SecretsConfig
		expectError bool
	}{
		{
			name: "NewSecretUrnFromConfig1",
			inputConfig: &common.SecretsConfigLocal{
				Path: "example/local_secrets_example.yaml",
			},
			expectError: false,
		},
		{
			name: "NewSecretUrnFromConfig2",
			inputConfig: &common.SecretesConfigAWS{
				SecretID: testSecID,
			},
			expectError: true, // haven't mocked aws here so it will error trying to get a real secret
		},
		{
			name:        "NewSecretUrnFromConfig3",
			inputConfig: &testSecretsConfig{},
			expectError: true,
		},
	}

	for _, tC := range cases {
		tC := tC

		t.Run(tC.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewSecretUrnFromConfig(tC.inputConfig)

			if tC.expectError == false && err != nil {
				t.Fatalf("expect %v, got %v", tC.expectError, err)
			}
			if tC.expectError == true && err == nil {
				t.Fatalf("expect %v, got %v", tC.expectError, err)
			}
		})
	}
}

func TestSecretsBind(t *testing.T) {
	t.Parallel()

	type SampleStruct struct {
		ItemBool        bool     `secret_key:"item_bool"`
		ItemInt         int      `secret_key:"item_int"`
		ItemString      string   `secret_key:"item_string"`
		ItemFloat64     float64  `secret_key:"item_float64"`
		ItemIntSlice    []int    `secret_key:"item_intslice"`
		ItemStringSlice []string `secret_key:"item_stringslice"`
	}

	cases := []struct {
		name           string
		testStruct     interface{}
		expectError    bool
		expectedStruct *SampleStruct
	}{
		{
			name:        "non pointer struct",
			testStruct:  struct{}{},
			expectError: true,
		},
		{
			name:        "struct with different types",
			testStruct:  &SampleStruct{},
			expectError: false,
			expectedStruct: &SampleStruct{
				ItemBool:        true,
				ItemInt:         1234,
				ItemString:      "1234",
				ItemFloat64:     12.34,
				ItemIntSlice:    []int{1, 2, 3, 4},
				ItemStringSlice: []string{"1", "2", "3", "4"},
			},
		},
	}

	for _, tC := range cases {
		tC := tC

		t.Run(tC.name, func(t *testing.T) {
			t.Parallel()

			secretUrn := SecretUrn{
				"item_bool":                   true,
				"item_int":                    1234,
				"item_string":                 "1234",
				"item_float64":                12.34,
				"item_intslice":               []int{1, 2, 3, 4},
				"item_stringslice":            []string{"1", "2", "3", "4"},
				"item_bool_err_type":          "true",
				"item_int_err_type":           12.34,
				"item_string_err_type":        1234,
				"item_float64_err_type":       1234,
				"item_intslice_err_type":      []string{"1", "2", "3", "4"},
				"item_intslice_err_type_1":    1,
				"item_stringslice_err_type":   []int{1, 2, 3, 4},
				"item_stringslice_err_type_1": 1,
			}

			err := secretUrn.Bind(tC.testStruct)

			if tC.expectError && err == nil {
				t.Fatalf("expected error, got none")
			}

			if !tC.expectError && err != nil {
				t.Fatalf("didn't expect error, got %v", err)
			}

			if !tC.expectError {
				if !reflect.DeepEqual(tC.expectedStruct, tC.testStruct) {
					t.Fatalf("expect \n%v got \n%v", tC.expectedStruct, tC.testStruct)
				}
			}
		})
	}
}

func BenchmarkSecretsBind(b *testing.B) {
	localConfig := &common.SecretsConfigLocal{
		Path: "example/local_secrets_example.yaml",
	}

	secretUrn, err := NewSecretUrnFromConfig(localConfig)
	if err != nil {
		return
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = secretUrn.Bind(&struct {
			t string `secret_key:"t"`
		}{})
	}
}
