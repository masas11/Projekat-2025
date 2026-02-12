package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"api-gateway/config"
	"api-gateway/internal/logger"
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

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// JWTAuth validates JWT token and extracts user claims
func JWTAuth(cfg *config.Config, log *logger.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Handle OPTIONS preflight requests - allow them through without auth
			if r.Method == "OPTIONS" {
				enableCORS(w, r)
				w.WriteHeader(http.StatusOK)
				return
			}

			ipAddress := getClientIP(r)
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				if log != nil {
					log.LogAccessControlFailure("", r.URL.Path, r.Method, "missing authorization header")
				}
				enableCORS(w, r)
				http.Error(w, "authorization header required", http.StatusUnauthorized)
				return
			}

			// Extract token from "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				if log != nil {
					log.LogAccessControlFailure("", r.URL.Path, r.Method, "invalid authorization header format")
				}
				enableCORS(w, r)
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			tokenPrefix := ""
			if len(tokenString) > 10 {
				tokenPrefix = tokenString[:10] + "..."
			} else {
				tokenPrefix = "***"
			}

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
				reason := "invalid token"
				if err != nil {
					if strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "exp") {
						reason = "expired token"
						if log != nil && claims.UserID != "" {
							log.LogExpiredToken(claims.UserID, ipAddress)
						} else {
							log.LogInvalidToken(tokenPrefix, reason, ipAddress)
						}
					} else {
						reason = err.Error()
						if log != nil {
							log.LogInvalidToken(tokenPrefix, reason, ipAddress)
						}
					}
				} else {
					if log != nil {
						log.LogInvalidToken(tokenPrefix, reason, ipAddress)
					}
				}
				enableCORS(w, r)
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Check if token is expired
			if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
				if log != nil {
					log.LogExpiredToken(claims.UserID, ipAddress)
				}
				enableCORS(w, r)
				http.Error(w, "expired token", http.StatusUnauthorized)
				return
			}

			// Add user claims to context
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next(w, r.WithContext(ctx))
		}
	}
}

// RequireAuth requires valid JWT token
func RequireAuth(cfg *config.Config, log *logger.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return JWTAuth(cfg, log)
}

// RequireRole checks if the user has the required role
func RequireRole(requiredRole string, cfg *config.Config, log *logger.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		auth := JWTAuth(cfg, log)
		return auth(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*UserClaims)
			if !ok || claims == nil {
				if log != nil {
					log.LogAccessControlFailure("", r.URL.Path, r.Method, "missing user claims in context")
				}
				enableCORS(w, r)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if claims.Role != requiredRole {
				if log != nil {
					log.LogAccessControlFailure(claims.UserID, r.URL.Path, r.Method, 
						"insufficient permissions: required role "+requiredRole+", user role "+claims.Role)
				}
				enableCORS(w, r)
				http.Error(w, "forbidden: "+requiredRole+" access required", http.StatusForbidden)
				return
			}

			next(w, r)
		})
	}
}

// OptionalAuth optionally validates JWT token if present
func OptionalAuth(cfg *config.Config, log *logger.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString := parts[1]
					tokenPrefix := ""
					if len(tokenString) > 10 {
						tokenPrefix = tokenString[:10] + "..."
					} else {
						tokenPrefix = "***"
					}
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
						// Check expiration
						if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
							if log != nil {
								log.LogExpiredToken(claims.UserID, getClientIP(r))
							}
						} else {
							ctx := context.WithValue(r.Context(), UserContextKey, claims)
							r = r.WithContext(ctx)
						}
					} else if err != nil && log != nil {
						reason := err.Error()
						if strings.Contains(reason, "expired") {
							log.LogExpiredToken(claims.UserID, getClientIP(r))
						} else {
							log.LogInvalidToken(tokenPrefix, reason, getClientIP(r))
						}
					}
				}
			}
			next(w, r)
		}
	}
}
