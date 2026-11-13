package repository

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestDatabaseRejectsCrossBoardTask(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not configured")
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.ConnConfig.Database != "taskflow_test" {
		t.Fatal("refusing to modify non-test database")
	}

	ctx := context.Background()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Fatal(err)
	}

	var migrationReady bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM pg_constraint
			WHERE conrelid = 'tasks'::regclass
			  AND conname = 'tasks_column_board_fkey'
		)
	`).Scan(&migrationReady); err != nil {
		t.Fatal(err)
	}
	if !migrationReady {
		t.Fatal("test schema is not initialized; run testdb-init")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	var userID string
	err = tx.QueryRow(ctx, `
		INSERT INTO users (name, email, password)
		VALUES ('Constraint Test', gen_random_uuid()::text || '@example.test', 'test-only')
		RETURNING id
	`).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	var boardA, boardB string

	for _, board := range []*string{&boardA, &boardB} {
		err = tx.QueryRow(ctx, `
			INSERT INTO boards (name, owner_id)
			VALUES ('Constraint Test Board', $1)
			RETURNING id
		`, userID).Scan(board)

		if err != nil {
			t.Fatal(err)
		}
	}

	var columnB string
	err = tx.QueryRow(ctx, `
		INSERT INTO columns (board_id, name)
		VALUES ($1, 'To Do')
		RETURNING id
	`, boardB).Scan(&columnB)
	if err != nil {
		t.Fatal(err)
	}

	// Direct SQL must not bypass the board consistency rule.
	_, err = tx.Exec(ctx, `
		INSERT INTO tasks (board_id, column_id, title)
		VALUES ($1, $2, 'Invalid cross-board task')
	`, boardA, columnB)

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("expected PostgreSQL constraint violation, got %v", err)
	}

	if pgErr.Code != "23503" ||
		pgErr.ConstraintName != "tasks_column_board_fkey" {
		t.Fatalf(
			"unexpected constraint error: code=%s constraint=%s",
			pgErr.Code,
			pgErr.ConstraintName,
		)
	}
}
