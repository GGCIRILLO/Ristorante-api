package repository

import (
	"context"
	"ristorante-api/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TavoloRepository struct {
	DB *pgxpool.Pool
}

func NewTavoloRepository(db *pgxpool.Pool) *TavoloRepository {
	return &TavoloRepository{DB: db}
}

func (r *TavoloRepository) GetAll(ctx context.Context) ([]models.Tavolo, error) {
	rows, err := r.DB.Query(ctx, "SELECT id_tavolo, max_posti, stato, id_ristorante FROM tavolo ORDER BY id_tavolo")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tavoli []models.Tavolo
	for rows.Next() {
		var t models.Tavolo
		if err := rows.Scan(&t.ID, &t.MaxPosti, &t.Stato, &t.IDRistorante); err != nil {
			return nil, err
		}
		tavoli = append(tavoli, t)
	}
	return tavoli, nil
}

// GetByID recupera un tavolo per ID
func (r *TavoloRepository) GetByID(ctx context.Context, id int) (models.Tavolo, error) {
	var t models.Tavolo
	err := r.DB.QueryRow(ctx, "SELECT id_tavolo, max_posti, stato, id_ristorante FROM tavolo WHERE id_tavolo = $1", id).
		Scan(&t.ID, &t.MaxPosti, &t.Stato, &t.IDRistorante)
	if err != nil {
		return models.Tavolo{}, err
	}
	return t, nil
}

// CambiaStato cambia lo stato di un tavolo
func (r *TavoloRepository) CambiaStato(ctx context.Context, id int, nuovoStato string) (models.Tavolo, error) {
	var t models.Tavolo
	err := r.DB.QueryRow(ctx, `
		UPDATE tavolo SET stato = $1
		WHERE id_tavolo = $2
		RETURNING id_tavolo, max_posti, stato, id_ristorante`, nuovoStato, id).
		Scan(&t.ID, &t.MaxPosti, &t.Stato, &t.IDRistorante)
	if err != nil {
		return models.Tavolo{}, err
	}
	return t, nil
}

func (r *TavoloRepository) Create(ctx context.Context, t *models.Tavolo) error {
	return r.DB.QueryRow(ctx, `
		INSERT INTO tavolo (max_posti, stato, id_ristorante)
		VALUES ($1, $2, $3)
		RETURNING id_tavolo`,
		t.MaxPosti, t.Stato, t.IDRistorante).
		Scan(&t.ID)
}
func (r *TavoloRepository) Update(ctx context.Context, id int, t models.Tavolo) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE tavolo SET max_posti = $1, stato = $2, id_ristorante = $3
		WHERE id_tavolo = $4`,
		t.MaxPosti, t.Stato, t.IDRistorante, id)
	return err
}

func (r *TavoloRepository) Delete(ctx context.Context, id int) error {
	_, err := r.DB.Exec(ctx, "DELETE FROM tavolo WHERE id_tavolo = $1", id)
	return err
}
