package main

import (
	"log"
	"os"
	"taskflow/internal/handlers"
	"taskflow/internal/middleware"
	"taskflow/internal/repository"
	"taskflow/internal/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	port := os.Getenv("PORT")
	frontendURL := os.Getenv("FRONTEND_URL")

	if dbURL == "" {
		dbURL = "postgres://taskflow:taskflow_pass@localhost:5434/taskflow?sslmode=disable"
	}
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	if jwtSecret == "" {
		jwtSecret = "dev_secret"
	}
	if port == "" {
		port = "8080"
	}
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	// Database
	pool, err := repository.NewPool(dbURL)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer pool.Close()

	if err := repository.Migrate(pool); err != nil {
		log.Fatal("failed to run migrations:", err)
	}
	if err := repository.MigratePasswordResets(pool); err != nil {
		log.Fatal("failed to run password reset migration:", err)
	}
	log.Println("✅ Database ready")

	// Repos
	userRepo := repository.NewUserRepo(pool)
	boardRepo := repository.NewBoardRepo(pool)
	columnRepo := repository.NewColumnRepo(pool)
	taskRepo := repository.NewTaskRepo(pool)
	activityRepo := repository.NewActivityRepo(pool)

	// WebSocket hub
	hub := websocket.NewHub()
	log.Println("✅ WebSocket hub initialized")

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo, jwtSecret)
	boardHandler := handlers.NewBoardHandler(boardRepo, columnRepo, taskRepo, activityRepo, userRepo, hub, frontendURL)

	// Router
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Routes
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthMiddleware(jwtSecret), authHandler.Me)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		boards := api.Group("/boards", middleware.AuthMiddleware(jwtSecret))
		{
			boards.GET("", boardHandler.ListBoards)
			boards.POST("", boardHandler.CreateBoard)
			boards.GET("/:boardId", boardHandler.GetBoard)
			boards.POST("/:boardId/columns", boardHandler.CreateColumn)
			boards.DELETE("/:boardId/columns/:columnId", boardHandler.DeleteColumn)
			boards.POST("/:boardId/tasks", boardHandler.CreateTask)
			boards.PUT("/:boardId/tasks/:taskId", boardHandler.UpdateTask)
			boards.PATCH("/:boardId/tasks/:taskId/move", boardHandler.MoveTask)
			boards.DELETE("/:boardId/tasks/:taskId", boardHandler.DeleteTask)
			boards.GET("/:boardId/activity", boardHandler.GetActivity)
			boards.POST("/:boardId/invite", boardHandler.InviteMember)
		}
	}

	// WebSocket (auth via query param for WS handshake)
	r.GET("/ws/:boardId", middleware.AuthMiddleware(jwtSecret), boardHandler.HandleWebSocket)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Printf("✅ Server running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
