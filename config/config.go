package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config struttura per la configurazione dell'applicazione
type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	RedisHost  string
	RedisPort  int
}

// LoadConfig carica la configurazione da variabili d'ambiente o file .env
func LoadConfig() (*Config, error) {
	// Tenta di caricare le variabili d'ambiente da .env file se presente
	_ = godotenv.Load()

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %v", err)
	}

	redisPort, err := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_PORT: %v", err)
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     dbPort,
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "ristorante"),
		RedisHost:  getEnv("REDIS_HOST", "localhost"),
		RedisPort:  redisPort,
	}, nil
}

// getEnv ottiene una variabile d'ambiente o restituisce un valore default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
