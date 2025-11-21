package main

import (
	// "log"
	// "os" // Aktifkan jika pakai env

	"github.com/joho/godotenv"
	"github.com/mohamadarif03/focus-room-be/internal/database"
	// "github.com/mohamadarif03/focus-room-be/internal/handler" // Import Handler
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/repository"
	"github.com/mohamadarif03/focus-room-be/internal/router"
	"github.com/mohamadarif03/focus-room-be/internal/service"
)

func main() {
	godotenv.Load()

	database.InitDB()
	database.DB.AutoMigrate(&model.User{}, &model.Task{}, &model.Material{}, &model.Package{},)
	// database.Seed() // Jalankan sekali saja

	geminiAPIKey := "AIzaSyDHEQWpBthrtuBhgUnVW3MkIvwfTPmBnQ8"
	// HAPUS YOUTUBE API KEY & INIT DI SINI

	db := database.DB

	// Repo
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	matRepo := repository.NewMaterialRepository(db)
	pkgRepo := repository.NewPackageRepository(db)
	// statsRepo := repository.NewStatsRepository(db)

	// Service
	userService := service.NewUserService(userRepo, taskRepo)
	authService := service.NewAuthService(userRepo)
	taskService := service.NewTaskService(taskRepo, userRepo)
	pkgService := service.NewPackageService(pkgRepo)
	// statsService := service.NewStatsService(statsRepo)
	aiService, _ := service.NewAIService(geminiAPIKey, matRepo, pkgRepo, userRepo)

	// Handler
	// statsHandler := handler.NewStatsHandler(statsService)
	// (Handler lain dibuat di dalam router jika Anda belum refactor router)

	r := router.SetupRouter(
		userService,
		authService,
		taskService,
		aiService,
		pkgService,
		// statsHandler, // Pastikan router.go menerima parameter ini
	)

	r.Run(":8080")
}