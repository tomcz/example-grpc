package server

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

func ErrorID() string {
	// short enough to remember
	// long enough to be unique-ish
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "XXXX" // unlikely
	}
	return strings.ToUpper(hex.EncodeToString(buf))
}
