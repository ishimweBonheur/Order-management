package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

func generateToken(t *testing.T, secret string, userID, role string) string {
	t.Helper()

	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "auth-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	return tokenString
}

func TestAuthenticateValidToken(t *testing.T) {
	secret := "test-secret"
	auth := NewAuthMiddleware(secret)

	tokenString := generateToken(t, secret, "user-123", "customer")

	handler := auth.Authenticate(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				t.Error("expected user ID in context")
			}

			if userID != "user-123" {
				t.Errorf("expected user ID user-123, got %s", userID)
			}

			role, ok := UserRoleFromContext(r.Context())
			if !ok {
				t.Error("expected role in context")
			}

			if role != "customer" {
				t.Errorf("expected role customer, got %s", role)
			}

			w.WriteHeader(http.StatusOK)
		},
	))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestAuthenticateMissingHeader(t *testing.T) {
	auth := NewAuthMiddleware("test-secret")

	handler := auth.Authenticate(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	))

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthenticateInvalidHeaderFormat(t *testing.T) {
	auth := NewAuthMiddleware("test-secret")

	handler := auth.Authenticate(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "InvalidFormat")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthenticateInvalidToken(t *testing.T) {
	auth := NewAuthMiddleware("test-secret")

	handler := auth.Authenticate(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthenticateExpiredToken(t *testing.T) {
	secret := "test-secret"
	auth := NewAuthMiddleware(secret)

	claims := &Claims{
		UserID: "user-123",
		Role:   "customer",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "auth-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	handler := auth.Authenticate(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestRequireRoleAdmin(t *testing.T) {
	secret := "test-secret"
	auth := NewAuthMiddleware(secret)

	tokenString := generateToken(t, secret, "admin-123", "admin")

	protected := middlewareRequireRoleAdmin(auth, tokenString)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRequireRoleForbidden(t *testing.T) {
	secret := "test-secret"
	auth := NewAuthMiddleware(secret)

	tokenString := generateToken(t, secret, "user-123", "customer")

	protected := middlewareRequireRoleAdmin(auth, tokenString)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestRequireRoleUnauthorized(t *testing.T) {
	auth := NewAuthMiddleware("test-secret")

	protected := middlewareRequireRoleAdmin(auth, "")

	req := httptest.NewRequest(http.MethodPost, "/", nil)

	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func middlewareRequireRoleAdmin(auth *AuthMiddleware, tokenString string) http.Handler {
	return auth.Authenticate(RequireRole("admin")(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	)))
}

func TestRedisRateLimiterAllowsWithinLimit(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	rl := NewRedisRateLimiter(client, 5, time.Minute)

	handler := rl.Middleware(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected status %d, got %d", i+1, http.StatusOK, rec.Code)
		}
	}
}

func TestRedisRateLimiterBlocksOverLimit(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	rl := NewRedisRateLimiter(client, 3, time.Minute)

	handler := rl.Middleware(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected status %d, got %d", i+1, http.StatusOK, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, rec.Code)
	}
}

func TestRedisRateLimiterDifferentIPs(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	rl := NewRedisRateLimiter(client, 1, time.Minute)

	handler := rl.Middleware(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	))

	ips := []string{
		"192.168.1.1:1234",
		"192.168.1.2:1234",
		"192.168.1.3:1234",
	}

	for _, ip := range ips {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request from %s: expected status %d, got %d", ip, http.StatusOK, rec.Code)
		}
	}
}

func TestUserIDFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, "user-123")

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		t.Fatal("expected user ID in context")
	}

	if userID != "user-123" {
		t.Errorf("expected user-123, got %s", userID)
	}
}

func TestUserIDFromContextMissing(t *testing.T) {
	_, ok := UserIDFromContext(context.Background())
	if ok {
		t.Fatal("expected no user ID in context")
	}
}

func TestUserRoleFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserRoleKey, "admin")

	role, ok := UserRoleFromContext(ctx)
	if !ok {
		t.Fatal("expected role in context")
	}

	if role != "admin" {
		t.Errorf("expected admin, got %s", role)
	}
}

func TestUserRoleFromContextMissing(t *testing.T) {
	_, ok := UserRoleFromContext(context.Background())
	if ok {
		t.Fatal("expected no role in context")
	}
}
