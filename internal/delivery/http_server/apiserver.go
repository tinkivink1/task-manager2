package apiserver

import (
	"TaskManager/internal/models"
	"errors"

	// "TaskManager/internal/storage/postgres"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Storage interface {
	CreateUser(username, password string) (int, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByUsernameAndPassword(username, password string) (*models.User, error)

	GetTasks(userID int) ([]models.Task, error)
	CreateTask(userID int, title, description string, scheduledFor time.Time) error
	GetTaskByID(userID, taskID int) (*models.Task, error)
	UpdateTask(userID int, task *models.Task) error
	DeleteTask(userID, taskID int) error
}

type Cache interface {
	Get(key string) (string, error)
	Set(key string, value string, expiration time.Duration) error
	Del(key, value string, expiration time.Duration) error
}

type APIServer struct {
	config  *Config
	logger  *logrus.Logger
	router  *gin.Engine
	storage Storage
	cache   Cache
}

func New(config *Config) (*APIServer, error) {
	if config == nil {
		return nil, errors.New("Config is nil")
	}

	return &APIServer{
		config: config,
		logger: logrus.New(),
		router: gin.Default(),
	}, nil
}

func (s *APIServer) Start() error {
	if err := s.configureLogger(); err != nil {
		return err
	}

	s.configureRouter()

	s.logger.Info("Starting API server")

	if s.config != nil {
		s.logger.Info(s.config)
	}

	return s.router.Run(s.config.BindAddr)
}

func (s *APIServer) UseDB(storage Storage) error {
	s.storage = storage

	return nil
}

func (s *APIServer) UseCache(cache Cache) error {
	s.cache = cache

	return nil
}

func (s *APIServer) configureLogger() error {
	level, err := logrus.ParseLevel(s.config.LogLevel)

	if err != nil {
		return err
	}

	s.logger.SetLevel(level)
	s.logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "02.01.2006 15:04:05",
		FullTimestamp:   true,
	})

	return nil
}

func (s *APIServer) configureRouter() {
	// s.router.Use(s.LoggerMiddleware())

	s.router.LoadHTMLGlob("static/*")

	publicGroup := s.router.Group("/")
	{
		publicGroup.GET("/", s.handleIndex)
		publicGroup.GET("/index", s.handleIndex)
		publicGroup.POST("/login", s.handleLogin)
		publicGroup.POST("/register", s.handleRegister)
	}

	privateGroup := s.router.Group("/tasks")
	privateGroup.Use(s.AuthMiddleware())
	{
		privateGroup.GET("", s.handleGetTasks)
		privateGroup.POST("", s.handleCreateTask)
		privateGroup.GET("/:id", s.handleGetTask)
		privateGroup.PUT("/:id", s.handleUpdateTask)
		privateGroup.DELETE("/:id", s.handleDeleteTask)
	}

	if s.config.Caching {
		privateGroup.Use(s.CacheMiddleware())
	}
}

func (s *APIServer) handleIndex(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.HTML(http.StatusOK, "index.html", nil)
}

// @Summary Handling login
// @Description Handling login using given login and password
// @Accept json
// @Produce json
// @Success 200 {object} TokenResponse
// @Failure 400,401,500 {object} ErrorResponse
// @Router /login [post]
func (s *APIServer) handleLogin(ctx *gin.Context) {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&loginData); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{"Invalid login data"})
		return
	}

	user, err := s.storage.GetUserByUsernameAndPassword(loginData.Username, loginData.Password)
	if err != nil || user == nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{"Invalid username or password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{"Failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, TokenResponse{tokenString})
}

// @Summary Handling user registration
// @Description Handling user registration using given username and password
// @Accept json
// @Produce json
// @Success 200 {object} TokenResponse "Successful response with an access token"
// @Failure 400,409,500 {object} ErrorResponse "Error response with details"
// @Router /register [post]
func (s *APIServer) handleRegister(ctx *gin.Context) {
	var registrationData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&registrationData); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{"Invalid registration data"})
		return
	}

	existingUser, err := s.storage.GetUserByUsername(registrationData.Username)
	if err == nil || existingUser != nil {
		ctx.JSON(http.StatusConflict, ErrorResponse{"Username already exists"})
		s.logger.Info(err)
		return
	}

	userID, err := s.storage.CreateUser(registrationData.Username, registrationData.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{"Failed to create user"})
		s.logger.Info(err)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{"Failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, TokenResponse{tokenString})
}

// @Summary Handling fetching tasks
// @Description Handling the request to fetch tasks for the authenticated user
// @Produce json
// @Param userID path int true "User ID"
// @Success 200 {array} models.Task "List of tasks"
// @Failure 400,401,500 {object} ErrorResponse "Error response with details"
// @Router /tasks [get]
func (s *APIServer) handleGetTasks(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")

	tasks, err := s.storage.GetTasks(int(userID.(float64)))
	if err != nil {
		s.logger.Error("Error fetching tasks: ", err)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{"Failed to fetch tasks"})
		return
	}

	ctx.JSON(http.StatusOK, tasks)
}

// @Summary Handling task creation
// @Description Handling the request to create a task for the authenticated user
// @Accept json
// @Produce json
// @Param input body models.Task true "Task data"
// @Success 200 {object} models.Task "Created task"
// @Failure 400,401,500 {object} ErrorResponse "Error response with details"
// @Router /tasks [post]
func (s *APIServer) handleCreateTask(ctx *gin.Context) {
	var task models.Task
	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{"Invalid task data"})
		return
	}

	userID, ok := ctx.Get("userID")

	if !ok {
		s.logger.Error("No userID")
		ctx.JSON(http.StatusBadRequest, ErrorResponse{"Wrong user id"})
	}
	err := s.storage.CreateTask(int(userID.(float64)), task.Title, task.Description, task.ScheduledFor)
	if err != nil {
		s.logger.Error("Error creating task: ", err)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{"Failed to create task"})
		return
	}

	ctx.JSON(http.StatusOK, task)
}

// @Summary Handling fetching a task
// @Description Handling the request to fetch a specific task for the authenticated user
// @Produce json
// @Param userID path int true "User ID"
// @Param taskID path int true "Task ID"
// @Success 200 {object} models.Task "Fetched task"
// @Failure 400,401,404 {object} ErrorResponse "Error response with details"
// @Router /tasks/{userID}/{taskID} [get]
func (s *APIServer) handleGetTask(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("userID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{"Invalid user ID"})
		return
	}
	taskID, err := strconv.Atoi(ctx.Param("taskID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{"Invalid task ID"})
		return
	}

	task, err := s.storage.GetTaskByID(userID, taskID)
	if err != nil {
		s.logger.Error("Error fetching task: ", err)
		ctx.JSON(http.StatusNotFound, ErrorResponse{"Task not found"})
		return
	}

	ctx.JSON(http.StatusOK, task)
}

// @Summary Handling updating a task
// @Description Handling the request to update a specific task for the authenticated user
// @Accept json
// @Produce json
// @Param userID path int true "User ID"
// @Param taskID path int true "Task ID"
// @Param input body models.Task true "Updated task data"
// @Success 200 {object} StatusResponse "Task updated successfully"
// @Failure 400,401,500 {object} ErrorResponse "Error response with details"
// @Router /tasks/{userID}/{taskID} [put]
func (s *APIServer) handleUpdateTask(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("userID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{"Invalid user ID"})
		return
	}
	taskID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{"Invalid task ID"})
		return
	}

	var task models.Task
	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{"Invalid task data"})
		return
	}

	task.ID = taskID

	if err := s.storage.UpdateTask(userID, &task); err != nil {
		s.logger.Error("Error updating task: ", err)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{"Failed to update task"})
		return
	}

	ctx.JSON(http.StatusOK, StatusResponse{"Task updated successfully"})
}

// @Summary Handling deleting a task
// @Description Handling the request to delete a specific task for the authenticated user
// @Produce json
// @Param userID path int true "User ID"
// @Param taskID path int true "Task ID"
// @Success 200 {object} ErrorResponse "Task deleted successfully"
// @Failure 400,401,500 {object} ErrorResponse "Error response with details"
// @Router /tasks/{userID}/{taskID} [delete]
func (s *APIServer) handleDeleteTask(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("userID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{"Invalid user ID"})
		return
	}
	taskID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{"Invalid task ID"})
		return
	}

	if err := s.storage.DeleteTask(userID, taskID); err != nil {
		s.logger.Error("Error deleting task: ", err)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{"Failed to delete task"})
		return
	}

	ctx.JSON(http.StatusOK, ErrorResponse{"Task deleted successfully"})
}
