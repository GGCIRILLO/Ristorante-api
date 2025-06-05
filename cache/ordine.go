package cache

import (
	"context"
	"encoding/json"
	"ristorante-api/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type OrdineCache struct {
	redis *redis.Client
}

func NewOrdineCache(rdb *redis.Client) *OrdineCache {
	return &OrdineCache{redis: rdb}
}

func (c *OrdineCache) GetAll(ctx context.Context) ([]models.Ordine, error) {
	key := "ordini:all"
	val, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // cache miss
	} else if err != nil {
		return nil, err
	}
	var ordini []models.Ordine
	err = json.Unmarshal([]byte(val), &ordini)
	return ordini, err
}

func (c *OrdineCache) SetAll(ctx context.Context, ordini []models.Ordine) error {
	data, err := json.Marshal(ordini)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, "ordini:all", data, 5*time.Minute).Err()
}

func (c *OrdineCache) Invalidate(ctx context.Context) error {
	return c.redis.Del(ctx, "ordini:all").Err()
}
