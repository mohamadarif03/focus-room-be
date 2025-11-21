package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mohamadarif03/focus-room-be/internal/handler"
	"github.com/mohamadarif03/focus-room-be/internal/middleware"
	"github.com/mohamadarif03/focus-room-be/internal/service"
)

func SetupRouter(
	userService *service.UserService,
	authService *service.AuthService,
	taskService *service.TaskService,
	aiService *service.AIService,
	packageService *service.PackageService,
) *gin.Engine {

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService)
	taskHandler := handler.NewTaskHandler(taskService)
	aiHandler := handler.NewAIHandler(aiService)
	packageHandler := handler.NewPackageHandler(packageService)

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
			packageGroup := studentGroup.Group("/packages")
			{
				packageGroup.POST("", packageHandler.CreatePackage)
				packageGroup.GET("", packageHandler.GetMyPackages)
				packageGroup.GET("/:id", packageHandler.GetPackageByID)
				packageGroup.PUT("/:id", packageHandler.UpdatePackage)
				packageGroup.DELETE("/:id", packageHandler.DeletePackage)
			}
			taskGroup := studentGroup.Group("/tasks")
			{
				taskGroup.POST("/", taskHandler.CreateTask)
				taskGroup.GET("/", taskHandler.GetTasks)
				taskGroup.PUT("/:id", taskHandler.UpdateTask)
				taskGroup.DELETE("/:id", taskHandler.DeleteTask)
			}

			materialGroup := studentGroup.Group("/materials")
			{
				materialGroup.GET("", aiHandler.GetMaterials)
				materialGroup.POST("/pdf", aiHandler.IngestPDF)
				materialGroup.POST("/youtube", aiHandler.IngestYouTube)
			}

			aiGroup := studentGroup.Group("/ai")
			{
				aiGroup.POST("/summarize", aiHandler.GenerateSummary)
				aiGroup.POST("/quiz", aiHandler.GenerateQuiz)
				aiGroup.POST("/flashcards", aiHandler.GenerateFlashcards)
			}

			dailyQuizGroup := studentGroup.Group("/daily-quiz")
			{
				dailyQuizGroup.GET("", aiHandler.GetDailyQuiz)
				dailyQuizGroup.POST("/claim", aiHandler.ClaimDailyStreak)
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
