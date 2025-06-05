package database

import (
	"context"
	"fmt"
	"log"
	"ristorante-api/config"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient è la struttura che contiene la connessione a Redis
type RedisClient struct {
	Client    *redis.Client
	HitCount  int64
	MissCount int64
}

// Chiavi cache
const (
	CacheTTL = 30 * time.Minute // TTL predefinito per la cache
)

// NewRedisClient crea un nuovo client Redis
func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
		Password: "", // Nessuna password impostata
		DB:       0,  // Usa il database predefinito
	})

	// Verifica la connessione
	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("unable to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis")

	return &RedisClient{
		Client: client,
	}, nil
}

// Close chiude la connessione a Redis
func (r *RedisClient) Close() {
	if r.Client != nil {
		if err := r.Client.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
			return
		}
		log.Println("Redis connection closed")
	}
}

// GetHitCount restituisce il numero di cache hit in modo thread-safe
func (r *RedisClient) GetHitCount() int64 {
	return atomic.LoadInt64(&r.HitCount)
}

// GetMissCount restituisce il numero di cache miss in modo thread-safe
func (r *RedisClient) GetMissCount() int64 {
	return atomic.LoadInt64(&r.MissCount)
}

// ResetCounters resetta i contatori di hit e miss
func (r *RedisClient) ResetCounters() {
	atomic.StoreInt64(&r.HitCount, 0)
	atomic.StoreInt64(&r.MissCount, 0)
}

func (r *RedisClient) GetClient() *redis.Client {
	return r.Client
}
