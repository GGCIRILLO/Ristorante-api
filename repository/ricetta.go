package repository

import (
	"context"
	"ristorante-api/cache"
	"ristorante-api/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RicettaRepository struct {
	DB    *pgxpool.Pool
	Cache *cache.RicettaCache
}

func NewRicettaRepository(db *pgxpool.Pool, cache *cache.RicettaCache) *RicettaRepository {
	return &RicettaRepository{
		DB:    db,
		Cache: cache,
	}
}

// GetByPietanzaID restituisce la ricetta associata a una pietanza
func (r *RicettaRepository) GetByPietanzaID(ctx context.Context, idPietanza int) (*models.Ricetta, error) {
	// Controlla prima nella cache
	if r.Cache != nil {
		cached, found, err := r.Cache.GetByPietanzaID(ctx, idPietanza)
		if err != nil {
			// Log error but continue to DB
		} else if found {
			return cached, nil
		}
	}

	// Se non trovato in cache, recupera dal database
	var ricetta models.Ricetta
	err := r.DB.QueryRow(ctx, `
		SELECT id_ricetta, nome, descrizione, id_pietanza, tempo_preparazione, istruzioni
		FROM ricetta
		WHERE id_pietanza = $1
	`, idPietanza).Scan(&ricetta.ID, &ricetta.Nome, &ricetta.Descrizione, &ricetta.IDPietanza, &ricetta.TempoPreparazione, &ricetta.Istruzioni)

	if err != nil {
		return nil, err
	}

	// Salva in cache per le future richieste
	if r.Cache != nil {
		if err := r.Cache.SetByPietanzaID(ctx, idPietanza, &ricetta); err != nil {
			// Log error but continue
		}
	}

	return &ricetta, nil
}

// GetIngredientiByRicettaID restituisce gli ingredienti necessari per una ricetta con le relative quantità
func (r *RicettaRepository) GetIngredientiByRicettaID(ctx context.Context, idRicetta int) ([]models.RicettaIngrediente, error) {
	// Controlla prima nella cache
	if r.Cache != nil {
		cached, found, err := r.Cache.GetIngredientiByRicettaID(ctx, idRicetta)
		if err != nil {
			// Log error but continue to DB
		} else if found {
			return cached, nil
		}
	}

	// Se non trovato in cache, recupera dal database
	rows, err := r.DB.Query(ctx, `
		SELECT ri.id_ricetta, ri.id_ingrediente, ri.quantita
		FROM ricetta_ingrediente ri
		WHERE ri.id_ricetta = $1
	`, idRicetta)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ingredienti []models.RicettaIngrediente
	for rows.Next() {
		var ri models.RicettaIngrediente
		err := rows.Scan(&ri.IDRicetta, &ri.IDIngrediente, &ri.Quantita)
		if err != nil {
			return nil, err
		}
		ingredienti = append(ingredienti, ri)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Salva in cache per le future richieste
	if r.Cache != nil {
		if err := r.Cache.SetIngredientiByRicettaID(ctx, idRicetta, ingredienti); err != nil {
			// Log error but continue
		}
	}

	return ingredienti, nil
}

// VerificaDisponibilitaIngredienti controlla se ci sono ingredienti sufficienti per preparare una pietanza
func (r *RicettaRepository) VerificaDisponibilitaIngredienti(ctx context.Context, idRicetta int, quantitaPietanze int, ingredienteCache *cache.IngredienteCache) (bool, map[int]float64, error) {
	// Recupera gli ingredienti necessari per la ricetta
	ingredienti, err := r.GetIngredientiByRicettaID(ctx, idRicetta)
	if err != nil {
		return false, nil, err
	}

	// Verifica la disponibilità di ciascun ingrediente
	disponibile := true
	ingredientiNecessari := make(map[int]float64)

	for _, ingrediente := range ingredienti {
		// Moltiplica per il numero di pietanze da preparare
		quantitaNecessaria := ingrediente.Quantita * float64(quantitaPietanze)
		ingredientiNecessari[ingrediente.IDIngrediente] = quantitaNecessaria

		// Controlla se l'ingrediente è disponibile in quantità sufficiente - prima dalla cache
		var quantitaDisponibile float64
		var err error

		if ingredienteCache != nil {
			// Tenta di recuperare l'ingrediente dalla cache
			cachedIngr, found, cacheErr := ingredienteCache.GetByID(ctx, ingrediente.IDIngrediente)
			if cacheErr == nil && found {
				quantitaDisponibile = cachedIngr.QuantitaDisponibile
			} else {
				// Se non trovato nella cache, recupera dal database
				err = r.DB.QueryRow(ctx, `
					SELECT quantita_disponibile 
					FROM ingrediente 
					WHERE id_ingrediente = $1
				`, ingrediente.IDIngrediente).Scan(&quantitaDisponibile)
			}
		} else {
			// Se non c'è cache, recupera direttamente dal database
			err = r.DB.QueryRow(ctx, `
				SELECT quantita_disponibile 
				FROM ingrediente 
				WHERE id_ingrediente = $1
			`, ingrediente.IDIngrediente).Scan(&quantitaDisponibile)
		}

		if err != nil {
			return false, nil, err
		}

		if quantitaDisponibile < quantitaNecessaria {
			disponibile = false
			break
		}
	}

	return disponibile, ingredientiNecessari, nil
}

// AggiornaIngredienti aggiorna la quantità degli ingredienti disponibili dopo la preparazione di una pietanza
// Viene eseguito all'interno di una transazione fornita
// Invalida anche la cache degli ingredienti aggiornati
func (r *RicettaRepository) AggiornaIngredienti(ctx context.Context, tx pgx.Tx, ingredientiNecessari map[int]float64, ingredienteCache *cache.IngredienteCache) error {
	for idIngrediente, quantita := range ingredientiNecessari {
		_, err := tx.Exec(ctx, `
			UPDATE ingrediente
			SET quantita_disponibile = quantita_disponibile - $1
			WHERE id_ingrediente = $2
		`, quantita, idIngrediente)
		if err != nil {
			return err
		}

		// Invalida la cache per questo ingrediente
		if ingredienteCache != nil {
			if err := ingredienteCache.InvalidateByID(ctx, idIngrediente); err != nil {
				// Log error but continue
			}
		}
	}

	// Invalida anche la cache degli ingredienti da riordinare, poiché potrebbe essere cambiata
	if ingredienteCache != nil {
		if err := ingredienteCache.InvalidateDaRiordinare(ctx); err != nil {
			// Log error but continue
		}
		if err := ingredienteCache.InvalidateAll(ctx); err != nil {
			// Log error but continue
		}
	}

	return nil
}

// GetRicettaCompletaByPietanzaID restituisce la ricetta completa con ingredienti per una pietanza
func (r *RicettaRepository) GetRicettaCompletaByPietanzaID(ctx context.Context, idPietanza int) (*models.RicettaCompleta, error) {
	// 1. Recupera la ricetta di base
	ricetta, err := r.GetByPietanzaID(ctx, idPietanza)
	if err != nil {
		return nil, err
	}

	// 2. Recupera gli ingredienti associati alla ricetta
	rows, err := r.DB.Query(ctx, `
		SELECT ri.id_ricetta, ri.id_ingrediente, ri.quantita, 
			i.id_ingrediente, i.nome, i.quantita_disponibile, i.unita_misura, i.soglia_riordino
		FROM ricetta_ingrediente ri
		JOIN ingrediente i ON ri.id_ingrediente = i.id_ingrediente
		WHERE ri.id_ricetta = $1
	`, ricetta.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ingredienti []models.IngredienteConQuantita
	for rows.Next() {
		var ing models.IngredienteConQuantita
		var idRicetta, idIngrediente int
		err := rows.Scan(
			&idRicetta, &idIngrediente, &ing.Quantita,
			&ing.Ingrediente.ID, &ing.Ingrediente.Nome, &ing.Ingrediente.QuantitaDisponibile,
			&ing.Ingrediente.UnitaMisura, &ing.Ingrediente.SogliaRiordino,
		)
		if err != nil {
			return nil, err
		}
		ingredienti = append(ingredienti, ing)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	ricettaCompleta := &models.RicettaCompleta{
		Ricetta:     *ricetta,
		Ingredienti: ingredienti,
	}

	return ricettaCompleta, nil
}
