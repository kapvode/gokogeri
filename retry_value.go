package gokogeri

import (
	"encoding/json"
	"fmt"
	"strconv"
)

var jsonFalse = []byte{'f', 'a', 'l', 's', 'e'}
var jsonTrue = []byte{'t', 'r', 'u', 'e'}

type retryValue struct {
	ok    bool
	times int
}

// MarshalJSON implements json.Marshaler.
func (r retryValue) MarshalJSON() ([]byte, error) {
	if !r.ok {
		return jsonFalse, nil
	}
	if r.times > 0 {
		return []byte(strconv.FormatInt(int64(r.times), 10)), nil
	}
	return jsonTrue, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *retryValue) UnmarshalJSON(b []byte) error {
	if isJSONTrue(b) {
		r.ok = true
		return nil
	}

	if isJSONFalse(b) {
		r.ok = false
		return nil
	}

	var n int
	err := json.Unmarshal(b, &n)
	if err != nil {
		return fmt.Errorf("decoding job retry value: %v", err)
	}

	r.ok = n > 0
	r.times = n
	return nil
}

func isJSONFalse(b []byte) bool {
	return len(b) == 5 && b[0] == 'f' && b[1] == 'a' && b[2] == 'l' && b[3] == 's' && b[4] == 'e'
}

func isJSONTrue(b []byte) bool {
	return len(b) == 4 && b[0] == 't' && b[1] == 'r' && b[2] == 'u' && b[3] == 'e'
}
