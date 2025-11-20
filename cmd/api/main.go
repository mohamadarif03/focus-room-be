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
		log.Print("Gagal memuat file .env")
	}

	database.InitDB()
	log.Println("Melakukan AutoMigrate...")
	database.DB.AutoMigrate(&model.User{}, &model.Task{}, &model.Material{}, &model.Package{})
	database.Seed()

	geminiAPIKey := "AIzaSyABmC0orVoBpegnBj6e4f9XL5_kdyYn2vU"
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

	aiService, err := service.NewAIService(geminiAPIKey, matRepo, packageRepo, userRepo)
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
