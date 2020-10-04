package grpcx

import (
	"context"
	"log"

	mw "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/pborman/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/tomcz/example-grpc/server"
)

func authMiddleware(authFunc mw.AuthFunc) []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.UnaryInterceptor(mw.UnaryServerInterceptor(authFunc)),
		grpc.StreamInterceptor(mw.StreamServerInterceptor(authFunc)),
	}
}

func newServerAuthFunc(auth server.Auth) mw.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		token, err := mw.AuthFromMD(ctx, auth.Scheme())
		if err != nil {
			return nil, err
		}
		username, err := auth.Authenticate(token)
		if err != nil {
			errorID := uuid.New()
			log.Printf("auth failed - error id: %s, error: %v\n", errorID, err)
			return nil, status.Error(codes.PermissionDenied, errorID)
		}
		return server.WithUserName(ctx, username), nil
	}
}

func newMTLSAuthFunc(next mw.AuthFunc) mw.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		if p, ok := peer.FromContext(ctx); ok {
			if tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo); ok {
				certs := tlsInfo.State.PeerCertificates
				if len(certs) > 0 {
					dnsNames := certs[0].DNSNames
					if len(dnsNames) > 0 {
						username := dnsNames[0]
						return server.WithUserName(ctx, username), nil
					}
				}
			}
		}
		return next(ctx)
	}
}
