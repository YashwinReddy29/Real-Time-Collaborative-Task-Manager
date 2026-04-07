package repository

import (
	"context"
	"taskflow/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ActivityRepo struct {
	db *pgxpool.Pool
}

func NewActivityRepo(db *pgxpool.Pool) *ActivityRepo {
	return &ActivityRepo{db: db}
}

func (r *ActivityRepo) Create(ctx context.Context, boardID, userID, userName, action string) (*models.Activity, error) {
	a := &models.Activity{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO activities (board_id, user_id, user_name, action)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, board_id, user_id, user_name, action, created_at`,
		boardID, userID, userName, action,
	).Scan(&a.ID, &a.BoardID, &a.UserID, &a.UserName, &a.Action, &a.CreatedAt)
	return a, err
}

func (r *ActivityRepo) GetByBoardID(ctx context.Context, boardID string, limit int) ([]models.Activity, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, board_id, user_id, user_name, action, created_at
		 FROM activities WHERE board_id = $1
		 ORDER BY created_at DESC LIMIT $2`,
		boardID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []models.Activity
	for rows.Next() {
		var a models.Activity
		if err := rows.Scan(&a.ID, &a.BoardID, &a.UserID, &a.UserName, &a.Action, &a.CreatedAt); err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}
	return activities, nil
}
