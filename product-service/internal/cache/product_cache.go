package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/ishimweBonheur/order-management/product-service/internal/model"
	"github.com/redis/go-redis/v9"
)

type ProductCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewProductCache(
	client *redis.Client,
	ttl time.Duration,
) *ProductCache {
	return &ProductCache{
		client: client,
		ttl:    ttl,
	}
}
func productKey(id uuid.UUID) string {
	return "product:" + id.String()
}

func (c *ProductCache) Get(
	ctx context.Context,
	id uuid.UUID,
) (*model.Product, error) {

	key := productKey(id)

	value, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var product model.Product

	if err := json.Unmarshal(
		[]byte(value),
		&product,
	); err != nil {
		return nil, err
	}

	return &product, nil
}

func (c *ProductCache) Set(
	ctx context.Context,
	product *model.Product,
) error {

	value, err := json.Marshal(product)
	if err != nil {
		return err
	}

	key := "product:" + product.ID.String()

	return c.client.Set(
		ctx,
		key,
		value,
		c.ttl,
	).Err()
}

func (c *ProductCache) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {

	key := "product:" + id.String()

	return c.client.Del(ctx, key).Err()
}
