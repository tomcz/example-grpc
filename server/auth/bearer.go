package auth

import (
	"strings"

	"github.com/tomcz/example-grpc/server"
)

type bearerAuth struct {
	tokens map[string]string
}

// NewBearerAuth represents bearer token authentication
func NewBearerAuth(tokens string) server.Auth {
	tokenMap := make(map[string]string)
	for _, tok := range strings.Split(tokens, ",") {
		pair := strings.SplitN(tok, ":", 2)
		tokenMap[pair[1]] = pair[0]
	}
	return &bearerAuth{
		tokens: tokenMap,
	}
}

func (b *bearerAuth) Scheme() string {
	return "bearer"
}

func (b *bearerAuth) Authenticate(token string) (string, error) {
	if username, ok := b.tokens[token]; ok {
		return username, nil
	}
	return "", server.ErrInvalidToken
}
