package router

import (
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

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)
	authService := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authService)


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