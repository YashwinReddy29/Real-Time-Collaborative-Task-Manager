package models

import "time"

type User struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Board struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	OwnerID     string    `json:"owner_id" db:"owner_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type Column struct {
	ID        string    `json:"id" db:"id"`
	BoardID   string    `json:"board_id" db:"board_id"`
	Name      string    `json:"name" db:"name"`
	Position  int       `json:"position" db:"position"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Task struct {
	ID          string    `json:"id" db:"id"`
	ColumnID    string    `json:"column_id" db:"column_id"`
	BoardID     string    `json:"board_id" db:"board_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Position    int       `json:"position" db:"position"`
	AssigneeID  *string   `json:"assignee_id" db:"assignee_id"`
	Priority    string    `json:"priority" db:"priority"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type BoardMember struct {
	BoardID string `json:"board_id" db:"board_id"`
	UserID  string `json:"user_id" db:"user_id"`
	Role    string `json:"role" db:"role"`
}

type Activity struct {
	ID        string    `json:"id" db:"id"`
	BoardID   string    `json:"board_id" db:"board_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	UserName  string    `json:"user_name" db:"user_name"`
	Action    string    `json:"action" db:"action"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// WebSocket event types
type WSEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	BoardID string      `json:"board_id"`
	UserID  string      `json:"user_id,omitempty"`
}

const (
	EventTaskCreated  = "task.created"
	EventTaskUpdated  = "task.updated"
	EventTaskMoved    = "task.moved"
	EventTaskDeleted  = "task.deleted"
	EventColumnCreated = "column.created"
	EventColumnDeleted = "column.deleted"
	EventUserJoined   = "user.joined"
	EventUserLeft     = "user.left"
	EventActivity     = "activity"
)
