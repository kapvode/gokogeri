package sidekiq

import (
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	t1 := time.Unix(1669852800, 0)
	t2 := ToTime(Time(t1))
	if !t1.Equal(t2) {
		t.Fatalf("want '%v', got '%v'", t1, t2)
	}
}
