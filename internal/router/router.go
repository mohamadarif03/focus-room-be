package router

import (
	"log"
	"os"

	"github.com/mohamadarif03/focus-room-be/internal/database"
	"github.com/mohamadarif03/focus-room-be/internal/handler"
	"github.com/mohamadarif03/focus-room-be/internal/middleware"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
	"github.com/mohamadarif03/focus-room-be/internal/service"
	"github.com/mohamadarif03/focus-room-be/pkg/utils"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	db := database.DB

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	youtubeAPIKey := os.Getenv("YOUTUBE_API_KEY")

	if err := utils.InitYouTubeService(youtubeAPIKey); err != nil {
		log.Fatalf("Gagal inisialisasi YouTube Service: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	authService := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authService)

	taskRepo := repository.NewTaskRepository(db)
	taskService := service.NewTaskService(taskRepo)
	taskHandler := handler.NewTaskHandler(taskService)

	matRepo := repository.NewMaterialRepository(db)

	aiService, err := service.NewAIService(geminiAPIKey, matRepo) // <-- Perlu matRepo
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
			taskGroup := studentGroup.Group("/tasks")
			{
				taskGroup.POST("/", taskHandler.CreateTask)
				taskGroup.GET("/", taskHandler.GetTasks)
				taskGroup.PUT("/:id", taskHandler.UpdateTask)
				taskGroup.DELETE("/:id", taskHandler.DeleteTask)
			}

			materialGroup := studentGroup.Group("/materials")
			{
				materialGroup.POST("/pdf", aiHandler.IngestPDF)
				materialGroup.POST("/youtube", aiHandler.IngestYouTube)
			}

			aiGroup := studentGroup.Group("/ai")
			{
				aiGroup.POST("/summarize", aiHandler.GenerateSummary)
				aiGroup.POST("/quiz", aiHandler.GenerateQuiz)
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
