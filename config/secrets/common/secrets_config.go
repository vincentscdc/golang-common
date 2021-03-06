package common

type SecretsConfig interface {
	Name() string
}

// SecretsConfigLocal represents secrets stored in local file.
// The secretID is not necessary for LocalProvider since local secrets are stored in separated files.
type SecretsConfigLocal struct {
	Path string `yaml:"path" json:"path" toml:"path"`
}

func (SecretsConfigLocal) Name() string {
	return "local secrets"
}

const (
	SecretsYAML = "yaml"
	SecretsJSON = "json"
	SecretsTOML = "toml"
)

type SecretsConfigAWS struct {
	SecretID string `yaml:"secret_id" json:"secret_id" toml:"secret_id"`
	Region   string `yaml:"region" json:"region" toml:"region"`
}

func (SecretsConfigAWS) Name() string {
	return "AWS secrets manager"
}
