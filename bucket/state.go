package bucket

import "time"

type State struct {
	Tokens     int
	LastRefill time.Time
}
