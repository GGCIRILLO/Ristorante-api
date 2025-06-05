package main

import (
	"log"
	"net/http"
	"ristorante-api/api"
	"ristorante-api/config"
	"ristorante-api/database"
)

func main() {
	// Caricamento configurazione
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connessione al database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Inizializzazione schema database
	if err := db.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Esegui il seeding
	if err := db.SeedFromFile("complete_seed.sql"); err != nil {
		log.Printf("Warning: error executing complete seed file: %v", err)
	}

	// Configurazione router
	router := api.SetupRoutes(db)

	// Avvio server
	addr := ":8080"
	log.Printf("Server starting on %s", addr)
	err = http.ListenAndServe(addr, router)
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
