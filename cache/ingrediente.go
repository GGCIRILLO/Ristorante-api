package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"ristorante-api/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type IngredienteCache struct {
	redis *redis.Client
}

func NewIngredienteCache(rdb *redis.Client) *IngredienteCache {
	return &IngredienteCache{redis: rdb}
}

// GetAll recupera tutti gli ingredienti dalla cache
func (c *IngredienteCache) GetAll(ctx context.Context) ([]models.Ingrediente, bool, error) {
	key := "ingredienti:all"
	val, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil // cache miss
	}
	if err != nil {
		return nil, false, err
	}
	var ingredienti []models.Ingrediente
	err = json.Unmarshal([]byte(val), &ingredienti)
	return ingredienti, true, err
}

// SetAll salva tutti gli ingredienti nella cache
func (c *IngredienteCache) SetAll(ctx context.Context, ingredienti []models.Ingrediente) error {
	data, err := json.Marshal(ingredienti)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, "ingredienti:all", data, 10*time.Minute).Err()
}

// InvalidateAll rimuove tutti gli ingredienti dalla cache
func (c *IngredienteCache) InvalidateAll(ctx context.Context) error {
	return c.redis.Del(ctx, "ingredienti:all").Err()
}

// GetByID recupera un ingrediente specifico dalla cache in base all'ID
func (c *IngredienteCache) GetByID(ctx context.Context, id int) (*models.Ingrediente, bool, error) {
	key := fmt.Sprintf("ingrediente:%d", id)
	val, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil // cache miss
	} else if err != nil {
		return nil, false, err
	}

	var ingrediente models.Ingrediente
	err = json.Unmarshal([]byte(val), &ingrediente)
	return &ingrediente, true, err
}

// SetByID salva un ingrediente specifico nella cache in base all'ID
func (c *IngredienteCache) SetByID(ctx context.Context, ingrediente *models.Ingrediente) error {
	key := fmt.Sprintf("ingrediente:%d", ingrediente.ID)
	data, err := json.Marshal(ingrediente)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, key, data, 10*time.Minute).Err()
}

// InvalidateByID rimuove un ingrediente specifico dalla cache in base all'ID
func (c *IngredienteCache) InvalidateByID(ctx context.Context, id int) error {
	key := fmt.Sprintf("ingrediente:%d", id)
	return c.redis.Del(ctx, key).Err()
}

// Get ingredienti da riordinare
func (c *IngredienteCache) GetDaRiordinare(ctx context.Context) ([]models.Ingrediente, bool, error) {
	key := "ingredienti:da_riordinare"
	val, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil // cache miss
	}
	if err != nil {
		return nil, false, err
	}
	var ingredienti []models.Ingrediente
	err = json.Unmarshal([]byte(val), &ingredienti)
	return ingredienti, true, err
}

// SetDaRiordinare salva gli ingredienti da riordinare nella cache
func (c *IngredienteCache) SetDaRiordinare(ctx context.Context, ingredienti []models.Ingrediente) error {
	data, err := json.Marshal(ingredienti)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, "ingredienti:da_riordinare", data, 10*time.Minute).Err()
}	
// InvalidateDaRiordinare rimuove gli ingredienti da riordinare dalla cache
func (c *IngredienteCache) InvalidateDaRiordinare(ctx context.Context) error {
	return c.redis.Del(ctx, "ingredienti:da_riordinare").Err()
}
