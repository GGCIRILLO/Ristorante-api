package repository

import (
	"context"
	"errors"
	"ristorante-api/cache"
	"ristorante-api/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Errori personalizzati
var (
	ErrPietanzaNonDisponibile   = errors.New("la pietanza non è disponibile")
	ErrIngredientiInsufficienti = errors.New("ingredienti insufficienti per preparare la pietanza")
	ErrMenuNonDisponibile       = errors.New("il menu non è disponibile: una o più pietanze non sono disponibili o mancano ingredienti")
)

type PietanzaRepository struct {
	DB *pgxpool.Pool
}

func NewPietanzaRepository(db *pgxpool.Pool) *PietanzaRepository {
	return &PietanzaRepository{DB: db}
}

// GetAll restituisce tutte le pietanze disponibili
func (r *PietanzaRepository) GetAll(ctx context.Context) ([]models.Pietanza, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT p.id_pietanza, p.nome, p.prezzo, p.id_categoria, p.disponibile
		FROM pietanza p
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pietanze []models.Pietanza
	for rows.Next() {
		var p models.Pietanza
		err := rows.Scan(&p.ID, &p.Nome, &p.Prezzo, &p.IDCategoria, &p.Disponibile)
		if err != nil {
			return nil, err
		}
		pietanze = append(pietanze, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pietanze, nil
}

// GetByID restituisce una pietanza specifica in base all'ID
func (r *PietanzaRepository) GetByID(ctx context.Context, id int) (*models.Pietanza, error) {
	var p models.Pietanza
	err := r.DB.QueryRow(ctx, `
		SELECT p.id_pietanza, p.nome, p.prezzo, p.id_categoria, p.disponibile
		FROM pietanza p
		WHERE p.id_pietanza = $1
	`, id).Scan(&p.ID, &p.Nome, &p.Prezzo, &p.IDCategoria, &p.Disponibile)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

// Create crea una nuova pietanza e restituisce l'ID generato
func (r *PietanzaRepository) Create(ctx context.Context, p *models.Pietanza) error {
	return r.DB.QueryRow(ctx, `
		INSERT INTO pietanza (nome, prezzo, id_categoria, disponibile)
		VALUES ($1, $2, $3, $4)
		RETURNING id_pietanza
	`, p.Nome, p.Prezzo, p.IDCategoria, p.Disponibile).Scan(&p.ID)
}

// Update aggiorna una pietanza esistente
func (r *PietanzaRepository) Update(ctx context.Context, p *models.Pietanza) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE pietanza
		SET nome = $1, prezzo = $2, id_categoria = $3, disponibile = $4
		WHERE id_pietanza = $5
	`, p.Nome, p.Prezzo, p.IDCategoria, p.Disponibile, p.ID)

	return err
}

// Delete elimina una pietanza in base all'ID
func (r *PietanzaRepository) Delete(ctx context.Context, id int) error {
	_, err := r.DB.Exec(ctx, "DELETE FROM pietanza WHERE id_pietanza = $1", id)
	return err
}

// Exists verifica se una pietanza esiste in base all'ID
func (r *PietanzaRepository) Exists(ctx context.Context, id int) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pietanza WHERE id_pietanza = $1)", id).Scan(&exists)
	return exists, err
}

// AddPietanzaToOrdine aggiunge una pietanza a un ordine esistente
// Verifica che la pietanza sia disponibile e che ci siano ingredienti sufficienti
// Restituisce un errore se la pietanza non è disponibile o se mancano ingredienti
func (r *PietanzaRepository) AddPietanzaToOrdine(ctx context.Context, idOrdine int, idPietanza int, quantita int, ricettaRepo *RicettaRepository, ingredienteCache *cache.IngredienteCache) error {
	// Inizia una transazione
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	// Rollback in caso di errore
	defer tx.Rollback(ctx)

	// 1. Verifica che la pietanza sia disponibile
	var disponibile bool
	err = tx.QueryRow(ctx, `
		SELECT disponibile
		FROM pietanza
		WHERE id_pietanza = $1
	`, idPietanza).Scan(&disponibile)

	if err != nil {
		return err
	}

	if !disponibile {
		return ErrPietanzaNonDisponibile
	}

	// 2. Recupera la ricetta associata alla pietanza
	ricetta, err := ricettaRepo.GetByPietanzaID(ctx, idPietanza)
	if err != nil {
		return err
	}

	// 3. Verifica la disponibilità degli ingredienti utilizzando la cache
	disponibilitaIngredienti, ingredientiNecessari, err := ricettaRepo.VerificaDisponibilitaIngredienti(ctx, ricetta.ID, quantita, ingredienteCache)
	if err != nil {
		return err
	}

	if !disponibilitaIngredienti {
		return ErrIngredientiInsufficienti
	}

	// 4. Aggiunge la pietanza all'ordine
	_, err = tx.Exec(ctx, `
		INSERT INTO dettaglio_ordine_pietanza (id_ordine, id_pietanza, quantita, parte_di_menu, id_menu)
		VALUES ($1, $2, $3, false, NULL)
		ON CONFLICT (id_ordine, id_pietanza, parte_di_menu, id_menu)
		DO UPDATE SET quantita = dettaglio_ordine_pietanza.quantita + EXCLUDED.quantita
	`, idOrdine, idPietanza, quantita)

	if err != nil {
		return err
	}

	// 5. Aggiorna gli ingredienti e invalida la cache
	err = ricettaRepo.AggiornaIngredienti(ctx, tx, ingredientiNecessari, ingredienteCache)
	if err != nil {
		return err
	}

	// 6. Aggiorna il costo totale dell'ordine
	ordineRepo := OrdineRepository{DB: r.DB}
	err = ordineRepo.AggiornaCostoTotale(ctx, tx, idOrdine, nil)
	if err != nil {
		return err
	}

	// Commit della transazione
	return tx.Commit(ctx)
}

// AddMenuFissoToOrdine aggiunge un menu fisso completo a un ordine
// Verifica che tutte le pietanze del menu siano disponibili e che ci siano ingredienti sufficienti
// Se una pietanza non è disponibile o mancano ingredienti, nessuna pietanza viene aggiunta
func (r *PietanzaRepository) AddMenuFissoToOrdine(ctx context.Context, idOrdine int, idMenu int, ricettaRepo *RicettaRepository, menuRepo *MenuFissoRepository, ingredienteCache *cache.IngredienteCache) error {
	// Inizia una transazione
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	// Rollback in caso di errore
	defer tx.Rollback(ctx)

	// 1. Recupera il menu fisso
	menuFisso, err := menuRepo.GetByID(ctx, idMenu)
	if err != nil {
		return err
	}

	// 2. Recupera tutte le pietanze del menu
	pietanze, err := menuRepo.GetComposizione(ctx, idMenu)
	if err != nil {
		return err
	}

	// Non ci sono pietanze nel menu
	if len(pietanze) == 0 {
		return errors.New("il menu fisso non contiene pietanze")
	}

	// 3. Verifica la disponibilità di tutte le pietanze e degli ingredienti
	// Accumuliamo tutte le risorse necessarie prima di effettuare qualsiasi modifica
	type pietanzaRisorse struct {
		ricetta              *models.Ricetta
		ingredientiNecessari map[int]float64
	}

	risorsePietanze := make(map[int]pietanzaRisorse)

	for _, p := range pietanze {
		// Verifica che la pietanza sia disponibile
		var disponibile bool
		err = tx.QueryRow(ctx, `
			SELECT disponibile
			FROM pietanza
			WHERE id_pietanza = $1
		`, p.ID).Scan(&disponibile)

		if err != nil {
			return err
		}

		if !disponibile {
			return ErrMenuNonDisponibile
		}

		// Recupera la ricetta associata alla pietanza
		ricetta, err := ricettaRepo.GetByPietanzaID(ctx, p.ID)
		if err != nil {
			return err
		}

		// Verifica la disponibilità degli ingredienti (quantità = 1 per ogni pietanza nel menu)
		// Utilizziamo la cache degli ingredienti per migliorare le performance
		disponibilitaIngredienti, ingredientiNecessari, err := ricettaRepo.VerificaDisponibilitaIngredienti(ctx, ricetta.ID, 1, ingredienteCache)
		if err != nil {
			return err
		}

		if !disponibilitaIngredienti {
			return ErrMenuNonDisponibile
		}

		// Salva le risorse necessarie per questa pietanza
		risorsePietanze[p.ID] = pietanzaRisorse{
			ricetta:              ricetta,
			ingredientiNecessari: ingredientiNecessari,
		}
	}

	// 4. Registra il menu fisso nell'ordine
	_, err = tx.Exec(ctx, `
		INSERT INTO dettaglio_ordine_menu (id_ordine, id_menu, quantita)
		VALUES ($1, $2, 1)
		ON CONFLICT (id_ordine, id_menu) 
		DO UPDATE SET quantita = dettaglio_ordine_menu.quantita + 1
	`, idOrdine, idMenu)

	if err != nil {
		return err
	}

	// 5. Aggiungi tutte le pietanze che compongono il menu all'ordine
	for _, p := range pietanze {
		_, err = tx.Exec(ctx, `
			INSERT INTO dettaglio_ordine_pietanza (id_ordine, id_pietanza, quantita, parte_di_menu, id_menu)
			VALUES ($1, $2, 1, true, $3)
			ON CONFLICT (id_ordine, id_pietanza, parte_di_menu, id_menu) 
			DO UPDATE SET quantita = dettaglio_ordine_pietanza.quantita + 1
		`, idOrdine, p.ID, idMenu)

		if err != nil {
			return err
		}

		// Aggiorna gli ingredienti per questa pietanza e invalida la cache
		risorsa := risorsePietanze[p.ID]
		err = ricettaRepo.AggiornaIngredienti(ctx, tx, risorsa.ingredientiNecessari, ingredienteCache)
		if err != nil {
			return err
		}
	}

	// 6. Aggiorna il costo totale dell'ordine con il prezzo del menu fisso
	ordineRepo := OrdineRepository{DB: r.DB}
	err = ordineRepo.AggiornaCostoTotale(ctx, tx, idOrdine, &menuFisso.Prezzo)
	if err != nil {
		return err
	}

	// Commit della transazione
	return tx.Commit(ctx)
}
