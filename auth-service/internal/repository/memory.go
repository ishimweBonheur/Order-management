package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/ishimweBonheur/order-management/auth-service/internal/model"
	"sync"
	"time"
)

type Memory struct {
	mu    sync.RWMutex
	users map[uuid.UUID]model.User
}

func NewMemory() *Memory { return &Memory{users: map[uuid.UUID]model.User{}} }
func (r *Memory) Create(_ context.Context, u *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, x := range r.users {
		if x.Email == u.Email {
			return ErrEmailExists
		}
	}
	r.users[u.ID] = *u
	return nil
}
func (r *Memory) ByEmail(_ context.Context, email string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Email == email {
			x := u
			return &x, nil
		}
	}
	return nil, ErrNotFound
}
func (r *Memory) ByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &u, nil
}
func (r *Memory) UpdateRole(_ context.Context, id uuid.UUID, role string) (*model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	u.Role = role
	u.UpdatedAt = time.Now().UTC()
	r.users[id] = u
	return &u, nil
}
