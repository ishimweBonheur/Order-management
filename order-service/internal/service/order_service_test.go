package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/ishimweBonheur/order-management/order-service/internal/model"
	"github.com/ishimweBonheur/order-management/order-service/internal/repository"
	"testing"
	"time"
)

type fakeRepo struct{ order *model.Order }

func (f *fakeRepo) Create(_ context.Context, u uuid.UUID, i []model.CreateItem) (*model.Order, error) {
	items := make([]model.Item, 0, len(i))
	for _, requested := range i {
		items = append(items, model.Item{ProductID: requested.ProductID, Quantity: requested.Quantity})
	}
	f.order = &model.Order{ID: uuid.New(), UserID: u, Status: model.StatusPending, Items: items, CreatedAt: time.Now()}
	return f.order, nil
}
func (f *fakeRepo) List(context.Context, *uuid.UUID, int, int) ([]model.Order, int, error) {
	return []model.Order{*f.order}, 1, nil
}
func (f *fakeRepo) ByID(context.Context, uuid.UUID) (*model.Order, error) { return f.order, nil }
func (f *fakeRepo) UpdateStatus(_ context.Context, _ uuid.UUID, s string) (*model.Order, error) {
	f.order.Status = s
	return f.order, nil
}
func (f *fakeRepo) AdminEmail(context.Context) (string, error) {
	return "admin@example.com", nil
}

type fakePublisher struct {
	called  bool
	payload any
}

func (p *fakePublisher) Publish(_ context.Context, _ string, payload any) error {
	p.called = true
	p.payload = payload
	return nil
}
func TestCreateOrder(t *testing.T) {
	repo := &fakeRepo{}
	pub := &fakePublisher{}
	svc := New(repo, pub)
	userID, productID := uuid.New(), uuid.New()
	o, err := svc.Create(context.Background(), userID, []model.CreateItem{{ProductID: productID, Quantity: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if o.UserID != userID || !pub.called {
		t.Fatal("order was not created and published")
	}
}
func TestCreateOrderWithMultipleProducts(t *testing.T) {
	repo, pub := &fakeRepo{}, &fakePublisher{}
	first, second := uuid.New(), uuid.New()
	order, err := New(repo, pub).Create(context.Background(), uuid.New(), []model.CreateItem{{ProductID: first, Quantity: 2}, {ProductID: second, Quantity: 3}})
	if err != nil {
		t.Fatal(err)
	}
	if len(order.Items) != 2 || !pub.called {
		t.Fatalf("expected two products and a published event: %+v", order)
	}
}
func TestCustomerCannotReadAnotherCustomersOrder(t *testing.T) {
	repo := &fakeRepo{order: &model.Order{ID: uuid.New(), UserID: uuid.New()}}
	if _, err := New(repo, nil).Get(context.Background(), repo.order.ID, uuid.New(), false); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected hidden order, got %v", err)
	}
}
func TestAdminCanReadAnyOrder(t *testing.T) {
	repo := &fakeRepo{order: &model.Order{ID: uuid.New(), UserID: uuid.New()}}
	if _, err := New(repo, nil).Get(context.Background(), repo.order.ID, uuid.New(), true); err != nil {
		t.Fatal(err)
	}
}
func TestRejectsInvalidStatus(t *testing.T) {
	if _, err := New(&fakeRepo{order: &model.Order{}}, nil).Status(context.Background(), uuid.New(), "unknown"); !errors.Is(err, ErrInvalidOrder) {
		t.Fatalf("expected invalid status, got %v", err)
	}
}
func TestCreateOrderRejectsInvalidItems(t *testing.T) {
	svc := New(&fakeRepo{}, nil)
	if _, err := svc.Create(context.Background(), uuid.New(), nil); err == nil {
		t.Fatal("expected validation error")
	}
	id := uuid.New()
	if _, err := svc.Create(context.Background(), uuid.New(), []model.CreateItem{{ProductID: id, Quantity: 1}, {ProductID: id, Quantity: 1}}); err == nil {
		t.Fatal("expected duplicate product validation error")
	}
}
