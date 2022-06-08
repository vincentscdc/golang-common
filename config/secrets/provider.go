package secrets

type Provider interface {
	GetSecret() (map[string]any, error)
}
