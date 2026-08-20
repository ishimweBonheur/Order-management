package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ishimweBonheur/order-management/auth-service/internal/model"
	sharedauth "github.com/ishimweBonheur/order-management/auth-service/internal/platform/auth"
	"github.com/ishimweBonheur/order-management/auth-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidInput       = errors.New("invalid input")
)

type AuthService struct {
	repo   repository.UserRepository
	secret string
	ttl    time.Duration
}

func New(repo repository.UserRepository, secret string, ttl time.Duration) *AuthService {
	return &AuthService{repo: repo, secret: secret, ttl: ttl}
}
func (s *AuthService) Register(ctx context.Context, name, email, password string) (*model.User, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	if name == "" || len(password) < 8 {
		return nil, ErrInvalidInput
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, ErrInvalidInput
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	u := &model.User{ID: uuid.New(), Name: name, Email: email, PasswordHash: string(hash), Role: model.RoleCustomer, CreatedAt: now, UpdatedAt: now}
	if err = s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	u, err := s.repo.ByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}
	return sharedauth.Generate(s.secret, "auth-service", u.ID.String(), u.Role, s.ttl)
}
func (s *AuthService) Me(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.repo.ByID(ctx, id)
}
func (s *AuthService) AssignRole(ctx context.Context, id uuid.UUID, role string) (*model.User, error) {
	role = strings.ToLower(strings.TrimSpace(role))
	if role != model.RoleAdmin && role != model.RoleCustomer {
		return nil, ErrInvalidInput
	}
	return s.repo.UpdateRole(ctx, id, role)
}
