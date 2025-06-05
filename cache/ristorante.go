package cache

import (
	"context"
	"encoding/json"
	"ristorante-api/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type RistoranteCache struct {
	Client *redis.Client
}

func NewRistoranteCache(rdb *redis.Client) *RistoranteCache {
	return &RistoranteCache{Client: rdb}
}

func (c *RistoranteCache) GetAll(ctx context.Context) ([]models.Ristorante, error) {
	key := "ristoranti:all"
	val, err := c.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	} else if err != nil {
		return nil, err
	}

	var result []models.Ristorante
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *RistoranteCache) SetAll(ctx context.Context, ristoranti []models.Ristorante) error {
	key := "ristoranti:all"
	data, err := json.Marshal(ristoranti)
	if err != nil {
		return err
	}
	return c.Client.Set(ctx, key, data, 5*time.Minute).Err()
}

func (c *RistoranteCache) DeleteAll(ctx context.Context) error {
	return c.Client.Del(ctx, "ristoranti:all").Err()
}
