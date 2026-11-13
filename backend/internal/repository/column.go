package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"taskflow/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ColumnRepo struct {
	db *pgxpool.Pool
}

func NewColumnRepo(db *pgxpool.Pool) *ColumnRepo {
	return &ColumnRepo{db: db}
}

func (r *ColumnRepo) GetByBoardID(ctx context.Context, boardID string) ([]models.Column, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, board_id, name, position, created_at FROM columns
		 WHERE board_id = $1 ORDER BY position ASC`,
		boardID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []models.Column
	for rows.Next() {
		var c models.Column
		if err := rows.Scan(&c.ID, &c.BoardID, &c.Name, &c.Position, &c.CreatedAt); err != nil {
			return nil, err
		}
		columns = append(columns, c)
	}
	return columns, nil
}

func (r *ColumnRepo) Create(ctx context.Context, boardID, name string, position int) (*models.Column, error) {
	col := &models.Column{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO columns (board_id, name, position) VALUES ($1, $2, $3)
		 RETURNING id, board_id, name, position, created_at`,
		boardID, name, position,
	).Scan(&col.ID, &col.BoardID, &col.Name, &col.Position, &col.CreatedAt)
	return col, err
}

func (r *ColumnRepo) Delete(ctx context.Context, boardID, id string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM columns WHERE id = $1 AND board_id = $2`, id, boardID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
