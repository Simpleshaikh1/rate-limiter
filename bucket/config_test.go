package bucket

import (
	"errors"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name           string
		capacity       int
		refillInterval time.Duration
		wantErr        error
	}{
		{
			name:           "valid configuration",
			capacity:       10,
			refillInterval: time.Second,
		},
		{
			name:           "zero capacity",
			capacity:       0,
			refillInterval: time.Second,
			wantErr:        ErrInvalidCapacity,
		},
		{
			name:           "negative capacity",
			capacity:       -1,
			refillInterval: time.Second,
			wantErr:        ErrInvalidCapacity,
		},
		{
			name:           "zero refill interval",
			capacity:       10,
			refillInterval: 0,
			wantErr:        ErrInvalidRefillInterval,
		},
		{
			name:           "negative refill interval",
			capacity:       10,
			refillInterval: -time.Second,
			wantErr:        ErrInvalidRefillInterval,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := NewConfig(tc.capacity, tc.refillInterval)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			}

			if tc.wantErr != nil {
				return
			}

			gotCapacity := cfg.Capacity()
			if gotCapacity != tc.capacity {
				t.Fatalf("capacity mismatch: got %d want %d",
					cfg.Capacity(), tc.capacity)
			}

			if cfg.RefillInterval() != tc.refillInterval {
				t.Fatalf("interval mismatch: got %v want %v",
					cfg.RefillInterval(), tc.refillInterval)
			}
		})
	}
}
