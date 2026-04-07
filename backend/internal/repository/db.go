package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}
	return pool, nil
}

func Migrate(pool *pgxpool.Pool) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS boards (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS board_members (
		board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		role TEXT NOT NULL DEFAULT 'member',
		PRIMARY KEY (board_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS columns (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		position INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		column_id UUID NOT NULL REFERENCES columns(id) ON DELETE CASCADE,
		board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
		title TEXT NOT NULL,
		description TEXT DEFAULT '',
		position INTEGER NOT NULL DEFAULT 0,
		assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
		priority TEXT NOT NULL DEFAULT 'medium',
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS activities (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		user_name TEXT NOT NULL,
		action TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_columns_board_id ON columns(board_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_column_id ON tasks(column_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_board_id ON tasks(board_id);
	CREATE INDEX IF NOT EXISTS idx_activities_board_id ON activities(board_id);
	`
	_, err := pool.Exec(context.Background(), schema)
	return err
}

func MigratePasswordResets(pool *pgxpool.Pool) error {
	_, err := pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS password_resets (
			token TEXT PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			expires_at TIMESTAMPTZ NOT NULL,
			used BOOLEAN DEFAULT FALSE
		);
	`)
	return err
}
