package repository

import (
	"context"
	"ristorante-api/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrdineRepository struct {
	DB *pgxpool.Pool
}

func NewOrdineRepository(db *pgxpool.Pool) *OrdineRepository {
	return &OrdineRepository{DB: db}
}

// Create crea un nuovo ordine e restituisce l'ID e la data dell'ordine
// Il Cameriere può creare un ordine per un tavolo specifico
func (r *OrdineRepository) Create(ctx context.Context, o *models.Ordine) error {
	return r.DB.QueryRow(ctx, `
		INSERT INTO ordine (id_tavolo, num_persone, stato, id_ristorante)
		VALUES ($1, $2, 'in_attesa', $3)
		RETURNING id_ordine, data_ordine, stato
	`, o.IDTavolo, o.NumPersone, o.IDRistorante).Scan(&o.ID, &o.DataOrdine, &o.Stato)
}

// GetAll restituisce tutti gli ordini - utile per il Cuoco
func (r *OrdineRepository) GetAll(ctx context.Context) ([]models.Ordine, error) {
	rows, err := r.DB.Query(ctx, `SELECT id_ordine, id_tavolo, num_persone, data_ordine, stato, id_ristorante, costo_totale FROM ordine`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ordini []models.Ordine
	for rows.Next() {
		var o models.Ordine
		err := rows.Scan(&o.ID, &o.IDTavolo, &o.NumPersone, &o.DataOrdine, &o.Stato, &o.IDRistorante, &o.CostoTotale)
		if err != nil {
			return nil, err
		}
		ordini = append(ordini, o)
	}
	return ordini, nil
}

func (r *OrdineRepository) GetByID(ctx context.Context, id int) (models.Ordine, error) {
	var o models.Ordine
	err := r.DB.QueryRow(ctx, `SELECT id_ordine, id_tavolo, num_persone, data_ordine, stato, id_ristorante, costo_totale FROM ordine WHERE id_ordine = $1`, id).
		Scan(&o.ID, &o.IDTavolo, &o.NumPersone, &o.DataOrdine, &o.Stato, &o.IDRistorante, &o.CostoTotale)
	if err != nil {
		return models.Ordine{}, err
	}
	return o, nil
}

// il Cuoco e il Cameriere possono aggiornare lo stato di un ordine (ad esempio da "in attesa" a "in preparazione" o "completato")
func (r *OrdineRepository) UpdateStato(ctx context.Context, id int, nuovoStato string) (models.Ordine, error) {
	var o models.Ordine
	err := r.DB.QueryRow(ctx, `
		UPDATE ordine SET stato = $1
		WHERE id_ordine = $2
		RETURNING id_ordine, id_tavolo, num_persone, data_ordine, stato, id_ristorante, costo_totale
	`, nuovoStato, id).Scan(&o.ID, &o.IDTavolo, &o.NumPersone, &o.DataOrdine, &o.Stato, &o.IDRistorante, &o.CostoTotale)
	if err != nil {
		return models.Ordine{}, err
	}
	return o, nil
}

// Delete elimina un ordine per ID
func (r *OrdineRepository) Delete(ctx context.Context, id int) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM ordine WHERE id_ordine = $1`, id)
	if err != nil {
		return err
	}
	return nil
}
