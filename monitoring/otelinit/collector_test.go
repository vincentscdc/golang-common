package otelinit

import (
	"context"
	"testing"
)

func TestWithGRPCTraceExporter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name: "happy path",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			p := &provider{}

			if err := WithGRPCTraceExporter(ctx, "dwd")(p); err != nil {
				t.Errorf("WithGRPCTraceExporter() expected error %t got %v", tt.wantErr, err)
			}

			if !tt.wantErr {
				sd, err := p.init(context.Background())
				if err != nil {
					t.Errorf("WithGRPCTraceExporter() expected to init but got %v", err)
				}

				_ = sd()
			}
		})
	}
}
