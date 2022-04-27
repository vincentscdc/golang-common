package otelinit

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel/sdk/resource"
)

func Test_newProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options []func(*provider) error
		wantErr bool
	}{
		{
			name:    "default local provider with stdout",
			options: nil,
			wantErr: false,
		},
		{
			name: "problematic option",
			options: []func(*provider) error{
				func(pvd *provider) error { return errors.New("error") },
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := newProvider("test", tt.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProvider() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("NewProvider() no provider returned")
			}
		})
	}
}

type testBadDetector struct {
	schemaURL string
}

func (tbd *testBadDetector) Detect(ctx context.Context) (*resource.Resource, error) {
	return resource.NewWithAttributes(tbd.schemaURL), nil
}

func TestProvider_init(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options []func(*provider) error
		wantErr bool
	}{
		{
			name:    "default provider with traces to stdout",
			options: nil,
			wantErr: false,
		},
		{
			name: "problematic resource option",
			options: []func(*provider) error{
				func(pvd *provider) error {
					pvd.resourceOptions = append(
						pvd.resourceOptions,
						resource.WithDetectors(&testBadDetector{schemaURL: "https://opentelemetry.io/schemas/1.4.0"}),
					)
					pvd.resourceOptions = append(
						pvd.resourceOptions,
						resource.WithDetectors(&testBadDetector{schemaURL: "https://opentelemetry.io/schemas/1.3.0"}),
					)

					return nil
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			w := &bytes.Buffer{}

			tt.options = append(tt.options, WithWriterTraceExporter(w))

			pvd, err := newProvider("test", tt.options...)
			if err != nil {
				t.Errorf("newProvider() error = %v", err)

				return
			}

			sd, err := pvd.init(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("init() error = %v when expected err %t", err, tt.wantErr)

				return
			}

			if sd != nil {
				_ = sd()
			}
		})
	}
}
