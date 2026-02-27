package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"content-service/config"
)

// OptionalAuth optionally validates JWT token if present
func OptionalAuth(cfg *config.Config) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			tokenString := ""

			// 1) Try Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}

			// 2) Fallback to ?token= query param (for audio <audio> requests)
			if tokenString == "" {
				if rawQuery := r.URL.RawQuery; rawQuery != "" {
					values, _ := url.ParseQuery(rawQuery)
					if t := values.Get("token"); t != "" {
						tokenString = t
					}
				}
			}

			if tokenString != "" {
				claims := &UserClaims{}
				token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, errors.New("invalid signing method")
					}
					return []byte(cfg.JWTSecret), nil
				})

				if err == nil && token.Valid {
					// Check expiration
					if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
						// Token expired, continue without user context
					} else {
						ctx := context.WithValue(r.Context(), UserContextKey, claims)
						r = r.WithContext(ctx)
					}
				}
			}
			next(w, r)
		}
	}
}
