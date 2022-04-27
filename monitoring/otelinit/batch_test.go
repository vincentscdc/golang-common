package otelinit

import (
	"testing"
)

func TestWithBatchSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		size int
	}{
		{
			name: "default batch size",
			size: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &provider{}
			_ = WithBatchSize(tt.size)(p)
			if len(p.batchProcessorOptions) != 1 {
				t.Errorf("WithBatchSize() did not add a export batch size")
			}
		})
	}
}
