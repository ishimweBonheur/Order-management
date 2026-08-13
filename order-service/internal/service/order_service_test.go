package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/ishimweBonheur/order-management/order-service/internal/model"
	"testing"
	"time"
)

type fakeRepo struct{ order *model.Order }

func (f *fakeRepo) Create(_ context.Context, u uuid.UUID, i []model.CreateItem) (*model.Order, error) {
	f.order = &model.Order{ID: uuid.New(), UserID: u, Status: model.StatusPending, Items: []model.Item{{ProductID: i[0].ProductID, Quantity: i[0].Quantity}}, CreatedAt: time.Now()}
	return f.order, nil
}
func (f *fakeRepo) List(context.Context, *uuid.UUID) ([]model.Order, error) {
	return []model.Order{*f.order}, nil
}
func (f *fakeRepo) ByID(context.Context, uuid.UUID) (*model.Order, error) { return f.order, nil }
func (f *fakeRepo) UpdateStatus(_ context.Context, _ uuid.UUID, s string) (*model.Order, error) {
	f.order.Status = s
	return f.order, nil
}
func (f *fakeRepo) AdminEmail(context.Context) (string, error) {
	return "admin@example.com", nil
}

type fakePublisher struct{ called bool }

func (p *fakePublisher) Publish(context.Context, string, any) error { p.called = true; return nil }
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
