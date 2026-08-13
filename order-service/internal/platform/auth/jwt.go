package auth

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

func authError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"code": code, "message": message}})
}

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	userRoleKey contextKey = "user_role"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func Parse(secret, raw string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid || claims.UserID == "" || claims.Role == "" {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
func Middleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Fields(r.Header.Get("Authorization"))
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				authError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required")
				return
			}
			claims, err := Parse(secret, parts[1])
			if err != nil {
				authError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Access token is invalid or expired")
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, userRoleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := Role(r.Context())
			if !ok {
				authError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required")
				return
			}
			for _, allowed := range roles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}
			authError(w, http.StatusForbidden, "FORBIDDEN", "Required role is missing")
		})
	}
}
func UserID(ctx context.Context) (string, bool) { v, ok := ctx.Value(userIDKey).(string); return v, ok }
func Role(ctx context.Context) (string, bool)   { v, ok := ctx.Value(userRoleKey).(string); return v, ok }
