package router

import (
	"github.com/mohamadarif03/focus-room-be/internal/database"
	"github.com/mohamadarif03/focus-room-be/internal/handler"
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

	api := r.Group("/api/v1")
	{
		userRoutes := api.Group("/users")
		{
			userRoutes.POST("/", userHandler.CreateUser)
			userRoutes.GET("/", userHandler.GetUsers)
		}
	}

	return r
}
