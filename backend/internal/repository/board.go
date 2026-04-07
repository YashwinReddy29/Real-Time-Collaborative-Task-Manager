package repository

import (
	"context"
	"taskflow/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BoardRepo struct {
	db *pgxpool.Pool
}

func NewBoardRepo(db *pgxpool.Pool) *BoardRepo {
	return &BoardRepo{db: db}
}

func (r *BoardRepo) Create(ctx context.Context, name, description, ownerID string) (*models.Board, error) {
	board := &models.Board{}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO boards (name, description, owner_id) VALUES ($1, $2, $3)
		 RETURNING id, name, description, owner_id, created_at`,
		name, description, ownerID,
	).Scan(&board.ID, &board.Name, &board.Description, &board.OwnerID, &board.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Add owner as member
	_, err = tx.Exec(ctx,
		`INSERT INTO board_members (board_id, user_id, role) VALUES ($1, $2, 'owner')`,
		board.ID, ownerID,
	)
	if err != nil {
		return nil, err
	}

	// Create default columns
	defaultColumns := []string{"To Do", "In Progress", "Done"}
	for i, col := range defaultColumns {
		_, err = tx.Exec(ctx,
			`INSERT INTO columns (board_id, name, position) VALUES ($1, $2, $3)`,
			board.ID, col, i,
		)
		if err != nil {
			return nil, err
		}
	}

	return board, tx.Commit(ctx)
}

func (r *BoardRepo) GetByUserID(ctx context.Context, userID string) ([]models.Board, error) {
	rows, err := r.db.Query(ctx,
		`SELECT b.id, b.name, b.description, b.owner_id, b.created_at
		 FROM boards b
		 JOIN board_members bm ON b.id = bm.board_id
		 WHERE bm.user_id = $1
		 ORDER BY b.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boards []models.Board
	for rows.Next() {
		var b models.Board
		if err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.OwnerID, &b.CreatedAt); err != nil {
			return nil, err
		}
		boards = append(boards, b)
	}
	return boards, nil
}

func (r *BoardRepo) GetByID(ctx context.Context, id string) (*models.Board, error) {
	board := &models.Board{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, owner_id, created_at FROM boards WHERE id = $1`,
		id,
	).Scan(&board.ID, &board.Name, &board.Description, &board.OwnerID, &board.CreatedAt)
	return board, err
}

func (r *BoardRepo) IsMember(ctx context.Context, boardID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM board_members WHERE board_id = $1 AND user_id = $2`,
		boardID, userID,
	).Scan(&count)
	return count > 0, err
}

func (r *BoardRepo) GetMembers(ctx context.Context, boardID string) ([]models.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT u.id, u.name, u.email, u.created_at
		 FROM users u
		 JOIN board_members bm ON u.id = bm.user_id
		 WHERE bm.board_id = $1`,
		boardID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, u)
	}
	return members, nil
}

func (r *BoardRepo) AddMember(ctx context.Context, boardID, userID string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO board_members (board_id, user_id, role) VALUES ($1, $2, 'member')
		 ON CONFLICT (board_id, user_id) DO NOTHING`,
		boardID, userID,
	)
	return err
}
