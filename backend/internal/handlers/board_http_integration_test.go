package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"taskflow/internal/auth"
	"taskflow/internal/middleware"
	"taskflow/internal/repository"
	"taskflow/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestBoardHTTPAuthorization(t *testing.T) {
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
	t.Cleanup(pool.Close)

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

	users := repository.NewUserRepo(pool)
	boards := repository.NewBoardRepo(pool)
	columns := repository.NewColumnRepo(pool)
	tasks := repository.NewTaskRepo(pool)
	activities := repository.NewActivityRepo(pool)

	hub := websocket.NewHub()
	handler := NewBoardHandler(
		boards, columns, tasks, activities,
		users, hub, "http://localhost:3000",
	)

	const secret = "test-only-jwt-secret-not-for-production"

	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group(
		"/api/boards",
		middleware.AuthMiddleware(secret),
	)
	group.GET("/:boardId", handler.GetBoard)
	group.POST("/:boardId/tasks", handler.CreateTask)
	group.PUT("/:boardId/tasks/:taskId", handler.UpdateTask)
	group.PATCH("/:boardId/tasks/:taskId/move", handler.MoveTask)
	group.DELETE("/:boardId/tasks/:taskId", handler.DeleteTask)
	group.DELETE("/:boardId/columns/:columnId", handler.DeleteColumn)

	suffix := uuid.NewString()

	var userA, userB string
	for i, target := range []*string{&userA, &userB} {
		err := pool.QueryRow(ctx, `
			INSERT INTO users (name, email, password)
			VALUES ($1, $2, 'test-password-placeholder')
			RETURNING id
		`,
			fmt.Sprintf("HTTP Test User %d", i),
			fmt.Sprintf("http-%d-%s@example.test", i, suffix),
		).Scan(target)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Cleanup(func() {
		for _, id := range []string{userA, userB} {
			if _, err := pool.Exec(
				context.Background(),
				"DELETE FROM users WHERE id=$1",
				id,
			); err != nil {
				t.Errorf("cleanup: %v", err)
			}
		}
	})

	boardA, err := boards.Create(ctx, "Board A", "", userA)
	if err != nil {
		t.Fatal(err)
	}
	boardB, err := boards.Create(ctx, "Board B", "", userB)
	if err != nil {
		t.Fatal(err)
	}

	getColumn := func(boardID string) string {
		t.Helper()
		cols, err := columns.GetByBoardID(ctx, boardID)
		if err != nil || len(cols) == 0 {
			t.Fatalf("missing columns: %v", err)
		}
		return cols[0].ID
	}

	columnA := getColumn(boardA.ID)
	columnB := getColumn(boardB.ID)

	privateTask, err := tasks.Create(
		ctx, columnB, boardB.ID,
		"Private Task", "", "medium", 0,
	)
	if err != nil {
		t.Fatal(err)
	}

	tokenA, err := auth.GenerateToken(
		userA, "a@example.test", "User A", secret,
	)
	if err != nil {
		t.Fatal(err)
	}

	request := func(
		method, path string,
		body any,
		token string,
	) *httptest.ResponseRecorder {
		t.Helper()

		var encoded []byte
		if body != nil {
			var err error
			encoded, err = json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
		}

		req := httptest.NewRequest(
			method, path, bytes.NewReader(encoded),
		)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	type testCase struct {
		name   string
		method string
		path   string
		body   any
		token  string
		want   int
	}

	prefix := "/api/boards/"

	cases := []testCase{
		{
			name:   "unauthenticated board access",
			method: http.MethodGet,
			path:   prefix + boardA.ID,
			want:   http.StatusUnauthorized,
		},
		{
			name:   "unauthorized board read",
			method: http.MethodGet,
			path:   prefix + boardB.ID,
			token:  tokenA,
			want:   http.StatusForbidden,
		},
		{
			name:   "authorized board read",
			method: http.MethodGet,
			path:   prefix + boardA.ID,
			token:  tokenA,
			want:   http.StatusOK,
		},
		{
			name:   "reject foreign task update",
			method: http.MethodPut,
			path:   prefix + boardA.ID + "/tasks/" + privateTask.ID,
			token:  tokenA,
			body: map[string]any{
				"title":    "Unauthorized Change",
				"priority": "high",
			},
			want: http.StatusNotFound,
		},
		{
			name:   "reject foreign task move",
			method: http.MethodPatch,
			path:   prefix + boardA.ID + "/tasks/" + privateTask.ID + "/move",
			token:  tokenA,
			body: map[string]any{
				"column_id": columnA,
				"position":  1,
			},
			want: http.StatusNotFound,
		},
		{
			name:   "reject foreign task delete",
			method: http.MethodDelete,
			path:   prefix + boardA.ID + "/tasks/" + privateTask.ID,
			token:  tokenA,
			want:   http.StatusNotFound,
		},
		{
			name:   "reject foreign column delete",
			method: http.MethodDelete,
			path:   prefix + boardA.ID + "/columns/" + columnB,
			token:  tokenA,
			want:   http.StatusNotFound,
		},
		{
			name:   "reject foreign column task creation",
			method: http.MethodPost,
			path:   prefix + boardA.ID + "/tasks",
			token:  tokenA,
			body: map[string]any{
				"column_id": columnB,
				"title":     "Unauthorized Task",
			},
			want: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := request(
				tc.method, tc.path,
				tc.body, tc.token,
			)
			if rec.Code != tc.want {
				t.Fatalf(
					"HTTP %d, want %d: %s",
					rec.Code, tc.want, rec.Body.String(),
				)
			}
		})
	}

	// Verify no unauthorized request changed the private task.
	unchanged, err := tasks.GetByID(
		ctx, boardB.ID, privateTask.ID,
	)
	if err != nil {
		t.Fatal(err)
	}
	if unchanged.Title != "Private Task" ||
		unchanged.ColumnID != columnB {
		t.Fatal("cross-board requests modified private task")
	}

	// Verify an authorized user can still create a task.
	rec := request(
		http.MethodPost,
		prefix+boardA.ID+"/tasks",
		map[string]any{
			"column_id": columnA,
			"title":     "Authorized Task",
		},
		tokenA,
	)
	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"authorized create returned %d: %s",
			rec.Code, rec.Body.String(),
		)
	}
}
