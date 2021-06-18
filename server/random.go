package server

import (
	"math/rand"
	"strings"
	"time"
)

// From https://www.grc.com/ppp.htm
const letters = "ABCDEFGHJKLMNPRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func ErrorID() string {
	// short enough to remember
	// long enough to be unique-ish
	return RandString(7)
}

func RandString(length int) string {
	var sb strings.Builder
	sb.Grow(length)
	for i := 0; i < length; i++ {
		sb.WriteByte(letters[rand.Intn(len(letters))])
	}
	return sb.String()
}
