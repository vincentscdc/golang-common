// This package allows you to init and enable tracing in your app
package otelinit

import (
	"context"
	"fmt"
)

func InitProvider(
	ctx context.Context,
	serviceName string,
	options ...func(*provider) error,
) (func() error, error) {
	pvd, err := newProvider(serviceName, options...)
	if err != nil {
		return nil, fmt.Errorf("newProvider() error = %w", err)
	}

	shutdown, err := pvd.init(ctx)
	if err != nil {
		return nil, fmt.Errorf("init() error = %w", err)
	}

	return shutdown, nil
}
