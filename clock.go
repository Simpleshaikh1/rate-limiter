package rate_limiter_token_bucket_leaky_bucket

import "time"

func Clock interface{
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}