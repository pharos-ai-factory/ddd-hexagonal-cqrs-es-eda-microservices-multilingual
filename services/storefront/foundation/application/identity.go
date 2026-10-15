package application

import (
	"crypto/sha256"
	"fmt"
)

// DerivedID gives an idempotent platform operation a stable UUIDv8 identity.
func DerivedID(purpose, key string) string {
	sum := sha256.Sum256([]byte("cafe-reference/v1\x00" + purpose + "\x00" + key))
	sum[6] = (sum[6] & 15) | 128
	sum[8] = (sum[8] & 63) | 128
	return fmt.Sprintf("%x-%x-%x-%x-%x", sum[:4], sum[4:6], sum[6:8], sum[8:10], sum[10:16])
}
