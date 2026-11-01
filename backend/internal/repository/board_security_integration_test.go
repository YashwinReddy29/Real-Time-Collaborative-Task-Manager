package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCrossBoardIsolation(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not configured")
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.ConnConfig.Database != "taskflow_test" {
		t.Fatal("refusing to test against non-test database")
	}

	ctx := context.Background()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
	})

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

	suffix := uuid.NewString()

	var userA, userB string
	var boardA, boardB string
	var columnA, columnB string

	for i, target := range []*string{&userA, &userB} {
		err := pool.QueryRow(ctx, `
			INSERT INTO users (name, email, password)
			VALUES ($1, $2, $3)
			RETURNING id
		`,
			fmt.Sprintf("Security Test User %d", i),
			fmt.Sprintf("security-%d-%s@example.test", i, suffix),
			"test-only-placeholder",
		).Scan(target)

		if err != nil {
			t.Fatal(err)
		}
	}

	t.Cleanup(func() {
		for _, id := range []string{userA, userB} {
			_, err := pool.Exec(
				context.Background(),
				`DELETE FROM users WHERE id=$1`, id,
			)
			if err != nil {
				t.Errorf("cleanup failed: %v", err)
			}
		}
	})

	for i, item := range []struct {
		owner string
		out   *string
	}{
		{userA, &boardA},
		{userB, &boardB},
	} {
		err := pool.QueryRow(ctx, `
			INSERT INTO boards (name, owner_id)
			VALUES ($1, $2)
			RETURNING id
		`,
			fmt.Sprintf("Security Test Board %d", i),
			item.owner,
		).Scan(item.out)

		if err != nil {
			t.Fatal(err)
		}
	}

	for _, item := range []struct {
		board string
		out   *string
	}{
		{boardA, &columnA},
		{boardB, &columnB},
	} {
		err := pool.QueryRow(ctx, `
			INSERT INTO columns (board_id, name)
			VALUES ($1, 'To Do')
			RETURNING id
		`, item.board).Scan(item.out)

		if err != nil {
			t.Fatal(err)
		}
	}

	tasks := NewTaskRepo(pool)
	columns := NewColumnRepo(pool)

	taskB, err := tasks.Create(
		ctx, columnB, boardB,
		"Private Board B Task", "", "medium", 0,
	)
	if err != nil {
		t.Fatalf("creating fixture task: %v", err)
	}

	assertTaskUnchanged := func(t *testing.T) {
		t.Helper()

		task, err := tasks.GetByID(ctx, boardB, taskB.ID)
		if err != nil {
			t.Fatalf("Board B task disappeared: %v", err)
		}

		if task.Title != "Private Board B Task" {
			t.Fatalf("task unexpectedly changed: %q", task.Title)
		}

		if task.ColumnID != columnB {
			t.Fatalf("task moved across boards: %s", task.ColumnID)
		}
	}

	t.Run("reject task creation in foreign column", func(t *testing.T) {
		_, err := tasks.Create(
			ctx, columnB, boardA,
			"Unauthorized task", "", "medium", 0,
		)
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("expected ErrNoRows, got %v", err)
		}
	})

	t.Run("reject foreign task update", func(t *testing.T) {
		_, err := tasks.Update(
			ctx, boardA, taskB.ID,
			"Compromised", "", "high",
		)
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("expected ErrNoRows, got %v", err)
		}
		assertTaskUnchanged(t)
	})

	t.Run("reject foreign task move", func(t *testing.T) {
		_, err := tasks.Move(
			ctx, boardA, taskB.ID, columnA, 1,
		)
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("expected ErrNoRows, got %v", err)
		}
		assertTaskUnchanged(t)
	})

	t.Run("reject cross-board target column", func(t *testing.T) {
		_, err := tasks.Move(
			ctx, boardB, taskB.ID, columnA, 1,
		)
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("expected ErrNoRows, got %v", err)
		}
		assertTaskUnchanged(t)
	})

	t.Run("foreign task deletion changes no data", func(t *testing.T) {
		if err := tasks.Delete(ctx, boardA, taskB.ID); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("expected ErrNoRows, got %v", err)
		}
		assertTaskUnchanged(t)
	})

	t.Run("foreign column deletion changes no data", func(t *testing.T) {
		if err := columns.Delete(ctx, boardA, columnB); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("expected ErrNoRows, got %v", err)
		}

		var count int
		err := pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM columns
			 WHERE id=$1 AND board_id=$2`,
			columnB, boardB,
		).Scan(&count)

		if err != nil || count != 1 {
			t.Fatalf("foreign column was removed: count=%d err=%v", count, err)
		}

		assertTaskUnchanged(t)
	})

	t.Run("legitimate update still works", func(t *testing.T) {
		updated, err := tasks.Update(
			ctx, boardB, taskB.ID,
			"Authorized Update", "", "medium",
		)
		if err != nil {
			t.Fatal(err)
		}
		if updated.Title != "Authorized Update" {
			t.Fatalf("unexpected title: %s", updated.Title)
		}
	})
}
