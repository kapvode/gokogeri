package redisutil

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

// DoMany sends buffered commands and returns their replies. If one of the replies is an error, it returns it instead.
//
// The command count is used only to verify that the number of replies matches expectations.
//
// See also: CheckReplies.
func DoMany(conn redis.Conn, commandCount int) ([]interface{}, error) {
	replies, err := redis.Values(conn.Do(""))
	if err != nil {
		return nil, err
	}

	err = CheckReplies(replies, commandCount)
	if err != nil {
		return nil, err
	}

	return replies, nil
}

// CheckReplies returns an error if the number of replies does not match the expected number, and returns the first
// error among the replies, if any.
func CheckReplies(replies []interface{}, expect int) error {
	if len(replies) != expect {
		return fmt.Errorf("expect replies: %d, got %d", expect, len(replies))
	}
	for i, r := range replies {
		if err, ok := r.(redis.Error); ok {
			return fmt.Errorf("reply %d: %v", i, err)
		}
	}
	return nil
}
