package sidekiq

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// JobID gets 12 bytes from a cryptographically secure random number generator and returns them as a string in
// hexadecimal encoding.
func JobID() (string, error) {
	b := make([]byte, 12)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("rand.Read: %v", err)
	}
	return hex.EncodeToString(b), nil
}
