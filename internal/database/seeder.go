package database

import (
	"log"

	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/pkg/utils"
)

func Seed() {
	log.Println("Menjalankan Seeder...")

	password, _ := utils.HashPassword("password")

	adminUser := model.User{
		Username:     "Admin",
		Email:        "admin@example.com",
		PasswordHash: password,
		Role:         "admin",
	}

	result := DB.Where(model.User{Email: adminUser.Email}).FirstOrCreate(&adminUser)

	if result.Error != nil {
		log.Printf("Gagal menjalankan seeder admin: %v", result.Error)
		return
	}

	if result.RowsAffected > 0 {
		log.Printf("Admin berhasil DIBUAT: %s", adminUser.Email)
	} else {
		log.Printf("User admin SUDAH ADA: %s", adminUser.Email)
	}

	log.Println("Seeder selesai dijalankan.")
}
