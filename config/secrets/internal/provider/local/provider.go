package local

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/monacohq/golang-common/config/secrets/common"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type SecretsProvider struct {
	config *common.SecretsConfigLocal
}

func NewFromConfig(config common.SecretsConfig) *SecretsProvider {
	localConfig, ok := config.(*common.SecretsConfigLocal)
	if !ok {
		return nil
	}

	return &SecretsProvider{
		config: localConfig,
	}
}

func (p *SecretsProvider) GetSecret() (map[string]any, error) {
	return readSecretsConfig(p.config.Path)
}

func readSecretsConfig(filePath string) (map[string]any, error) {
	configFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file err: %w", err)
	}

	defer configFile.Close()

	fileFormat := filepath.Ext(filePath)
	if fileFormat == "" {
		return nil, common.SecretFileFormatError("filename without extension")
	}

	fileFormat = fileFormat[1:] // remove the dot prefix in filename extension
	switch fileFormat {
	case common.SecretsYAML:
		return decodeConfig(yaml.NewDecoder(configFile))
	case common.SecretsJSON:
		return decodeConfig(json.NewDecoder(configFile))
	case common.SecretsTOML:
		return decodeConfig(toml.NewDecoder(configFile))
	}

	return nil, common.SecretFileFormatError(fileFormat)
}

type decoder interface {
	Decode(v any) error
}

func decodeConfig(d decoder) (map[string]any, error) {
	var m map[string]any
	if err := d.Decode(&m); err != nil {
		return nil, fmt.Errorf("parse file err: %w", err)
	}

	return m, nil
}
