package awssm

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/monacohq/golang-common/config/secrets/common"
)

type mockGetSecretValue func(ctx context.Context,
	params *secretsmanager.GetSecretValueInput,
	optFns ...func(*secretsmanager.Options),
) (*secretsmanager.GetSecretValueOutput, error)

func (m mockGetSecretValue) GetSecretValue(ctx context.Context,
	params *secretsmanager.GetSecretValueInput,
	optFns ...func(*secretsmanager.Options),
) (*secretsmanager.GetSecretValueOutput, error) {
	return m(ctx, params, optFns...)
}

type ClientImplMock struct {
	getSecretValueFn func(ctx context.Context,
		api SecretsManagerGetSecretValueAPI,
		secretID string,
	) (*string, error)
	unmarshalFn         func(data []byte, v any) error
	loadDefaultConfigFn func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error)
}

func (c *ClientImplMock) getSecretValue(ctx context.Context,
	api SecretsManagerGetSecretValueAPI,
	secretID string,
) (*string, error) {
	if c != nil && c.getSecretValueFn != nil {
		return c.getSecretValueFn(context.TODO(), api, secretID)
	}

	secMock := `{"k1":"v1", "k2":"v2"}`

	return &secMock, nil
}

func (c *ClientImplMock) Unmarshal(data []byte, v any) error {
	if c != nil && c.unmarshalFn != nil {
		return c.unmarshalFn(data, &v)
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("unmarshal aws secrets error: %w", err)
	}

	return nil
}

func (c *ClientImplMock) LoadDefaultConfig(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	if c != nil && c.loadDefaultConfigFn != nil {
		return c.loadDefaultConfigFn(ctx, optFns...)
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return cfg, fmt.Errorf("load default config error %w", err)
	}

	return cfg, nil
}

func TestGetSecretValue(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		client        func(t *testing.T) SecretsManagerGetSecretValueAPI
		secretID      string
		expect        string
		secretsGetter func(t *testing.T) Client
	}{
		{
			name: "HappyPath1",
			client: func(t *testing.T) SecretsManagerGetSecretValueAPI {
				t.Helper()

				return mockGetSecretValue(func(ctx context.Context,
					params *secretsmanager.GetSecretValueInput,
					optFns ...func(*secretsmanager.Options),
				) (*secretsmanager.GetSecretValueOutput, error) {
					t.Helper()

					if params.SecretId == nil {
						t.Fatalf("expect SecretId to not be nil")
					}

					return &secretsmanager.GetSecretValueOutput{
						SecretString: aws.String("test value"),
					}, nil
				})
			},
			secretID: "test key",
			expect:   "test value",
			secretsGetter: func(t *testing.T) Client {
				t.Helper()

				return &ClientImpl{}
			},
		},
		{
			name: "HappyPath2",
			client: func(t *testing.T) SecretsManagerGetSecretValueAPI {
				t.Helper()

				return mockGetSecretValue(func(ctx context.Context,
					params *secretsmanager.GetSecretValueInput,
					optFns ...func(*secretsmanager.Options),
				) (*secretsmanager.GetSecretValueOutput, error) {
					return &secretsmanager.GetSecretValueOutput{
						SecretString: aws.String(""),
					}, fmt.Errorf("test aws get secret value error")
				})
			},
			secretID: "test key",
			expect:   "",
			secretsGetter: func(t *testing.T) Client {
				t.Helper()

				return &ClientImpl{}
			},
		},
	}

	for _, tt := range cases {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			secret, _ := tt.secretsGetter(t).getSecretValue(context.TODO(), tt.client(t), tt.secretID)
			if secret != nil && tt.expect != *secret {
				t.Fatalf("expect %v, got %v", tt.expect, secret)
			}
		})
	}
}

func TestNewFromConfig(t *testing.T) {
	t.Parallel()

	awsSecretsConfigTest1 := &common.SecretsConfigAWS{}

	awsSecretsConfigInvalidTest2 := common.SecretsConfigAWS{}
	cases := []struct {
		name           string
		config         common.SecretsConfig
		expectedConfig common.SecretsConfig
		client         func(t *testing.T) *ClientImplMock
	}{
		{
			name:           "configTest1",
			config:         awsSecretsConfigTest1,
			expectedConfig: awsSecretsConfigTest1,
			client: func(t *testing.T) *ClientImplMock {
				t.Helper()

				return &ClientImplMock{
					loadDefaultConfigFn: func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
						cfg, err := config.LoadDefaultConfig(ctx, optFns...)
						if err != nil {
							return cfg, fmt.Errorf("load default config error %w", err)
						}

						return cfg, nil
					},
				}
			},
		},
		{
			name:           "configTest2",
			config:         awsSecretsConfigInvalidTest2,
			expectedConfig: nil,
			client: func(t *testing.T) *ClientImplMock {
				t.Helper()

				return &ClientImplMock{
					loadDefaultConfigFn: func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
						cfg, err := config.LoadDefaultConfig(ctx, optFns...)
						if err != nil {
							return cfg, fmt.Errorf("load default config error %w", err)
						}

						return cfg, nil
					},
				}
			},
		},
		{
			name:           "configTest3",
			config:         awsSecretsConfigTest1,
			expectedConfig: nil,
			client: func(t *testing.T) *ClientImplMock {
				t.Helper()

				return &ClientImplMock{
					loadDefaultConfigFn: func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
						return aws.Config{}, fmt.Errorf("load default config error")
					},
				}
			},
		},
	}

	for _, tC := range cases {
		tC := tC

		t.Run(tC.name, func(t *testing.T) {
			t.Parallel()

			testConfig := NewFromConfig(tC.config, tC.client(t))
			if tC.expectedConfig == nil && testConfig != nil {
				t.Fatalf("expect %v, got %v", tC.expectedConfig, testConfig)
			}
			if (tC.expectedConfig != nil && testConfig != nil) && (tC.expectedConfig != testConfig.config) {
				t.Fatalf("expect %v, got %v", tC.expectedConfig, testConfig)
			}
		})
	}
}

func TestGetSecret(t *testing.T) {
	t.Parallel()

	secIDTest1 := "secret0"
	secIDTest1Values := `{"k3":"v3", "k4":"v4"}`
	awsSecretsConfigTest1 := &common.SecretsConfigAWS{
		SecretID: secIDTest1,
	}

	var secretIDTest1ValuesOb map[string]any

	if err := json.Unmarshal([]byte(secIDTest1Values), &secretIDTest1ValuesOb); err != nil {
		panic(err)
	}

	cases := []struct {
		name                      string
		client                    func(t *testing.T) *ClientImplMock
		expectedSecretValues      map[string]any
		expectGetSecretValueError bool
	}{
		{
			name: "getSecretTest1",
			client: func(t *testing.T) *ClientImplMock {
				t.Helper()

				return &ClientImplMock{
					getSecretValueFn: func(ctx context.Context, api SecretsManagerGetSecretValueAPI, secretID string) (*string, error) {
						secretMock := secIDTest1Values

						return &secretMock, nil
					},
					loadDefaultConfigFn: func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
						cfg, err := config.LoadDefaultConfig(ctx, optFns...)
						if err != nil {
							return cfg, fmt.Errorf("load default config error %w", err)
						}

						return cfg, nil
					},
				}
			},
			expectedSecretValues:      secretIDTest1ValuesOb,
			expectGetSecretValueError: false,
		},
		{
			name: "getSecretTest2",
			client: func(t *testing.T) *ClientImplMock {
				t.Helper()

				return &ClientImplMock{
					getSecretValueFn: func(ctx context.Context, api SecretsManagerGetSecretValueAPI, secretID string) (*string, error) {
						return nil, fmt.Errorf("test get secret value error")
					},
					loadDefaultConfigFn: func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
						cfg, err := config.LoadDefaultConfig(ctx, optFns...)
						if err != nil {
							return cfg, fmt.Errorf("load default config error %w", err)
						}

						return cfg, nil
					},
				}
			},
			expectedSecretValues:      nil,
			expectGetSecretValueError: true,
		},
		{
			name: "getSecretTest3",
			client: func(t *testing.T) *ClientImplMock {
				t.Helper()

				return &ClientImplMock{
					getSecretValueFn: func(ctx context.Context, api SecretsManagerGetSecretValueAPI, secretID string) (*string, error) {
						secretMock := secIDTest1Values

						return &secretMock, nil
					},
					unmarshalFn: func(data []byte, v any) error {
						return fmt.Errorf("test unmarshal error")
					},
					loadDefaultConfigFn: func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
						cfg, err := config.LoadDefaultConfig(ctx, optFns...)
						if err != nil {
							return cfg, fmt.Errorf("load default config error %w", err)
						}

						return cfg, nil
					},
				}
			},
			expectedSecretValues:      nil,
			expectGetSecretValueError: true,
		},
	}

	for _, tC := range cases {
		tC := tC

		t.Run(tC.name, func(t *testing.T) {
			t.Parallel()

			testSecretsProvider := NewFromConfig(awsSecretsConfigTest1, tC.client(t))
			secretValues, err := testSecretsProvider.GetSecret()
			if (tC.expectGetSecretValueError == false && err == nil) && (!reflect.DeepEqual(tC.expectedSecretValues, secretValues)) {
				t.Fatalf("expect %v, got %v", tC.expectedSecretValues, secretValues)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	awsSecretsConfigTest1 := &common.SecretsConfigAWS{}
	secIDTest1Values := `{"k3":"v3", "k4":"v4"}`
	secIDTest1InvalidValues := `abc`

	var secretIDTest1ValuesOb map[string]any

	if err := json.Unmarshal([]byte(secIDTest1Values), &secretIDTest1ValuesOb); err != nil {
		panic(err)
	}

	cases := []struct {
		name        string
		expectError bool
		input       string
	}{
		{
			name:        "unmarshalTest1",
			input:       secIDTest1Values,
			expectError: false,
		},
		{
			name:        "unmarshalTest2",
			input:       secIDTest1InvalidValues,
			expectError: true,
		},
	}

	for _, tC := range cases {
		tC := tC

		t.Run(tC.name, func(t *testing.T) {
			t.Parallel()

			testSecretsProvider := NewFromConfig(awsSecretsConfigTest1, ClientImpl{})
			var output map[string]any
			err := testSecretsProvider.Unmarshal([]byte(tC.input), &output)

			if tC.expectError == false && err != nil {
				t.Fatalf("expect %v, got %v", tC.expectError, err)
			}
		})
	}
}

func TestLoadDefaultConfig(t *testing.T) {
	t.Parallel()

	awsSecretsConfigTest1 := &common.SecretsConfigAWS{}

	cases := []struct {
		name        string
		expectError bool
		input       func(*config.LoadOptions) error
	}{
		{
			name: "loadDefaultConfigTest1",
			input: func(lo *config.LoadOptions) error {
				return fmt.Errorf("test error")
			},
			expectError: true,
		},
		{
			name: "loadDefaultConfigTest2",
			input: func(lo *config.LoadOptions) error {
				return nil
			},
			expectError: false,
		},
	}

	for _, tC := range cases {
		tC := tC

		t.Run(tC.name, func(t *testing.T) {
			t.Parallel()

			testSecretsProvider := NewFromConfig(awsSecretsConfigTest1, ClientImpl{})

			_, err := testSecretsProvider.LoadDefaultConfig(context.TODO(), tC.input)

			if tC.expectError == false && err != nil {
				t.Fatalf("expect %v, got %v", tC.expectError, err)
			}
			if tC.expectError == true && err == nil {
				t.Fatalf("expect %v, got %v", tC.expectError, err)
			}
		})
	}
}
