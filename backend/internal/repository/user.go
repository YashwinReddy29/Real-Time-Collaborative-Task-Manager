package repository

import (
	"context"
	"taskflow/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, name, email, hashedPassword string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (name, email, password) VALUES ($1, $2, $3)
		 RETURNING id, name, email, password, created_at`,
		name, email, hashedPassword,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	return user, err
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, password, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	return user, err
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, password, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	return user, err
}

func (r *UserRepo) CreatePasswordReset(ctx context.Context, token, userID string, expiresAt interface{}) (string, error) {
	_, err := r.db.Exec(ctx,
		`INSERT INTO password_resets (token, user_id, expires_at) VALUES ($1, $2, $3)`,
		token, userID, expiresAt,
	)
	return token, err
}

func (r *UserRepo) ValidateResetToken(ctx context.Context, token string) (string, error) {
	var userID string
	err := r.db.QueryRow(ctx,
		`SELECT user_id FROM password_resets
		 WHERE token = $1 AND used = FALSE AND expires_at > NOW()`,
		token,
	).Scan(&userID)
	return userID, err
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID, hashedPassword string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET password = $1 WHERE id = $2`,
		hashedPassword, userID,
	)
	return err
}

func (r *UserRepo) MarkResetTokenUsed(ctx context.Context, token string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE password_resets SET used = TRUE WHERE token = $1`,
		token,
	)
	return err
}
