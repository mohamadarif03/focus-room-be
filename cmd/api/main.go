package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mohamadarif03/focus-room-be/internal/database"
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
	"github.com/mohamadarif03/focus-room-be/internal/router"
	"github.com/mohamadarif03/focus-room-be/internal/service"
	"github.com/mohamadarif03/focus-room-be/pkg/utils"
)

func main() {
	_, isRunningOnRailway := os.LookupEnv("RAILWAY_ENVIRONMENT")

	if !isRunningOnRailway {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Gagal memuat file .env")
		}
	}

	database.InitDB()
	log.Println("Melakukan AutoMigrate untuk User, Task, dan Material...")
	database.DB.AutoMigrate(&model.User{}, &model.Task{}, &model.Material{})
	database.Seed()

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	youtubeAPIKey := os.Getenv("YOUTUBE_API_KEY")

	if err := utils.InitYouTubeService(youtubeAPIKey); err != nil {
		log.Fatalf("Gagal inisialisasi YouTube Service: %v", err)
	}

	db := database.DB

	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	matRepo := repository.NewMaterialRepository(db)

	userService := service.NewUserService(userRepo, taskRepo)

	authService := service.NewAuthService(userRepo)

	taskService := service.NewTaskService(taskRepo, userRepo)

	aiService, err := service.NewAIService(geminiAPIKey, matRepo)
	if err != nil {
		log.Fatalf("Gagal inisialisasi AI Service: %v", err)
	}

	r := router.SetupRouter(
		userService,
		authService,
		taskService,
		aiService,
	)

	log.Println("Server berjalan di port 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Gagal menjalankan server:", err)
	}
}
