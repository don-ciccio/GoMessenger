package auth

import (
	"context"
	"net/http"
)

type contextKey string

const UserIDKey contextKey = "userId"

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Try Authorization: Bearer header first (used by REST API calls)
		tokenString := ""
		if authHeader := r.Header.Get("Authorization"); authHeader != "" {
			const prefix = "Bearer "
			if len(authHeader) > len(prefix) && authHeader[:len(prefix)] == prefix {
				tokenString = authHeader[len(prefix):]
			}
		}

		// 2. Fall back to ?token= query parameter (used by WebSocket connections)
		if tokenString == "" {
			tokenString = r.URL.Query().Get("token")
		}

		if tokenString == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		claims, err := verifyToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
