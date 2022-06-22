package awssm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/monacohq/golang-common/config/secrets/common"
)

type JSONClient interface {
	Unmarshal(data []byte, v any) error
}

type AWSConfigClient interface {
	LoadDefaultConfig(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error)
}

type SecretsClient interface {
	getSecretValue(ctx context.Context,
		api SecretsManagerGetSecretValueAPI,
		secretID string,
	) (*string, error)
}

type Client interface {
	SecretsClient
	JSONClient
	AWSConfigClient
}

type ClientImpl struct{}

type SecretsProvider struct {
	config    *common.SecretsConfigAWS
	awsConfig aws.Config
	Client
}

var _ common.Provider = (*SecretsProvider)(nil)

func NewFromConfig(ctx context.Context, sconfig common.SecretsConfig, client Client) *SecretsProvider {
	scAws, ok := sconfig.(*common.SecretsConfigAWS)
	if !ok {
		return nil
	}

	awsConfig, err := client.LoadDefaultConfig(ctx, config.WithRegion(scAws.Region))
	if err != nil {
		return nil
	}

	secretManager := &SecretsProvider{
		config:    scAws,
		awsConfig: awsConfig,
		Client:    client,
	}

	return secretManager
}

func (p *SecretsProvider) GetSecret(ctx context.Context) (map[string]any, error) {
	client := secretsmanager.NewFromConfig(p.awsConfig)
	secretID := p.config.SecretID

	var err error

	result, err := p.getSecretValue(context.TODO(), client, secretID)
	if err != nil {
		return nil, err
	}

	var secret map[string]any
	if err := p.Unmarshal([]byte(*result), &secret); err != nil {
		return nil, fmt.Errorf("unmarshal aws secrets error: %w", err)
	}

	return secret, nil
}

type SecretsManagerGetSecretValueAPI interface {
	GetSecretValue(ctx context.Context,
		params *secretsmanager.GetSecretValueInput,
		optFns ...func(*secretsmanager.Options),
	) (*secretsmanager.GetSecretValueOutput, error)
}

func (c ClientImpl) getSecretValue(ctx context.Context,
	api SecretsManagerGetSecretValueAPI,
	secretID string,
) (*string, error) {
	var output *secretsmanager.GetSecretValueOutput

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	}

	output, err := api.GetSecretValue(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("call GetSecretValue api error: %w", err)
	}

	return output.SecretString, nil
}

func (c ClientImpl) Unmarshal(data []byte, v any) error {
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("unmarshal aws secrets error: %w", err)
	}

	return nil
}

func (c ClientImpl) LoadDefaultConfig(ctx context.Context,
	optFns ...func(*config.LoadOptions) error,
) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return cfg, fmt.Errorf("load default config error %w", err)
	}

	return cfg, nil
}
