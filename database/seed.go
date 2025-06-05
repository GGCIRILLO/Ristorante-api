package database

import (
	"context"
	"log"
	"os"
	"path/filepath"
)


// SeedFromFile esegue uno script SQL per popolare il database con dati di esempio
func (db *DB) SeedFromFile(fileName string) error {
	log.Printf("Seeding database from file: %s", fileName)

	// Verifica se eseguire il seeding in base a conteggi
	// Prima controlla se ci sono già ristoranti nel database
	var count int
	err := db.Pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM ristorante").Scan(&count)
	if err != nil {
		log.Printf("Errore nel controllo dei ristoranti: %v", err)
	}

	// Se ci sono già dati, salta il seeding
	if count > 0 {
		log.Println("Il database contiene già dati. Seeding skippato.")
		return nil
	}

	// Determina il percorso del file
	scriptPath := filepath.Join("database", "migrations", fileName)
	
	// Controlla se il file esiste
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		log.Printf("File di seed non trovato: %s", scriptPath)
		
		// Prova un percorso alternativo
		scriptPath = filepath.Join("migrations", fileName)
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			return err
		}
	}
	
	// Leggi il contenuto del file SQL
	sqlBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}
	
	// Esegui lo script SQL
	_, err = db.Pool.Exec(context.Background(), string(sqlBytes))
	if err != nil {
		return err
	}
	
	log.Printf("Seeding completato con successo da file: %s", fileName)
	return nil
}
