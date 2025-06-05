package repository

import (
	"context"
	"ristorante-api/models"

	"github.com/jackc/pgx/v5/pgxpool"
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
func (r *PietanzaRepository) AddPietanzaToOrdine(ctx context.Context, idOrdine int, idPietanza int, quantita int) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO dettaglio_ordine_pietanza (id_ordine, id_pietanza, quantita)
		VALUES ($1, $2, $3)
		ON CONFLICT (id_ordine, id_pietanza) 
		DO UPDATE SET quantita = dettaglio_ordine_pietanza.quantita + EXCLUDED.quantita
	`, idOrdine, idPietanza, quantita)
	return err
}
