package repository

import (
	"context"
	"ristorante-api/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RistoranteRepository struct {
	DB *pgxpool.Pool
}

func NewRistoranteRepository(db *pgxpool.Pool) *RistoranteRepository {
	return &RistoranteRepository{DB: db}
}

func (r *RistoranteRepository) GetAll(ctx context.Context) ([]models.Ristorante, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id_ristorante, nome, numero_tavoli, costo_coperto
		FROM ristorante`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Ristorante
	for rows.Next() {
		var risto models.Ristorante
		if err := rows.Scan(&risto.ID, &risto.Nome, &risto.NumeroTavoli, &risto.CostoCoperto); err != nil {
			return nil, err
		}
		result = append(result, risto)
	}
	return result, nil
}

func (r *RistoranteRepository) GetByID(ctx context.Context, id int) (models.Ristorante, error) {
	var risto models.Ristorante
	err := r.DB.QueryRow(ctx, `
		SELECT id_ristorante, nome, numero_tavoli, costo_coperto
		FROM ristorante WHERE id_ristorante = $1`, id).
		Scan(&risto.ID, &risto.Nome, &risto.NumeroTavoli, &risto.CostoCoperto)
	return risto, err
}

func (r *RistoranteRepository) Create(ctx context.Context, risto *models.Ristorante) error {
	return r.DB.QueryRow(ctx, `
		INSERT INTO ristorante (nome, numero_tavoli, costo_coperto)
		VALUES ($1, $2, $3)
		RETURNING id_ristorante`,
		risto.Nome, risto.NumeroTavoli, risto.CostoCoperto).
		Scan(&risto.ID)
}

func (r *RistoranteRepository) Update(ctx context.Context, id int, risto models.Ristorante) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE ristorante SET nome = $1, numero_tavoli = $2, costo_coperto = $3
		WHERE id_ristorante = $4`,
		risto.Nome, risto.NumeroTavoli, risto.CostoCoperto, id)
	return err
}

func (r *RistoranteRepository) Delete(ctx context.Context, id int) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM ristorante WHERE id_ristorante = $1`, id)
	return err
}

func (r *RistoranteRepository) Exists(ctx context.Context, id int) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ristorante WHERE id_ristorante = $1)`, id).Scan(&exists)
	return exists, err
}
