package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"api-gateway/config"
)

type contextKey string

const UserContextKey contextKey = "user"

type UserClaims struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// enableCORS adds CORS headers to the response
func enableCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

// JWTAuth validates JWT token and extracts user claims
func JWTAuth(cfg *config.Config) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Handle OPTIONS preflight requests - allow them through without auth
			if r.Method == "OPTIONS" {
				enableCORS(w, r)
				w.WriteHeader(http.StatusOK)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				enableCORS(w, r)
				http.Error(w, "authorization header required", http.StatusUnauthorized)
				return
			}

			// Extract token from "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				enableCORS(w, r)
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Parse and validate token
			claims := &UserClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				// Validate signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("invalid signing method")
				}
				// Use JWT secret from users-service config (should be shared)
				jwtSecret := cfg.JWTSecret
				if jwtSecret == "" {
					jwtSecret = "your-secret-key-change-in-production" // Default, should match users-service
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				enableCORS(w, r)
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Add user claims to context
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next(w, r.WithContext(ctx))
		}
	}
}

// RequireAuth requires valid JWT token
func RequireAuth(cfg *config.Config) func(http.HandlerFunc) http.HandlerFunc {
	return JWTAuth(cfg)
}

// RequireRole checks if the user has the required role
func RequireRole(requiredRole string, cfg *config.Config) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		auth := JWTAuth(cfg)
		return auth(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*UserClaims)
			if !ok || claims == nil {
				enableCORS(w, r)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if claims.Role != requiredRole {
				enableCORS(w, r)
				http.Error(w, "forbidden: "+requiredRole+" access required", http.StatusForbidden)
				return
			}

			next(w, r)
		})
	}
}

// OptionalAuth optionally validates JWT token if present
func OptionalAuth(cfg *config.Config) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString := parts[1]
					claims := &UserClaims{}
					token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
						if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
							return nil, errors.New("invalid signing method")
						}
						jwtSecret := cfg.JWTSecret
						if jwtSecret == "" {
							jwtSecret = "your-secret-key-change-in-production"
						}
						return []byte(jwtSecret), nil
					})

					if err == nil && token.Valid {
						ctx := context.WithValue(r.Context(), UserContextKey, claims)
						r = r.WithContext(ctx)
					}
				}
			}
			next(w, r)
		}
	}
}
