package server

import (
	"context"
	"log"

	authn "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/pborman/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

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

// NewMTLSAuthFunc allows for optional client authentication via mTLS
func NewMTLSAuthFunc(next authn.AuthFunc) authn.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		username := UserName(ctx)
		if username != "" {
			// already authenticated
			return ctx, nil
		}
		if p, ok := peer.FromContext(ctx); ok {
			if tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo); ok {
				certs := tlsInfo.State.PeerCertificates
				if len(certs) > 0 {
					dnsNames := certs[0].DNSNames
					if len(dnsNames) > 0 {
						username = dnsNames[0]
						return WithUserName(ctx, username), nil
					}
				}
			}
		}
		return next(ctx)
	}
}
