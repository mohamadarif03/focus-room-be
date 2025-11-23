package main

import (
	// "log"
	// "os" // Aktifkan jika pakai env

	"log"
	"os"

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
	database.DB.AutoMigrate(&model.User{}, &model.Task{}, &model.Material{}, &model.Package{}, &model.Quiz{},
		&model.QuizQuestion{}, &model.QuizAttempt{},
		&model.QuizAttemptDetail{})

	db := database.DB

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	query := `
			TRUNCATE TABLE 
				users, 
				materials, 
				packages, 
				quizzes, 
				quiz_questions, 
				quiz_attempts, 
				tasks 
			RESTART IDENTITY CASCADE;
		`

	if err := db.Exec(query).Error; err != nil {
		log.Fatalf("Gagal mereset database: %v", err)
	}

	log.Println("Sukses! Semua data telah dihapus bersih.")
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	matRepo := repository.NewMaterialRepository(db)
	pkgRepo := repository.NewPackageRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	quizRepo := repository.NewQuizRepository(db)

	userService := service.NewUserService(userRepo, taskRepo)
	authService := service.NewAuthService(userRepo)
	quizService := service.NewQuizService(quizRepo, matRepo)
	taskService := service.NewTaskService(taskRepo, userRepo)
	pkgService := service.NewPackageService(pkgRepo)
	statsService := service.NewStatsService(statsRepo)
	aiService, _ := service.NewAIService(geminiAPIKey, matRepo, pkgRepo, userRepo, statsRepo, quizRepo)

	r := router.SetupRouter(
		userService,
		authService,
		taskService,
		aiService,
		pkgService,
		statsService,
		quizService,
	)

	r.Run(":8080")
}
