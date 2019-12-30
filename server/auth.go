package server

import (
	"context"
	"errors"
	"log"

	authn "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/pborman/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// NewAuthFunc adapts Auth as gRPC middleware
func NewAuthFunc(auth Auth) authn.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		token, err := authn.AuthFromMD(ctx, auth.Scheme())
		if err != nil {
			return nil, err
		}
		username, err := auth.Authenticate(token)
		if err != nil {
			errorID := uuid.New()
			log.Printf("auth failed - error id: %s, error: %v\n", errorID, err)
			return nil, status.Error(codes.PermissionDenied, errorID)
		}
		return WithUserName(ctx, username), nil
	}
}
