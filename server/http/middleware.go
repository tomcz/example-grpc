package http

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/tomcz/example-grpc/server"
)

// AuthMiddleware optional authentication middleware
func AuthMiddleware(auth server.Auth) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			username, err := auth.Authenticate(pair[1])
			if err != nil {
				errorID := uuid.New().String()
				log.Printf("auth failed - error id: %s, error: %v\n", errorID, err)
				http.Error(w, fmt.Sprintf("Authorization failed: %s", errorID), http.StatusForbidden)
				return
			}
			r = r.WithContext(server.WithUserName(r.Context(), username))
			next.ServeHTTP(w, r)
		})
	}
}
