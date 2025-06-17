package repository

import (
	"context"
	"ristorante-api/models"

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
	rows, err := r.DB.Query(ctx, `SELECT id_ordine, id_tavolo, num_persone, data_ordine, stato, id_ristorante, costo_totale FROM ordine WHERE stato != 'pagato' AND stato !='consegnato' ORDER BY data_ordine ASC`)
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

	// Altrimenti ricalcola il totale dalle pietanze dell'ordine
	_, err := tx.Exec(ctx, `
		UPDATE ordine o
		SET costo_totale = (
			-- Pietanze normali (non parte di menu fisso)
			SELECT COALESCE(SUM(p.prezzo * d.quantita), 0)
			FROM dettaglio_ordine_pietanza d
			JOIN pietanza p ON d.id_pietanza = p.id_pietanza
			WHERE d.id_ordine = o.id_ordine AND (d.parte_di_menu = false OR d.parte_di_menu IS NULL)
		) + (
			-- Menu fissi (calcolati dai loro ID distinti)
			SELECT COALESCE(SUM(m.prezzo), 0)
			FROM (
				SELECT DISTINCT id_menu 
				FROM dettaglio_ordine_pietanza 
				WHERE id_ordine = o.id_ordine AND parte_di_menu = true AND id_menu IS NOT NULL
			) AS menu_ids
			JOIN menu_fisso m ON menu_ids.id_menu = m.id_menu
		)
		WHERE o.id_ordine = $1
	`, idOrdine)

	return err
}

// CalcolaScontrino calcola lo scontrino per l'ordine di un tavolo
// Prende l'ordine in stato "consegnato" associato al tavolo
// Aggiunge il costo del coperto per ogni persona al costo totale
// Non modifica lo stato dell'ordine (il pagamento verrà gestito separatamente)
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
		// Verifica se l'errore è dovuto all'assenza di righe
		if err == pgx.ErrNoRows {
			return nil, &models.ErrOrdineNonTrovato{
				IDTavolo:        idTavolo,
				StatoRichiesto:  "consegnato",
				MessaggioErrore: "Nessun ordine in stato 'consegnato' trovato per il tavolo specificato",
			}
		}
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

	// 4. Crea lo scontrino
	scontrino := &models.Scontrino{
		IDOrdine:          ordine.ID,
		IDTavolo:          ordine.IDTavolo,
		DataOrdine:        ordine.DataOrdine,
		CostoTotale:       ordine.CostoTotale,
		NumCoperti:        ordine.NumPersone,
		CostoCoperto:      costoCoperto,
		ImportoCoperto:    importoCoperto,
		TotaleComplessivo: totaleComplessivo,
	}

	// Commit della transazione
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return scontrino, nil
}

// GetOrdineCompleto recupera un ordine per ID inclusi tutti i dettagli
// delle pietanze e dei menu fissi associati
func (r *OrdineRepository) GetOrdineCompleto(ctx context.Context, id int) (*models.OrdineCompleto, error) {
	// 1. Recupera l'ordine base
	ordine, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Recupera tutte le pietanze dell'ordine con i dettagli
	rows, err := r.DB.Query(ctx, `
		SELECT 
			d.id_dettaglio, d.id_ordine, d.id_pietanza, d.quantita, 
			d.parte_di_menu, d.id_menu,
			p.id_pietanza, p.nome, p.prezzo, p.id_categoria, p.disponibile
		FROM dettaglio_ordine_pietanza d
		JOIN pietanza p ON d.id_pietanza = p.id_pietanza
		WHERE d.id_ordine = $1
		ORDER BY d.id_menu NULLS FIRST, d.id_dettaglio
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dettagliPietanze []models.DettaglioPietanza
	menuMap := make(map[int][]models.DettaglioPietanza)

	for rows.Next() {
		var dettaglio models.DettaglioPietanza
		var idMenu *int
		err := rows.Scan(
			&dettaglio.ID, &dettaglio.IDOrdine, &dettaglio.Pietanza.ID, &dettaglio.Quantita,
			&dettaglio.ParteDiMenu, &idMenu,
			&dettaglio.Pietanza.ID, &dettaglio.Pietanza.Nome, &dettaglio.Pietanza.Prezzo,
			&dettaglio.Pietanza.IDCategoria, &dettaglio.Pietanza.Disponibile,
		)
		if err != nil {
			return nil, err
		}

		dettaglio.IDMenu = idMenu

		// Se è parte di un menu, aggiungilo alla mappa dei menu
		if dettaglio.ParteDiMenu && idMenu != nil {
			menuMap[*idMenu] = append(menuMap[*idMenu], dettaglio)
		} else {
			// Altrimenti aggiungilo come pietanza indipendente
			dettagliPietanze = append(dettagliPietanze, dettaglio)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// 3. Per ogni menu trovato, recupera i dettagli del menu fisso
	var dettagliMenuFissi []models.DettaglioMenuFisso
	for idMenu, pietanzeMenu := range menuMap {
		// Recupera i dettagli del menu fisso
		var menu models.MenuFisso
		err := r.DB.QueryRow(ctx, `
			SELECT id_menu, nome, prezzo, descrizione
			FROM menu_fisso
			WHERE id_menu = $1
		`, idMenu).Scan(&menu.ID, &menu.Nome, &menu.Prezzo, &menu.Descrizione)
		if err != nil {
			return nil, err
		}

		dettaglioMenu := models.DettaglioMenuFisso{
			Menu:     menu,
			Pietanze: pietanzeMenu,
		}
		dettagliMenuFissi = append(dettagliMenuFissi, dettaglioMenu)
	}

	// 4. Componi l'ordine completo
	ordineCompleto := &models.OrdineCompleto{
		Ordine:    ordine,
		Pietanze:  dettagliPietanze,
		MenuFissi: dettagliMenuFissi,
	}

	return ordineCompleto, nil
}

// GetAllOrdiniCompleti recupera tutti gli ordini con i dettagli completi
func (r *OrdineRepository) GetAllOrdiniCompleti(ctx context.Context) ([]*models.OrdineCompleto, error) {
	// 1. Recupera tutti gli ordini base
	ordini, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var ordiniCompleti []*models.OrdineCompleto

	// 2. Per ogni ordine, recupera i dettagli completi
	for _, ordine := range ordini {
		ordineCompleto, err := r.GetOrdineCompleto(ctx, ordine.ID)
		if err != nil {
			return nil, err
		}
		ordiniCompleti = append(ordiniCompleti, ordineCompleto)
	}

	return ordiniCompleti, nil
}
