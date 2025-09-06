package domain

import (
	"time"
)

type Task struct {
	ID          string     `json:"id" db:"id"`
	TeamID      string     `json:"team_id" db:"team_id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Status      TaskStatus `json:"status" db:"status"`
	Priority    Priority   `json:"priority" db:"priority"`
	AssigneeID  *string    `json:"assignee_id,omitempty" db:"assignee_id"`
	CreatedBy   string     `json:"created_by" db:"created_by"`
	DueDate     *time.Time `json:"due_date,omitempty" db:"due_date"`
	Tags        []string   `json:"tags,omitempty"`
	Attachments []string   `json:"attachments,omitempty"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusReview     TaskStatus = "review"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

type TaskComment struct {
	ID        string    `json:"id" db:"id"`
	TaskID    string    `json:"task_id" db:"task_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type TaskActivity struct {
	ID          string       `json:"id" db:"id"`
	TaskID      string       `json:"task_id" db:"task_id"`
	UserID      string       `json:"user_id" db:"user_id"`
	Action      string       `json:"action" db:"action"`
	Description string       `json:"description" db:"description"`
	Metadata    interface{}  `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
}

type CreateTask struct {
	TeamID      string     `json:"team_id" validate:"required"`
	Title       string     `json:"title" validate:"required,min=1,max=200"`
	Description string     `json:"description" validate:"max=2000"`
	Status      TaskStatus `json:"status" validate:"omitempty,oneof=todo in_progress review done cancelled"`
	Priority    Priority   `json:"priority" validate:"required,oneof=low medium high urgent"`
	AssigneeID  *string    `json:"assignee_id,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Tags        []string   `json:"tags,omitempty" validate:"dive,min=1,max=50"`
}

type UpdateTask struct {
	Title       string     `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description string     `json:"description,omitempty" validate:"omitempty,max=2000"`
	Status      TaskStatus `json:"status,omitempty" validate:"omitempty,oneof=todo in_progress review done cancelled"`
	Priority    Priority   `json:"priority,omitempty" validate:"omitempty,oneof=low medium high urgent"`
	AssigneeID  *string    `json:"assignee_id,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Tags        []string   `json:"tags,omitempty" validate:"omitempty,dive,min=1,max=50"`
}

type CreateTaskComment struct {
	TaskID  string `json:"task_id" validate:"required"`
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

type TaskFilter struct {
	TeamID     string      `json:"team_id"`
	Status     *TaskStatus `json:"status,omitempty"`
	Priority   *Priority   `json:"priority,omitempty"`
	AssigneeID *string     `json:"assignee_id,omitempty"`
	CreatedBy  *string     `json:"created_by,omitempty"`
	Search     string      `json:"search,omitempty"`
	Tags       []string    `json:"tags,omitempty"`
	FromDate   *time.Time  `json:"from_date,omitempty"`
	ToDate     *time.Time  `json:"to_date,omitempty"`
	Limit      int         `json:"limit,omitempty"`
	Offset     int         `json:"offset,omitempty"`
}