package service

import (
	"context"
	"errors"
	"testing"
	"time"

	sharedauth "github.com/ishimweBonheur/order-management/auth-service/internal/platform/auth"
	"github.com/ishimweBonheur/order-management/auth-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterLoginAndJWT(t *testing.T) {
	repo := repository.NewMemory()
	const secret = "test-secret-that-is-at-least-32-characters"
	svc := New(repo, secret, time.Hour)
	user, err := svc.Register(context.Background(), " Alice ", "ALICE@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if user.Role != "customer" || user.Email != "alice@example.com" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if user.PasswordHash == "password123" || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")) != nil {
		t.Fatal("password was not hashed correctly")
	}
	token, err := svc.Login(context.Background(), "alice@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := sharedauth.Parse(secret, token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != user.ID.String() || claims.Role != "customer" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestRegistrationAndLoginValidation(t *testing.T) {
	svc := New(repository.NewMemory(), "test-secret-that-is-at-least-32-characters", time.Hour)
	if _, err := svc.Register(context.Background(), "A", "bad-email", "short"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if _, err := svc.Login(context.Background(), "missing@example.com", "wrong"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestAssignRole(t *testing.T) {
	svc := New(repository.NewMemory(), "test-secret-that-is-at-least-32-characters", time.Hour)
	user, err := svc.Register(context.Background(), "Alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	updated, err := svc.AssignRole(context.Background(), user.ID, " ADMIN ")
	if err != nil || updated.Role != "admin" {
		t.Fatalf("unexpected role update: user=%+v err=%v", updated, err)
	}
	if _, err = svc.AssignRole(context.Background(), user.ID, "owner"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid role error, got %v", err)
	}
}
