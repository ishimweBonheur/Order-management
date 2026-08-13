package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/ishimweBonheur/order-management/order-service/internal/messaging"
	"github.com/ishimweBonheur/order-management/order-service/internal/model"
	"github.com/ishimweBonheur/order-management/order-service/internal/repository"
)

var ErrInvalidOrder = errors.New("order must contain valid items")

type Service struct {
	repo      repository.Repository
	publisher messaging.Publisher
}

func New(repo repository.Repository, p messaging.Publisher) *Service {
	return &Service{repo: repo, publisher: p}
}
func (s *Service) Create(ctx context.Context, userID uuid.UUID, items []model.CreateItem) (*model.Order, error) {
	if len(items) == 0 {
		return nil, ErrInvalidOrder
	}
	seen := map[uuid.UUID]bool{}
	for _, i := range items {
		if i.ProductID == uuid.Nil || i.Quantity < 1 || seen[i.ProductID] {
			return nil, ErrInvalidOrder
		}
		seen[i.ProductID] = true
	}
	o, err := s.repo.Create(ctx, userID, items)
	if err != nil {
		return nil, err
	}
	if s.publisher != nil {
		err = s.publisher.Publish(ctx, "order.created", map[string]any{"order_id": o.ID, "user_id": o.UserID, "total_amount": o.TotalAmount, "created_at": o.CreatedAt})
	}
	return o, err
}
func (s *Service) List(ctx context.Context, userID *uuid.UUID) ([]model.Order, error) {
	return s.repo.List(ctx, userID)
}
func (s *Service) Get(ctx context.Context, id, userID uuid.UUID, isAdmin bool) (*model.Order, error) {
	o, err := s.repo.ByID(ctx, id)
	if err == nil && !isAdmin && o.UserID != userID {
		return nil, repository.ErrNotFound
	}
	return o, err
}
func (s *Service) Status(ctx context.Context, id uuid.UUID, status string) (*model.Order, error) {
	valid := map[string]bool{model.StatusPending: true, model.StatusConfirmed: true, model.StatusProcessing: true, model.StatusCompleted: true, model.StatusCancelled: true}
	if !valid[status] {
		return nil, ErrInvalidOrder
	}
	return s.repo.UpdateStatus(ctx, id, status)
}
