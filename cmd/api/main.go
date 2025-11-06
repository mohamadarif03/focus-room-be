package main

import (
	"github.com/mohamadarif03/focus-room-be/internal/config"
	"github.com/mohamadarif03/focus-room-be/internal/database"
	"github.com/mohamadarif03/focus-room-be/internal/model"
	"github.com/mohamadarif03/focus-room-be/internal/router"
	"log"
)

func main() {
	config.LoadConfig()

	database.InitDB()

	err := database.DB.AutoMigrate(&model.User{})
	if err != nil {
		log.Fatal("Failed to auto-migrate database:", err)
	}
	
	log.Println("Database migration completed")


	r := router.SetupRouter()

	log.Println("Starting server on port :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}