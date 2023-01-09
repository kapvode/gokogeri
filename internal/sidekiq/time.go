package sidekiq

import "time"

// Time converts a Go time value to a floating point value expected by Sidekiq (Time#to_f in Ruby).
func Time(t time.Time) float64 {
	return float64(t.UnixNano()) / 1e9
}

// ToTime converts a Ruby floating point time value, as used by Sidekiq, to a Go time value.
func ToTime(f float64) time.Time {
	return time.Unix(0, int64(f*1e9))
}
