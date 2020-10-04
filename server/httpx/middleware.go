package httpx

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/pborman/uuid"

	"github.com/tomcz/example-grpc/server"
)

func authMiddleware(auth server.TokenAuth, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := server.UserName(r.Context())
		if username != "" {
			// already authenticated
			next.ServeHTTP(w, r)
			return
		}
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}
		pair := strings.SplitN(header, " ", 2)
		if len(pair) != 2 {
			http.Error(w, "Bad Authorization header", http.StatusUnauthorized)
			return
		}
		if strings.ToLower(pair[0]) != strings.ToLower(auth.Scheme()) {
			http.Error(w, "Unsupported Authorization scheme", http.StatusUnauthorized)
			return
		}
		var err error
		username, err = auth.Authenticate(pair[1])
		if err != nil {
			authFailed(w, err)
			return
		}
		r = r.WithContext(server.WithUserName(r.Context(), username))
		next.ServeHTTP(w, r)
	})
}

func mtlsMiddleware(mtls server.AllowList, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		certs := r.TLS.PeerCertificates
		if len(certs) > 0 {
			cert := certs[0]
			username, err := mtls.Allow(cert)
			if err != nil {
				authFailed(w, err)
				return
			}
			if username != "" {
				r = r.WithContext(server.WithUserName(r.Context(), username))
			}
		}
		next.ServeHTTP(w, r)
	})
}

func authFailed(w http.ResponseWriter, err error) {
	errorID := uuid.New()
	log.Printf("auth failed - error id: %s, error: %v\n", errorID, err)
	http.Error(w, fmt.Sprintf("Authorization failed: %s", errorID), http.StatusForbidden)
}
