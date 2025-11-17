package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/mohamadarif03/focus-room-be/internal/database"
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
	"github.com/mohamadarif03/focus-room-be/internal/router"
	"github.com/mohamadarif03/focus-room-be/internal/service"
	"github.com/mohamadarif03/focus-room-be/pkg/utils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Gagal memuat file .env")
	}

	database.InitDB()
	log.Println("Melakukan AutoMigrate untuk User, Task, Package dan Material...")
	database.DB.AutoMigrate(&model.User{}, &model.Task{}, &model.Material{}, &model.Package{})
	database.Seed()

	geminiAPIKey := "AIzaSyDHEQWpBthrtuBhgUnVW3MkIvwfTPmBnQ8"
	youtubeAPIKey := "AIzaSyDHEQWpBthrtuBhgUnVW3MkIvwfTPmBnQ8"

	if err := utils.InitYouTubeService(youtubeAPIKey); err != nil {
		log.Fatalf("Gagal inisialisasi YouTube Service: %v", err)
	}

	db := database.DB

	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	matRepo := repository.NewMaterialRepository(db)
	packageRepo := repository.NewPackageRepository(db)

	userService := service.NewUserService(userRepo, taskRepo)

	authService := service.NewAuthService(userRepo)

	taskService := service.NewTaskService(taskRepo, userRepo)
	packageService := service.NewPackageService(packageRepo)
	aiService, err := service.NewAIService(geminiAPIKey, matRepo)
	if err != nil {
		log.Fatalf("Gagal inisialisasi AI Service: %v", err)
	}

	r := router.SetupRouter(
		userService,
		authService,
		taskService,
		aiService,
		packageService,
	)

	log.Println("Server berjalan di port 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Gagal menjalankan server:", err)
	}
}
