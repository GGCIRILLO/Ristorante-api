package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"ristorante-api/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type RicettaCache struct {
	redis *redis.Client
}

func NewRicettaCache(rdb *redis.Client) *RicettaCache {
	return &RicettaCache{redis: rdb}
}

// GetByPietanzaID recupera una ricetta dalla cache in base all'ID della pietanza
func (c *RicettaCache) GetByPietanzaID(ctx context.Context, idPietanza int) (*models.Ricetta, bool, error) {
	key := fmt.Sprintf("ricetta:pietanza:%d", idPietanza)
	val, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil // cache miss
	} else if err != nil {
		return nil, false, err
	}

	var ricetta models.Ricetta
	err = json.Unmarshal([]byte(val), &ricetta)
	return &ricetta, true, err
}

// SetByPietanzaID salva una ricetta nella cache associandola all'ID della pietanza
func (c *RicettaCache) SetByPietanzaID(ctx context.Context, idPietanza int, ricetta *models.Ricetta) error {
	key := fmt.Sprintf("ricetta:pietanza:%d", idPietanza)
	data, err := json.Marshal(ricetta)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, key, data, 30*time.Minute).Err()
}

// InvalidateByPietanzaID rimuove una ricetta dalla cache
func (c *RicettaCache) InvalidateByPietanzaID(ctx context.Context, idPietanza int) error {
	key := fmt.Sprintf("ricetta:pietanza:%d", idPietanza)
	return c.redis.Del(ctx, key).Err()
}

// GetIngredientiByRicettaID recupera gli ingredienti di una ricetta dalla cache
func (c *RicettaCache) GetIngredientiByRicettaID(ctx context.Context, idRicetta int) ([]models.RicettaIngrediente, bool, error) {
	key := fmt.Sprintf("ricetta:%d:ingredienti", idRicetta)
	val, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil // cache miss
	} else if err != nil {
		return nil, false, err
	}

	var ingredienti []models.RicettaIngrediente
	err = json.Unmarshal([]byte(val), &ingredienti)
	return ingredienti, true, err
}

// SetIngredientiByRicettaID salva gli ingredienti di una ricetta nella cache
func (c *RicettaCache) SetIngredientiByRicettaID(ctx context.Context, idRicetta int, ingredienti []models.RicettaIngrediente) error {
	key := fmt.Sprintf("ricetta:%d:ingredienti", idRicetta)
	data, err := json.Marshal(ingredienti)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, key, data, 30*time.Minute).Err()
}

// InvalidateIngredientiByRicettaID rimuove gli ingredienti di una ricetta dalla cache
func (c *RicettaCache) InvalidateIngredientiByRicettaID(ctx context.Context, idRicetta int) error {
	key := fmt.Sprintf("ricetta:%d:ingredienti", idRicetta)
	return c.redis.Del(ctx, key).Err()
}

// GetRicettaCompletaByPietanzaID recupera una ricetta completa dalla cache in base all'ID della pietanza
func (c *RicettaCache) GetRicettaCompletaByPietanzaID(ctx context.Context, idPietanza int) (*models.RicettaCompleta, bool, error) {
	key := fmt.Sprintf("ricetta:completa:pietanza:%d", idPietanza)
	val, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil // cache miss
	} else if err != nil {
		return nil, false, err
	}

	var ricettaCompleta models.RicettaCompleta
	if err := json.Unmarshal([]byte(val), &ricettaCompleta); err != nil {
		return nil, false, err
	}

	return &ricettaCompleta, true, nil
}

// SetRicettaCompletaByPietanzaID salva una ricetta completa nella cache in base all'ID della pietanza
func (c *RicettaCache) SetRicettaCompletaByPietanzaID(ctx context.Context, idPietanza int, ricetta *models.RicettaCompleta) error {
	key := fmt.Sprintf("ricetta:completa:pietanza:%d", idPietanza)
	val, err := json.Marshal(ricetta)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, val, 30*time.Minute).Err()
}

// InvalidateRicettaCompletaByPietanzaID invalida la cache della ricetta completa per una pietanza specifica
func (c *RicettaCache) InvalidateRicettaCompletaByPietanzaID(ctx context.Context, idPietanza int) error {
	key := fmt.Sprintf("ricetta:completa:pietanza:%d", idPietanza)
	return c.redis.Del(ctx, key).Err()
}
