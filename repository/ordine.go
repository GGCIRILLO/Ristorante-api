package repository

import (
	"context"
	"ristorante-api/models"
	"time"

	"github.com/jackc/pgx/v5"
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

// AggiornaCostoTotale aggiorna il costo totale di un ordine
// Se specificato un importo fisso (es. per un menu fisso), viene utilizzato quello,
// altrimenti viene calcolato dalla somma delle pietanze
func (r *OrdineRepository) AggiornaCostoTotale(ctx context.Context, tx pgx.Tx, idOrdine int, importoFisso *float64) error {
	if importoFisso != nil {
		// Aggiorna con l'importo fisso specificato (es. prezzo del menu fisso)
		_, err := tx.Exec(ctx, `
			UPDATE ordine
			SET costo_totale = costo_totale + $1
			WHERE id_ordine = $2
		`, *importoFisso, idOrdine)
		return err
	}

	// Altrimenti ricalcola il totale dalle pietanze dell'ordine (escluse quelle nei menu fissi)
	_, err := tx.Exec(ctx, `
		UPDATE ordine o
		SET costo_totale = (
			SELECT COALESCE(SUM(p.prezzo * d.quantita), 0)
			FROM dettaglio_ordine_pietanza d
			JOIN pietanza p ON d.id_pietanza = p.id_pietanza
			WHERE d.id_ordine = o.id_ordine AND (d.parte_di_menu = false OR d.parte_di_menu IS NULL)
		) + (
			SELECT COALESCE(SUM(m.prezzo), 0)
			FROM dettaglio_ordine_menu d
			JOIN menu_fisso m ON d.id_menu = m.id_menu
			WHERE d.id_ordine = o.id_ordine
		)
		WHERE o.id_ordine = $1
	`, idOrdine)

	return err
}

// CalcolaScontrino calcola lo scontrino per l'ordine di un tavolo
// Prende l'ordine in stato "consegnato" associato al tavolo
// Aggiunge il costo del coperto per ogni persona al costo totale
// Imposta lo stato dell'ordine come "pagato"
func (r *OrdineRepository) CalcolaScontrino(ctx context.Context, idTavolo int) (*models.Scontrino, error) {
	// Inizia una transazione
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Recupera l'ordine in stato "consegnato" per il tavolo specificato
	var ordine models.Ordine
	err = tx.QueryRow(ctx, `
		SELECT id_ordine, id_tavolo, num_persone, data_ordine, stato, id_ristorante, costo_totale
		FROM ordine 
		WHERE id_tavolo = $1 AND stato = 'consegnato'
		ORDER BY data_ordine DESC
		LIMIT 1
	`, idTavolo).Scan(
		&ordine.ID, &ordine.IDTavolo, &ordine.NumPersone, &ordine.DataOrdine,
		&ordine.Stato, &ordine.IDRistorante, &ordine.CostoTotale,
	)
	if err != nil {
		return nil, err
	}

	// 2. Recupera il costo del coperto dal ristorante
	var costoCoperto float64
	err = tx.QueryRow(ctx, `
		SELECT costo_coperto
		FROM ristorante
		WHERE id_ristorante = $1
	`, ordine.IDRistorante).Scan(&costoCoperto)
	if err != nil {
		return nil, err
	}

	// 3. Calcola l'importo totale del coperto
	importoCoperto := costoCoperto * float64(ordine.NumPersone)
	totaleComplessivo := ordine.CostoTotale + importoCoperto

	// 4. Aggiorna lo stato dell'ordine a "pagato"
	dataPagamento := time.Now()
	_, err = tx.Exec(ctx, `
		UPDATE ordine
		SET stato = 'pagato', data_pagamento = $1
		WHERE id_ordine = $2
	`, dataPagamento, ordine.ID)
	if err != nil {
		return nil, err
	}

	// 5. Crea lo scontrino
	scontrino := &models.Scontrino{
		IDOrdine:          ordine.ID,
		IDTavolo:          ordine.IDTavolo,
		DataOrdine:        ordine.DataOrdine,
		CostoTotale:       ordine.CostoTotale,
		NumCoperti:        ordine.NumPersone,
		CostoCoperto:      costoCoperto,
		ImportoCoperto:    importoCoperto,
		TotaleComplessivo: totaleComplessivo,
		DataPagamento:     dataPagamento,
	}

	// Commit della transazione
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return scontrino, nil
}
