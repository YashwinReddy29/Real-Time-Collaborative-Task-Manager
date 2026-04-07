package repository

import (
	"context"
	"taskflow/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepo struct {
	db *pgxpool.Pool
}

func NewTaskRepo(db *pgxpool.Pool) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) GetByBoardID(ctx context.Context, boardID string) ([]models.Task, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, column_id, board_id, title, description, position,
		        assignee_id, priority, created_at, updated_at
		 FROM tasks WHERE board_id = $1 ORDER BY position ASC`,
		boardID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(
			&t.ID, &t.ColumnID, &t.BoardID, &t.Title, &t.Description,
			&t.Position, &t.AssigneeID, &t.Priority, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskRepo) Create(ctx context.Context, columnID, boardID, title, description, priority string, position int) (*models.Task, error) {
	task := &models.Task{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO tasks (column_id, board_id, title, description, priority, position)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, column_id, board_id, title, description, position,
		           assignee_id, priority, created_at, updated_at`,
		columnID, boardID, title, description, priority, position,
	).Scan(
		&task.ID, &task.ColumnID, &task.BoardID, &task.Title, &task.Description,
		&task.Position, &task.AssigneeID, &task.Priority, &task.CreatedAt, &task.UpdatedAt,
	)
	return task, err
}

func (r *TaskRepo) Update(ctx context.Context, id, title, description, priority string) (*models.Task, error) {
	task := &models.Task{}
	err := r.db.QueryRow(ctx,
		`UPDATE tasks SET title=$1, description=$2, priority=$3, updated_at=NOW()
		 WHERE id=$4
		 RETURNING id, column_id, board_id, title, description, position,
		           assignee_id, priority, created_at, updated_at`,
		title, description, priority, id,
	).Scan(
		&task.ID, &task.ColumnID, &task.BoardID, &task.Title, &task.Description,
		&task.Position, &task.AssigneeID, &task.Priority, &task.CreatedAt, &task.UpdatedAt,
	)
	return task, err
}

func (r *TaskRepo) Move(ctx context.Context, taskID, newColumnID string, newPosition int) (*models.Task, error) {
	task := &models.Task{}
	err := r.db.QueryRow(ctx,
		`UPDATE tasks SET column_id=$1, position=$2, updated_at=NOW()
		 WHERE id=$3
		 RETURNING id, column_id, board_id, title, description, position,
		           assignee_id, priority, created_at, updated_at`,
		newColumnID, newPosition, taskID,
	).Scan(
		&task.ID, &task.ColumnID, &task.BoardID, &task.Title, &task.Description,
		&task.Position, &task.AssigneeID, &task.Priority, &task.CreatedAt, &task.UpdatedAt,
	)
	return task, err
}

func (r *TaskRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	return err
}

func (r *TaskRepo) GetByID(ctx context.Context, id string) (*models.Task, error) {
	task := &models.Task{}
	err := r.db.QueryRow(ctx,
		`SELECT id, column_id, board_id, title, description, position,
		        assignee_id, priority, created_at, updated_at
		 FROM tasks WHERE id = $1`, id,
	).Scan(
		&task.ID, &task.ColumnID, &task.BoardID, &task.Title, &task.Description,
		&task.Position, &task.AssigneeID, &task.Priority, &task.CreatedAt, &task.UpdatedAt,
	)
	return task, err
}
