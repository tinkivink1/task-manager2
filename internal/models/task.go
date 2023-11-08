package models

import "time"

// Task represents a task entity.
// swagger:model
type Task struct {
	// ID of the task.
    // required: true
    // example: 1
	ID          int       `db:"id" json:"id"`

	// Title of the task.
    // required: true
    // example: Task Title
	Title       string    `db:"title" json:"title"`

	// Description of the task.
    // example: Task Description
	Description string    `db:"description" json:"description"`

	// CreatedAt represents the timestamp when the task was created.
    // required: true
    // example: "2023-11-08T12:00:00Z"
	CreatedAt   time.Time `db:"created_at" json:"created_at"`

	// ManagedAt represents the timestamp when the task was last managed.
    // required: true
    // example: "2023-11-08T12:30:00Z"
	ManagedAt   time.Time `db:"managed_at" json:"managed_at"`

	// UserID represents the ID of the user to whom the task belongs.
    // required: true
    // example: 1
	UserID      int       `db:"user_id" json:"user_id"`
}

func NewTask(id int, title, description string, createdAt, managedAt time.Time, userId int, done bool) Task {
	return Task{
		ID:   id,
		Title: title,
		Description: description,
		CreatedAt: createdAt,
		ManagedAt: managedAt,
		UserID: userId,
	}
}