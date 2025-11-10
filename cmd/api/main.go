package main

import (
	"log"

	"github.com/mohamadarif03/focus-room-be/internal/config"
	"github.com/mohamadarif03/focus-room-be/internal/database"
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/router"
)

func main() {
	config.LoadConfig()

	database.InitDB()

	database.DB.AutoMigrate(&model.User{})
	database.DB.AutoMigrate(&model.Task{})

	log.Println("Database migration completed")

	r := router.SetupRouter()
	database.Seed()

	log.Println("Starting server on port :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
