package postgres

import (
	"TaskManager/internal/models"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Storage struct {
	config *Config
	db     *sqlx.DB
}

func New(config *Config) *Storage {
	return &Storage{
		config: config,
	}
}

func (s *Storage) Open() error {
	db, err := sqlx.Open("postgres", s.config.DataBaseURL)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	s.db = db

	return nil
}

func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) CreateUser(username, password string) (int, error) {
	var userID int
	err := s.db.QueryRow("INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id", username, password).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (s *Storage) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := s.db.Get(&user, "SELECT * FROM users WHERE username=$1", username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Storage) GetUserByUsernameAndPassword(username, password string) (*models.User, error) {
	var user models.User
	err := s.db.Get(&user, "SELECT * FROM users WHERE username=$1 AND password=$2", username, password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Storage) GetTasks(userID int) ([]models.Task, error) {
	var tasks []models.Task
	err := s.db.Select(&tasks, "SELECT * FROM tasks WHERE user_id=$1", userID)
	return tasks, err
}

func (s *Storage) CreateTask(userID int, title, description string, scheduledFor time.Time) error {
	_, err := s.db.Exec("INSERT INTO tasks (title, description, created_at, scheduled_for, user_id) VALUES ($1, $2, $3, $4, $5)",
		title, description, time.Now(), scheduledFor, userID)
	return err
}

func (s *Storage) GetTaskByID(userID, taskID int) (*models.Task, error) {
	var task models.Task
	err := s.db.Get(&task, "SELECT * FROM tasks WHERE id=$1 AND user_id=$2", taskID, userID)
	return &task, err
}

func (s *Storage) UpdateTask(userID int, task *models.Task) error {
	_, err := s.db.Exec("UPDATE tasks SET title=$1, description=$2 WHERE id=$3 AND user_id=$4",
		task.Title, task.Description, task.ID, userID)
	return err
}

func (s *Storage) DeleteTask(userID, taskID int) error {
	_, err := s.db.Exec("DELETE FROM tasks WHERE id=$1 AND user_id=$2", taskID, userID)
	return err
}
