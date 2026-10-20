package tokenhash

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256Hex(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
