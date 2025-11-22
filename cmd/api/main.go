package main

import (
	// "log"
	// "os" // Aktifkan jika pakai env

	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mohamadarif03/focus-room-be/internal/database"

	// "github.com/mohamadarif03/focus-room-be/internal/handler" // Import Handler
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
	"github.com/mohamadarif03/focus-room-be/internal/router"
	"github.com/mohamadarif03/focus-room-be/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Info: File .env tidak ditemukan. Menggunakan System Environment Variables (Railway/Docker).")
	}
	database.InitDB()
	database.DB.AutoMigrate(&model.User{}, &model.Task{}, &model.Material{}, &model.Package{})

	fmt.Println("--- LIST ENV VARIABLES TERDETEKSI ---")
	for _, e := range os.Environ() {
		// Kita split biar cuma print NAMA variable-nya saja (KEY), jangan print VALUE-nya
		pair := strings.SplitN(e, "=", 2)
		fmt.Println(pair[0])
	}
	fmt.Println("-------------------------------------")
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")

	db := database.DB

	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	matRepo := repository.NewMaterialRepository(db)
	pkgRepo := repository.NewPackageRepository(db)
	statsRepo := repository.NewStatsRepository(db)

	// Service
	userService := service.NewUserService(userRepo, taskRepo)
	authService := service.NewAuthService(userRepo)
	taskService := service.NewTaskService(taskRepo, userRepo)
	pkgService := service.NewPackageService(pkgRepo)
	statsService := service.NewStatsService(statsRepo)
	aiService, _ := service.NewAIService(geminiAPIKey, matRepo, pkgRepo, userRepo, statsRepo)

	r := router.SetupRouter(
		userService,
		authService,
		taskService,
		aiService,
		pkgService,
		statsService,
	)

	r.Run(":8080")
}
