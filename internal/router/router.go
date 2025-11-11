package router

import (
	"log"
	"os"

	"github.com/mohamadarif03/focus-room-be/internal/database"
	"github.com/mohamadarif03/focus-room-be/internal/handler"
	"github.com/mohamadarif03/focus-room-be/internal/middleware"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
	"github.com/mohamadarif03/focus-room-be/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	db := database.DB

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)
	authService := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authService)

	taskRepo := repository.NewTaskRepository(db)
	taskService := service.NewTaskService(taskRepo)
	taskHandler := handler.NewTaskHandler(taskService)

	aiService, err := service.NewAIService(geminiAPIKey)
	if err != nil {
		log.Fatalf("Gagal inisialisasi AI Service: %v", err)
	}
	aiHandler := handler.NewAIHandler(aiService)

	api := r.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		authedGroup := api.Group("/")
		authedGroup.Use(middleware.AuthMiddleware())
		{
			authedGroup.GET("/users/me", userHandler.GetSelf)

		}

		studentGroup := api.Group("/student")
		studentGroup.Use(middleware.AuthMiddleware())
		studentGroup.Use(middleware.StudentMiddleware())
		{
			studentGroup.POST("/tasks", taskHandler.CreateTask)
			studentGroup.GET("/tasks", taskHandler.GetTasks)
			studentGroup.PUT("tasks/:id", taskHandler.UpdateTask)
			studentGroup.DELETE("tasks/:id", taskHandler.DeleteTask)

			aiGroup := studentGroup.Group("/ai")
			{
				aiGroup.POST("/summarize/pdf", aiHandler.SummarizePDF)
			}
		}

		adminGroup := api.Group("/admin")
		adminGroup.Use(middleware.AuthMiddleware())
		adminGroup.Use(middleware.AdminMiddleware())
		{
			adminGroup.GET("/users", userHandler.GetUsers)

			adminGroup.GET("/users/:id", userHandler.GetUserByID)

			adminGroup.PUT("/users/:id", userHandler.UpdateUser)

			adminGroup.DELETE("/users/:id", userHandler.DeleteUser)
		}
	}

	return r
}
