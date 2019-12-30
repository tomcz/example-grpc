package server

import (
	"context"
	"errors"
)

// ErrInvalidToken authentication failure
var ErrInvalidToken = errors.New("invalid token")

type contextKey int

const (
	usernameKey contextKey = iota
)

// Auth represents way of resolving tokens to usernames
type Auth interface {
	Scheme() string
	Authenticate(token string) (username string, err error)
}

// WithUserName store the username under a well-known context key
func WithUserName(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameKey, username)
}

// UserName retrieves the existing username, or returns an empty string
func UserName(ctx context.Context) string {
	if username, ok := ctx.Value(usernameKey).(string); ok {
		return username
	}
	return ""
}
