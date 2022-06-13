package common

type SecretsConfig interface {
	Name() string
}

// SecretsConfigLocal represents secrets stored in local file.
// The secretID is not necessary for LocalProvider since local secrets are stored in separated files.
type SecretsConfigLocal struct {
	Path string
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
	SecretID string
}

func (SecretsConfigAWS) Name() string {
	return "AWS secrets manager"
}
