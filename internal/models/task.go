package models

import "time"

// Task represents a task in the system.
// @Summary Task details
// @Description Task details with ID, title, description, creation time, management time, and associated user ID.
// @ID Task
// @Produce json
// @Param id path int true "Task ID"
// @Param title body string true "Task title"
// @Param description body string true "Task description"
// @Param created_at body string true "Task creation time in RFC3339 format"
// @Param scheduled_for body string true "Task management time in RFC3339 format"
// @Param user_id body int true "User ID associated with the task"
type Task struct {
	ID           int       `db:"id" json:"id"`
	Title        string    `db:"title" json:"title"`
	Description  string    `db:"description" json:"description"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	ScheduledFor time.Time `db:"scheduled_for" json:"scheduled_for"`
	UserID       int       `db:"user_id" json:"user_id"`
}

func NewTask(id int, title, description string, createdAt, scheduledFor time.Time, userId int, done bool) Task {
	return Task{
		ID:           id,
		Title:        title,
		Description:  description,
		CreatedAt:    createdAt,
		ScheduledFor: scheduledFor,
		UserID:       userId,
	}
}
