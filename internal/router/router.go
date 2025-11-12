package router

import (
	"github.com/mohamadarif03/focus-room-be/internal/handler"
	"github.com/mohamadarif03/focus-room-be/internal/middleware"
	"github.com/mohamadarif03/focus-room-be/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	userService *service.UserService,
	authService *service.AuthService,
	taskService *service.TaskService,
	aiService *service.AIService,
) *gin.Engine {

	r := gin.Default()

	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService)
	taskHandler := handler.NewTaskHandler(taskService)
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

			streakGroup := studentGroup.Group("/streaks")
			{
				streakGroup.POST("/check", userHandler.CheckAndUpdateStreak)
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
