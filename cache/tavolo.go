package cache

import (
	"context"
	"encoding/json"
	"ristorante-api/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type TavoloCache struct {
	redis *redis.Client
}

func NewTavoloCache(client *redis.Client) *TavoloCache {
	return &TavoloCache{
		redis: client,
	}
}

const tavoliKey = "ristorante:1:tavoli:all"

// GetTavoli restituisce tutti i tavoli dalla cache
func (c *TavoloCache) GetTavoli(ctx context.Context) ([]models.Tavolo, error) {
	data, err := c.redis.Get(ctx, tavoliKey).Result()
	if err == redis.Nil {
		return nil, nil // cache miss
	}
	if err != nil {
		return nil, err
	}

	var tavoli []models.Tavolo
	if err := json.Unmarshal([]byte(data), &tavoli); err != nil {
		return nil, err
	}
	return tavoli, nil
}

// SetTavoli salva tutti i tavoli in cache
func (c *TavoloCache) SetTavoli(ctx context.Context, tavoli []models.Tavolo) error {
	data, err := json.Marshal(tavoli)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, tavoliKey, data, 5*time.Minute).Err()
}

// InvalidateTavoli cancella la cache dei tavoli
func (c *TavoloCache) InvalidateTavoli(ctx context.Context) error {
	return c.redis.Del(ctx, tavoliKey).Err()
}
