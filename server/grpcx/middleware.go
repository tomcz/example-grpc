package grpcx

import (
	"context"

	mw "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/tomcz/example-grpc/server"
)

func authMiddleware(authFunc mw.AuthFunc) []grpc.ServerOption {
	return []grpc.ServerOption{
		// echo service only has unary endpoints, but ...
		grpc.UnaryInterceptor(mw.UnaryServerInterceptor(authFunc)),
		// grpcurl uses a streaming endpoint for reflection,
		// so let's make sure the user is allowed to reflect
		grpc.StreamInterceptor(mw.StreamServerInterceptor(authFunc)),
	}
}

func newServerAuthFunc(auth server.TokenAuth) mw.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		token, err := mw.AuthFromMD(ctx, auth.Scheme())
		if err != nil {
			return nil, err
		}
		username, err := auth.Authenticate(token)
		if err != nil {
			return authFailed(err)
		}
		return server.WithUserName(ctx, username), nil
	}
}

func newMTLSAuthFunc(mtls server.AllowList, next mw.AuthFunc) mw.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		if p, ok := peer.FromContext(ctx); ok {
			if tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo); ok {
				certs := tlsInfo.State.PeerCertificates
				if len(certs) > 0 {
					// we want the first cert in the chain as that is the actual client cert
					username, err := mtls.Allow(certs[0])
					if err != nil {
						return authFailed(err)
					}
					return server.WithUserName(ctx, username), nil
				}
			}
		}
		return next(ctx)
	}
}

func authFailed(err error) (context.Context, error) {
	errorID := server.ErrorID()
	log.WithError(err).WithField("error_id", errorID).Warn("auth failed")
	return nil, status.Errorf(codes.PermissionDenied, "error_id: %s", errorID)
}
