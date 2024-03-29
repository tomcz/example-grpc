package server

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"strings"

	"github.com/tomcz/gotools/maps"
)

// ErrInvalidToken authentication failure
var ErrInvalidToken = errors.New("invalid token")

// ErrNoCertMatch authentication failure
var ErrNoCertMatch = errors.New("no certificate match")

type contextKey int

const (
	usernameKey contextKey = iota
)

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

// TokenAuth represents a way of resolving tokens to usernames.
type TokenAuth interface {
	Authenticate(token string) (username string, err error)
	Scheme() string
}

type bearerAuth struct {
	tokens map[string]string
}

// NewBearerAuth represents bearer token authentication.
func NewBearerAuth(tokens string) TokenAuth {
	tokenMap := make(map[string]string)
	for _, tok := range strings.Split(tokens, ",") {
		pair := strings.SplitN(tok, ":", 2)
		if len(pair) == 2 {
			tokenMap[pair[1]] = pair[0]
		}
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
	return "", ErrInvalidToken
}

// AllowList describes a list of allowed TLS certificates.
type AllowList interface {
	Allow(cert *x509.Certificate) (username string, err error)
	Enabled() bool
}

type domainAllowList struct {
	allowed map[string]bool
}

// NewDomainAllowList creates an allowed list from a comma-separated set of domains.
func NewDomainAllowList(domains string) AllowList {
	return &domainAllowList{
		allowed: maps.NewSet(strings.Split(domains, ",")...),
	}
}

func (d *domainAllowList) Allow(cert *x509.Certificate) (username string, err error) {
	// MAYBE: fail if the certificate has been revoked by the issuer
	cn := cert.Subject.CommonName
	if d.allowed[cn] {
		return cn, nil
	}
	for _, san := range cert.DNSNames {
		if d.allowed[san] {
			return san, nil
		}
	}
	return "", fmt.Errorf("%w - CN: %s", ErrNoCertMatch, cn)
}

func (d *domainAllowList) Enabled() bool {
	return len(d.allowed) != 0
}
