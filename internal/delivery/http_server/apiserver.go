package apiserver

import (
	"TaskManager/internal/models"
	"TaskManager/internal/storage/postgres"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// MessageResponse represents a response message.
// swagger:response messageResponse
type MessageResponse struct {
    // The message string.
    // Example: Task deleted successfully.
    Message string `json:"message"`
}

// ErrorResponse represents an error response.
// swagger:response errorResponse
type ErrorResponse struct {
    // The error message.
    // Example: Invalid task ID.
    Error string `json:"error"`
}

// APIServer represents the API server.
type APIServer struct{
	config *Config
	logger *logrus.Logger
	router *gin.Engine
	storage *postgres.Storage

}

func New(config *Config) *APIServer{
	return &APIServer{
		config: config,
		logger: logrus.New(),
		router: gin.Default(),
	}
}

func (s *APIServer) Start() error {
	if err := s.configureLogger(); err != nil{ 
		return err
	}
	
	s.configureRouter()

	if err := s.configureStore(); err != nil{ 
		return err
	}

	s.logger.Info("Starting API server")

	return s.router.Run(s.config.BindAddr)
}

func (s *APIServer) configureLogger() error {
	level, err := logrus.ParseLevel(s.config.LogLevel)

	if err != nil {
		return err
	}

	s.logger.SetLevel(level)
	s.logger.SetFormatter( &logrus.TextFormatter{
		ForceColors: true,
		TimestampFormat : "02.01.2006 15:04:05",
		FullTimestamp:true,
	})

	return nil
}

func (s *APIServer) configureRouter() {
	// s.router.Use(s.LoggerMiddleware())

	s.router.LoadHTMLGlob("static/*")

	s.router.GET("/", s.handleIndex)
	s.router.GET("/index", s.handleIndex)
		
	s.router.POST("/login", s.handleLogin)
	s.router.POST("/register", s.handleRegister)

	s.router.Use(s.AuthMiddleware()).GET("/tasks", s.handleGetTasks)
	s.router.Use(s.AuthMiddleware()).POST("/tasks", s.handleCreateTask)
	s.router.Use(s.AuthMiddleware()).GET("/tasks/:id", s.handleGetTask)
	s.router.Use(s.AuthMiddleware()).PUT("/tasks/:id", s.handleUpdateTask)
	s.router.Use(s.AuthMiddleware()).DELETE("/tasks/:id", s.handleDeleteTask)

}

func (s *APIServer) configureStore() error {
	store := postgres.New(s.config.Storage)
	if err := store.Open(); err != nil { 
		return err
	}

	s.storage = store
	return nil
}

// func (s *APIServer) LoggerMiddleware() gin.HandlerFunc {
//     return func(ctx *gin.Context) {
//         start := time.Now()

//         // Вызываем следующий обработчик запроса
//         ctx.Next()

//         // Завершаем запись в логах после обработки запроса
//         latency := time.Since(start)
//         clientIP := ctx.ClientIP()
//         method := ctx.Request.Method
//         path := ctx.Request.URL.Path
//         status := ctx.Writer.Status()

//         s.logger.WithFields(logrus.Fields{
//             "timestamp": time.Now().Format("02/01/2006 15:04:05"),
//             "latency":   latency,
//             "clientIP":  clientIP,
//             "method":    method,
//             "path":      path,
//             "status":    status,
//         }).Info("Request processed")
//     }
// }

func (s *APIServer) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader("Authorization")
		if authorizationHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			ctx.Abort()
			return
		}

		tokenString := strings.Split(authorizationHeader, " ")[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(s.config.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			ctx.Abort()
			return
		}

		ctx.Set("userID", claims["sub"])
		ctx.Next()
	}
}

func (s *APIServer) handleIndex(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.HTML(http.StatusOK, "index.html", nil)
}

// handleLogin authenticates a user and generates a JWT token.
// swagger:route POST /login auth handleLogin
//
// Authenticate user and generate JWT token.
//
// Responses:
//   200: tokenResponse
//   400: errorResponse
//   401: errorResponse
//   500: errorResponse
func (s *APIServer) handleLogin(ctx *gin.Context) {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&loginData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login data"})
		return
	}

	user, err := s.storage.GetUserByUsernameAndPassword(loginData.Username, loginData.Password)
	if err != nil || user == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// handleRegister registers a new user and generates a JWT token.
// swagger:route POST /register auth handleRegister
//
// Register a new user and generate JWT token.
//
// Responses:
//   200: tokenResponse
//   400: errorResponse
//   409: errorResponse
//   500: errorResponse
func (s *APIServer) handleRegister(ctx *gin.Context) {
	var registrationData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&registrationData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid registration data"})
		return
	}

	existingUser, err := s.storage.GetUserByUsername(registrationData.Username)
	if err == nil || existingUser != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		s.logger.Info(err)
		return
	}

	userID, err := s.storage.CreateUser(registrationData.Username, registrationData.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		s.logger.Info(err)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// handleGetTasks retrieves tasks for the authenticated user.
// swagger:route GET /tasks tasks handleGetTasks
//
// Get tasks for the authenticated user.
//
// Responses:
//   200: tasksResponse
//   400: errorResponse
//   401: errorResponse
//   500: errorResponse
func (s *APIServer) handleGetTasks(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")

	tasks, err := s.storage.GetTasks(int(userID.(float64)))
	if err != nil {
		s.logger.Error("Error fetching tasks: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	ctx.JSON(http.StatusOK, tasks)
}


// handleCreateTask retrieves a specific task for the authenticated user.
// swagger:route post /tasks/{taskID} tasks handleCreateTask
//
// Create a specific task for the authenticated user.
//
// Responses:
//   200: Task response
//   400: errorResponse
//   401: errorResponse
//   404: errorResponse
//   500: errorResponse
func (s *APIServer) handleCreateTask(ctx *gin.Context) {
	var task models.Task
	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task data"})
		return
	}

	userID, ok := ctx.Get("userID")

	if !ok {
		s.logger.Error("No userID")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Wrong user id"})
	}
	err := s.storage.CreateTask(int(userID.(float64)), task.Title, task.Description)
	if err != nil {
		s.logger.Error("Error creating task: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	ctx.JSON(http.StatusOK, task)
}

// handleGetTask retrieves a specific task for the authenticated user.
// swagger:route GET /tasks/{taskID} tasks handleGetTask
//
// Get a specific task for the authenticated user.
//
// Responses:
//   200: taskResponse
//   400: errorResponse
//   401: errorResponse
//   404: errorResponse
//   500: errorResponse
func (s *APIServer) handleGetTask(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("userID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	taskID, err := strconv.Atoi(ctx.Param("taskID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := s.storage.GetTaskByID(userID, taskID)
	if err != nil {
		s.logger.Error("Error fetching task: ", err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	ctx.JSON(http.StatusOK, task)
}

// handleUpdateTask updates a specific task for the authenticated user.
// swagger:route PUT /tasks/{taskID} tasks handleUpdateTask
//
// Update a specific task for the authenticated user.
//
// Responses:
//   200: messageResponse
//   400: errorResponse
//   401: errorResponse
//   404: errorResponse
//   500: errorResponse
func (s *APIServer) handleUpdateTask(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("userID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	taskID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var task models.Task
	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task data"})
		return
	}

	task.ID = taskID

	if err := s.storage.UpdateTask(userID, &task); err != nil {
		s.logger.Error("Error updating task: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Task updated successfully"})
}

// handleDeleteTask deletes a specific task for the authenticated user.
// swagger:route DELETE /tasks/{id} tasks handleDeleteTask
//
// Delete a specific task for the authenticated user.
//
// Responses:
//   200: messageResponse
//   400: errorResponse
//   401: errorResponse
//   404: errorResponse
//   500: errorResponse
func (s *APIServer) handleDeleteTask(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("userID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	taskID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := s.storage.DeleteTask(userID, taskID); err != nil {
		s.logger.Error("Error deleting task: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}