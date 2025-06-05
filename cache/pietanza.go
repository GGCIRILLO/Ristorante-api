package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"ristorante-api/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type PietanzaCache struct {
	redis *redis.Client
}

func NewPietanzaCache(rdb *redis.Client) *PietanzaCache {
	return &PietanzaCache{redis: rdb}
}

// GetAll recupera tutte le pietanze dalla cache
func (c *PietanzaCache) GetAll(ctx context.Context) ([]models.Pietanza, bool, error) {
	key := "pietanze:all"
	val, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil // cache miss
	} else if err != nil {
		return nil, false, err
	}

	var pietanze []models.Pietanza
	err = json.Unmarshal([]byte(val), &pietanze)
	return pietanze, true, err
}

// SetAll salva tutte le pietanze nella cache
func (c *PietanzaCache) SetAll(ctx context.Context, pietanze []models.Pietanza) error {
	data, err := json.Marshal(pietanze)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, "pietanze:all", data, 10*time.Minute).Err()
}

// InvalidateAll rimuove tutte le pietanze dalla cache
func (c *PietanzaCache) InvalidateAll(ctx context.Context) error {
	return c.redis.Del(ctx, "pietanze:all").Err()
}

// GetByID recupera una pietanza specifica dalla cache in base all'ID
func (c *PietanzaCache) GetByID(ctx context.Context, id int) (*models.Pietanza, bool, error) {
	key := fmt.Sprintf("pietanza:%d", id)
	val, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil // cache miss
	} else if err != nil {
		return nil, false, err
	}

	var pietanza models.Pietanza
	err = json.Unmarshal([]byte(val), &pietanza)
	return &pietanza, true, err
}

// SetByID salva una pietanza specifica nella cache
func (c *PietanzaCache) SetByID(ctx context.Context, id int, pietanza *models.Pietanza) error {
	data, err := json.Marshal(pietanza)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, fmt.Sprintf("pietanza:%d", id), data, 10*time.Minute).Err()
}

// InvalidateByID rimuove una pietanza specifica dalla cache
func (c *PietanzaCache) InvalidateByID(ctx context.Context, id int) error {
	return c.redis.Del(ctx, fmt.Sprintf("pietanza:%d", id)).Err()
}
