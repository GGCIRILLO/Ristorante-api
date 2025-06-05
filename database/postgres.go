package database

import (
	"context"
	"fmt"
	"log"
	"ristorante-api/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB è la struttura che contiene le connessioni al database e cache
type DB struct {
	Pool  *pgxpool.Pool
	Redis *RedisClient
}

// New crea una nuova connessione al database PostgreSQL e Redis
func New(cfg *config.Config) (*DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	// Crea un pool di connessioni
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %v", err)
	}

	// Connessione al database
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	// Verifica connessione
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	log.Println("Connected to PostgreSQL database")

	// Crea client Redis
	redis, err := NewRedisClient(cfg)
	if err != nil {
		// Chiudi il pool PostgreSQL prima di uscire
		pool.Close()
		return nil, err
	}

	return &DB{
		Pool:  pool,
		Redis: redis,
	}, nil
}

// Close chiude le connessioni al database e cache
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		log.Println("Database connection closed")
	}

	if db.Redis != nil {
		db.Redis.Close()
	}
}

// InitSchema inizializza lo schema del database
func (db *DB) InitSchema() error {
	var err error

	// Inizializza lo schema
	// Tabella Ristorante
	_, err = db.Pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ristorante (
		  id_ristorante SERIAL PRIMARY KEY,
		  nome VARCHAR(100) NOT NULL,
		  numero_tavoli INTEGER NOT NULL,
		  costo_coperto DECIMAL(10,2) NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create ristorante table: %v", err)
	}

	// Tabella Tavolo
	_, err = db.Pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS tavolo (
		  id_tavolo SERIAL PRIMARY KEY,
		  max_posti INTEGER NOT NULL,
		  stato VARCHAR(10) NOT NULL DEFAULT 'libero' CHECK (stato IN ('libero', 'occupato')),
		  id_ristorante INTEGER NOT NULL,
		  FOREIGN KEY (id_ristorante) REFERENCES ristorante (id_ristorante) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tavolo table: %v", err)
	}

	// Tabella Ingrediente
	_, err = db.Pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ingrediente (
			id_ingrediente SERIAL PRIMARY KEY,
			nome VARCHAR(100) NOT NULL,
			quantita_disponibile FLOAT NOT NULL,
			unita_misura VARCHAR(20) NOT NULL,
			soglia_riordino FLOAT NOT NULL
		)
  `)
	if err != nil {
		return fmt.Errorf("failed to create ingrediente table: %v", err)
	}

	// Tabella Categoria Pietanza
	_, err = db.Pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS categoria_pietanza (
		  id_categoria SERIAL PRIMARY KEY,
		  nome VARCHAR(50) NOT NULL UNIQUE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create categoria_pietanza table: %v", err)
	}

	// Tabella Pietanza
	_, err = db.Pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS pietanza (
		  id_pietanza SERIAL PRIMARY KEY,
		  nome VARCHAR(100) NOT NULL,
		  prezzo DECIMAL(10,2) NOT NULL,
		  id_categoria INTEGER,
		  disponibile BOOLEAN NOT NULL DEFAULT TRUE,
		  FOREIGN KEY (id_categoria) REFERENCES categoria_pietanza (id_categoria) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create pietanza table: %v", err)
	}

	// Tabella Menu
	_, err = db.Pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS menu (
		  id_ristorante INTEGER NOT NULL,
		  id_pietanza INTEGER NOT NULL,
		  PRIMARY KEY (id_ristorante, id_pietanza),
		  FOREIGN KEY (id_ristorante) REFERENCES ristorante (id_ristorante) ON DELETE CASCADE,
		  FOREIGN KEY (id_pietanza) REFERENCES pietanza (id_pietanza) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create menu table: %v", err)
	}

	// Tabella Ricetta
	_, err = db.Pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ricetta (
		  id_ricetta SERIAL PRIMARY KEY,
		  nome VARCHAR(100) NOT NULL,
		  descrizione TEXT NOT NULL,
		  id_pietanza INTEGER NOT NULL,			
		  tempo_preparazione INTEGER,
		  istruzioni TEXT,
		  FOREIGN KEY (id_pietanza) REFERENCES pietanza (id_pietanza) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create ricetta table: %v", err)
	}

	// Tabella Ricetta Ingrediente
	_, err = db.Pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ricetta_ingrediente (
		  id_ricetta INTEGER NOT NULL,
		  id_ingrediente INTEGER NOT NULL,
		  quantita FLOAT NOT NULL,
		  PRIMARY KEY (id_ricetta, id_ingrediente),
		  FOREIGN KEY (id_ricetta) REFERENCES ricetta (id_ricetta) ON DELETE CASCADE,
		  FOREIGN KEY (id_ingrediente) REFERENCES ingrediente (id_ingrediente) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create ricetta_ingrediente table: %v", err)
	}

	// Tabella Menu Fisso
	_, err = db.Pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS menu_fisso (
		  id_menu SERIAL PRIMARY KEY,
		  nome VARCHAR(100) NOT NULL,
		  prezzo DECIMAL(10,2) NOT NULL,
		  descrizione TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create menu_fisso table: %v", err)
	}

	// Tabella Composizione Menu Fisso
	_, err = db.Pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS composizione_menu_fisso (
		  id_menu INTEGER NOT NULL,
		  id_pietanza INTEGER NOT NULL,
		  PRIMARY KEY (id_menu, id_pietanza),
		  FOREIGN KEY (id_menu) REFERENCES menu_fisso (id_menu) ON DELETE CASCADE,
		  FOREIGN KEY (id_pietanza) REFERENCES pietanza (id_pietanza) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create composizione_menu_fisso table: %v", err)
	}

	// Tabella Ordine
	_, err = db.Pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ordine (
		  id_ordine SERIAL PRIMARY KEY,
		  id_tavolo INTEGER NOT NULL,
		  num_persone INTEGER NOT NULL,
		  data_ordine TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		  stato VARCHAR(20) NOT NULL DEFAULT 'in_attesa' CHECK (stato IN ('in_attesa', 'in_preparazione', 'pronto', 'consegnato', 'pagato')),
		  id_ristorante INTEGER NOT NULL,
		  costo_totale DECIMAL(10,2) NOT NULL DEFAULT 0.00,
		  FOREIGN KEY (id_tavolo) REFERENCES tavolo (id_tavolo) ON DELETE CASCADE,
		  FOREIGN KEY (id_ristorante) REFERENCES ristorante (id_ristorante) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create ordine table: %v", err)
	}

	// Tabella Dettaglio Ordine
	_, err = db.Pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS dettaglio_ordine_pietanza (
		  id_dettaglio SERIAL PRIMARY KEY,
		  id_ordine INTEGER NOT NULL,
		  id_pietanza INTEGER NOT NULL,
		  quantita INTEGER NOT NULL DEFAULT 1,
		  parte_di_menu BOOLEAN NOT NULL DEFAULT FALSE,
		  id_menu INTEGER DEFAULT NULL,
		  FOREIGN KEY (id_ordine) REFERENCES ordine (id_ordine) ON DELETE CASCADE,
		  FOREIGN KEY (id_pietanza) REFERENCES pietanza (id_pietanza) ON DELETE CASCADE,
		  UNIQUE (id_ordine, id_pietanza)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create dettaglio_ordine_pietanza table: %v", err)
	}

	// Indici per migliorare le performance
	_, err = db.Pool.Exec(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_ordine_tavolo ON ordine (id_tavolo);
		CREATE INDEX IF NOT EXISTS idx_ordine_ristorante ON ordine (id_ristorante);
		CREATE INDEX IF NOT EXISTS idx_ordine_stato ON ordine (stato);
		CREATE INDEX IF NOT EXISTS idx_dettaglio_ordine ON dettaglio_ordine_pietanza (id_ordine);
	`)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %v", err)
	}

	log.Println("Database schema initialized")
	return nil
}
