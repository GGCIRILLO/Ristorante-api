package repository

import (
	"context"
	"ristorante-api/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MenuFissoRepository struct {
	DB *pgxpool.Pool
}

func NewMenuFissoRepository(db *pgxpool.Pool) *MenuFissoRepository {
	return &MenuFissoRepository{DB: db}
}

// GetAll restituisce tutti i menu fissi disponibili
func (r *MenuFissoRepository) GetAll(ctx context.Context) ([]models.MenuFisso, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT m.id_menu, m.nome, m.prezzo, m.descrizione
		FROM menu_fisso m
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menuFissi []models.MenuFisso
	for rows.Next() {
		var m models.MenuFisso
		err := rows.Scan(&m.ID, &m.Nome, &m.Prezzo, &m.Descrizione)
		if err != nil {
			return nil, err
		}
		menuFissi = append(menuFissi, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return menuFissi, nil
}

// GetByID restituisce un menu fisso specifico in base all'ID
func (r *MenuFissoRepository) GetByID(ctx context.Context, id int) (*models.MenuFisso, error) {
	var m models.MenuFisso
	err := r.DB.QueryRow(ctx, `
		SELECT m.id_menu, m.nome, m.prezzo, m.descrizione
		FROM menu_fisso m
		WHERE m.id_menu = $1
	`, id).Scan(&m.ID, &m.Nome, &m.Prezzo, &m.Descrizione)

	if err != nil {
		return nil, err
	}

	return &m, nil
}

// Create crea un nuovo menu fisso e restituisce l'ID generato
func (r *MenuFissoRepository) Create(ctx context.Context, m *models.MenuFisso) error {
	return r.DB.QueryRow(ctx, `
		INSERT INTO menu_fisso (nome, prezzo, descrizione)
		VALUES ($1, $2, $3)
		RETURNING id_menu
	`, m.Nome, m.Prezzo, m.Descrizione).Scan(&m.ID)
}

// Update aggiorna un menu fisso esistente
func (r *MenuFissoRepository) Update(ctx context.Context, m *models.MenuFisso) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE menu_fisso
		SET nome = $1, prezzo = $2, descrizione = $3
		WHERE id_menu = $4
	`, m.Nome, m.Prezzo, m.Descrizione, m.ID)

	return err
}

// Delete elimina un menu fisso in base all'ID
func (r *MenuFissoRepository) Delete(ctx context.Context, id int) error {
	_, err := r.DB.Exec(ctx, "DELETE FROM menu_fisso WHERE id_menu = $1", id)
	return err
}

// Exists verifica se un menu fisso esiste in base all'ID
func (r *MenuFissoRepository) Exists(ctx context.Context, id int) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM menu_fisso WHERE id_menu = $1)", id).Scan(&exists)
	return exists, err
}

// GetComposizione restituisce tutte le pietanze associate a un menu fisso
func (r *MenuFissoRepository) GetComposizione(ctx context.Context, idMenu int) ([]models.Pietanza, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT p.id_pietanza, p.nome, p.prezzo, p.id_categoria, p.disponibile
		FROM pietanza p
		JOIN composizione_menu_fisso c ON p.id_pietanza = c.id_pietanza
		WHERE c.id_menu = $1
	`, idMenu)
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

// AddPietanzaToMenu aggiunge una pietanza a un menu fisso
func (r *MenuFissoRepository) AddPietanzaToMenu(ctx context.Context, idMenu int, idPietanza int) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO composizione_menu_fisso (id_menu, id_pietanza)
		VALUES ($1, $2)
		ON CONFLICT (id_menu, id_pietanza) DO NOTHING
	`, idMenu, idPietanza)
	return err
}

// RemovePietanzaFromMenu rimuove una pietanza da un menu fisso
func (r *MenuFissoRepository) RemovePietanzaFromMenu(ctx context.Context, idMenu int, idPietanza int) error {
	_, err := r.DB.Exec(ctx, `
		DELETE FROM composizione_menu_fisso
		WHERE id_menu = $1 AND id_pietanza = $2
	`, idMenu, idPietanza)
	return err
}
