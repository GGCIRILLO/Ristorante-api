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
	// Invalida sia la cache degli ordini normali sia quella degli ordini completi
	_, err := c.redis.Del(ctx, "ordini:all", "ordini:completi:all").Result()
	return err
}

// GetAllOrdiniCompleti recupera tutti gli ordini completi dalla cache
func (c *OrdineCache) GetAllOrdiniCompleti(ctx context.Context) ([]*models.OrdineCompleto, error) {
	key := "ordini:completi:all"
	val, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // cache miss
	} else if err != nil {
		return nil, err
	}

	var ordiniCompleti []*models.OrdineCompleto
	err = json.Unmarshal([]byte(val), &ordiniCompleti)
	return ordiniCompleti, err
}

// SetAllOrdiniCompleti salva tutti gli ordini completi nella cache
func (c *OrdineCache) SetAllOrdiniCompleti(ctx context.Context, ordiniCompleti []*models.OrdineCompleto) error {
	data, err := json.Marshal(ordiniCompleti)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, "ordini:completi:all", data, 5*time.Minute).Err()
}
