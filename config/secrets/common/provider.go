package common

import "context"

type Provider interface {
	GetSecret(ctx context.Context) (map[string]any, error)
}
