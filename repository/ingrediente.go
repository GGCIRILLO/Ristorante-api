package repository

// type Ingrediente struct {
// 	ID                 int     `json:"id"`
// 	Nome               string  `json:"nome"`
// 	QuantitaDisponibile float64 `json:"quantita_disponibile"`
// 	UnitaMisura        string  `json:"unita_misura"`
// 	SogliaRiordino     float64 `json:"soglia_riordino"`
// }

import (
	"context"
	"ristorante-api/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IngredienteRepository struct {
	DB *pgxpool.Pool
}

func NewIngredienteRepository(db *pgxpool.Pool) *IngredienteRepository {
	return &IngredienteRepository{DB: db}
}

// GetAll restituisce tutti gli ingredienti disponibili
func (r *IngredienteRepository) GetAll(ctx context.Context) ([]models.Ingrediente, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id_ingrediente, nome, quantita_disponibile, unita_misura, soglia_riordino
		FROM ingrediente
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ingredienti []models.Ingrediente
	for rows.Next() {
		var i models.Ingrediente
		err := rows.Scan(&i.ID, &i.Nome, &i.QuantitaDisponibile, &i.UnitaMisura, &i.SogliaRiordino)
		if err != nil {
			return nil, err
		}
		ingredienti = append(ingredienti, i)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ingredienti, nil
}

// GetByID restituisce un ingrediente specifico in base all'ID
func (r *IngredienteRepository) GetByID(ctx context.Context, id int) (*models.Ingrediente, error) {
	var i models.Ingrediente
	err := r.DB.QueryRow(ctx, `
		SELECT id_ingrediente, nome, quantita_disponibile, unita_misura, soglia_riordino
		FROM ingrediente
		WHERE id_ingrediente = $1
	`, id).Scan(&i.ID, &i.Nome, &i.QuantitaDisponibile, &i.UnitaMisura, &i.SogliaRiordino)

	if err != nil {
		return nil, err
	}

	return &i, nil
}

// Create aggiunge un nuovo ingrediente al magazzino
func (r *IngredienteRepository) Create(ctx context.Context, i *models.Ingrediente) error {
	return r.DB.QueryRow(ctx, `
		INSERT INTO ingrediente (nome, quantita_disponibile, unita_misura, soglia_riordino)
		VALUES ($1, $2, $3, $4)
		RETURNING id_ingrediente
	`, i.Nome, i.QuantitaDisponibile, i.UnitaMisura, i.SogliaRiordino).Scan(&i.ID)
}

// Update aggiorna un ingrediente esistente
func (r *IngredienteRepository) Update(ctx context.Context, i *models.Ingrediente) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE ingrediente
		SET nome = $1, quantita_disponibile = $2, unita_misura = $3, soglia_riordino = $4
		WHERE id_ingrediente = $5
	`, i.Nome, i.QuantitaDisponibile, i.UnitaMisura, i.SogliaRiordino, i.ID)
	return err
}

// Delete elimina un ingrediente per ID
func (r *IngredienteRepository) Delete(ctx context.Context, id int) error {
	_, err := r.DB.Exec(ctx, `
		DELETE FROM ingrediente
		WHERE id_ingrediente = $1
	`, id)
	if err != nil {
		return err
	}
	return nil
}

// IngredientiDaRiordinare restituisce gli ingredienti sotto la soglia di riordino
func (r *IngredienteRepository) IngredientiDaRiordinare(ctx context.Context) ([]models.Ingrediente, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id_ingrediente, nome, quantita_disponibile, unita_misura, soglia_riordino
		FROM ingrediente
		WHERE quantita_disponibile < soglia_riordino
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ingredienti []models.Ingrediente
	for rows.Next() {
		var i models.Ingrediente
		err := rows.Scan(&i.ID, &i.Nome, &i.QuantitaDisponibile, &i.UnitaMisura, &i.SogliaRiordino)
		if err != nil {
			return nil, err
		}
		ingredienti = append(ingredienti, i)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ingredienti, nil
}

// Prenota un ingrediente di una certa quantità
func (r *IngredienteRepository) Prenota(ctx context.Context, id int, quantita float64) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE ingrediente
		SET quantita_disponibile = quantita_disponibile - $1
		WHERE id_ingrediente = $2 AND quantita_disponibile >= $1
	`, quantita, id)
	if err != nil {
		return err
	}
	return nil
}

// Rifornisci un ingrediente con una certa quantità
func (r *IngredienteRepository) Rifornisci(ctx context.Context, id int, quantita float64) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE ingrediente
		SET quantita_disponibile = quantita_disponibile + $1
		WHERE id_ingrediente = $2
	`, quantita, id)
	if err != nil {
		return err
	}
	return nil
}
