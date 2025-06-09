package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"ristorante-api/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type MenuFissoCache struct {
	redis *redis.Client
}

func NewMenuFissoCache(rdb *redis.Client) *MenuFissoCache {
	return &MenuFissoCache{redis: rdb}
}

// GetAll recupera tutti i menu fissi dalla cache
func (c *MenuFissoCache) GetAll(ctx context.Context) ([]models.MenuFisso, bool, error) {
	key := "menu_fissi:all"
	val, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil // cache miss
	} else if err != nil {
		return nil, false, err
	}

	var menuFissi []models.MenuFisso
	err = json.Unmarshal([]byte(val), &menuFissi)
	return menuFissi, true, err
}

// SetAll salva tutti i menu fissi nella cache
func (c *MenuFissoCache) SetAll(ctx context.Context, menuFissi []models.MenuFisso) error {
	data, err := json.Marshal(menuFissi)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, "menu_fissi:all", data, 10*time.Minute).Err()
}

// InvalidateAll rimuove tutti i menu fissi dalla cache
func (c *MenuFissoCache) InvalidateAll(ctx context.Context) error {
	return c.redis.Del(ctx, "menu_fissi:all").Err()
}

// GetByID recupera un menu fisso specifico dalla cache in base all'ID
func (c *MenuFissoCache) GetByID(ctx context.Context, id int) (*models.MenuFisso, bool, error) {
	key := fmt.Sprintf("menu_fisso:%d", id)
	val, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil // cache miss
	} else if err != nil {
		return nil, false, err
	}

	var menuFisso models.MenuFisso
	err = json.Unmarshal([]byte(val), &menuFisso)
	return &menuFisso, true, err
}

// SetByID salva un menu fisso specifico nella cache
func (c *MenuFissoCache) SetByID(ctx context.Context, id int, menuFisso *models.MenuFisso) error {
	data, err := json.Marshal(menuFisso)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, fmt.Sprintf("menu_fisso:%d", id), data, 10*time.Minute).Err()
}

// InvalidateByID rimuove un menu fisso specifico dalla cache
func (c *MenuFissoCache) InvalidateByID(ctx context.Context, id int) error {
	return c.redis.Del(ctx, fmt.Sprintf("menu_fisso:%d", id)).Err()
}

// GetComposizione recupera la composizione di un menu fisso dalla cache
func (c *MenuFissoCache) GetComposizione(ctx context.Context, idMenu int) ([]models.Pietanza, bool, error) {
	key := fmt.Sprintf("menu_fisso:%d:composizione", idMenu)
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

// SetComposizione salva la composizione di un menu fisso nella cache
func (c *MenuFissoCache) SetComposizione(ctx context.Context, idMenu int, pietanze []models.Pietanza) error {
	data, err := json.Marshal(pietanze)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, fmt.Sprintf("menu_fisso:%d:composizione", idMenu), data, 10*time.Minute).Err()
}

// InvalidateComposizione rimuove la composizione di un menu fisso dalla cache
func (c *MenuFissoCache) InvalidateComposizione(ctx context.Context, idMenu int) error {
	return c.redis.Del(ctx, fmt.Sprintf("menu_fisso:%d:composizione", idMenu)).Err()
}
