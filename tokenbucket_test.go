package rate_limiter_token_bucket_leaky_bucket

import "testing"

func TestNewTokenBucket(t *testing.T) {
	tests := []struct {
		name       string
		capacity   int
		refillRate int
		wantErr    bool
	}{
		{
			name:       "valid configuration",
			capacity:   10,
			refillRate: 5,
			wantErr:    false,
		},
		{
			name:       "zero capacity",
			capacity:   0,
			refillRate: 5,
			wantErr:    true,
		},
		{
			name:       "negative capacity",
			capacity:   -1,
			refillRate: 5,
			wantErr:    true,
		},
		{
			name:       "zero refill rate",
			capacity:   10,
			refillRate: 0,
			wantErr:    true,
		},
		{
			name:       "negative refill rate",
			capacity:   10,
			refillRate: -5,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bucket, err := NewTokenBucket(tc.capacity, tc.refillRate)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error but got nil")
				}
				if bucket != nil {
					t.Fatalf("expected nil bucket on error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if bucket.capacity != tc.capacity {
				t.Fatalf("capacity = %d, want %d", bucket.capacity, tc.capacity)
			}

			if bucket.tokens != tc.capacity {
				t.Fatalf("tokens = %d, want %d", bucket.tokens, tc.capacity)
			}

			if bucket.refillRate != tc.refillRate {
				t.Fatalf("refillRate = %d, want %d", bucket.refillRate, tc.refillRate)
			}

			if bucket.lastRefill.IsZero() {
				t.Fatalf("lastRefill should be initialized")
			}
		})
	}
}
