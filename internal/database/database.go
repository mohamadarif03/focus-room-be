package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var dsn string

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		dsn = "postgresql://postgres:hByTXiRNwcGYKZvstdqKywOzonClARId@trolley.proxy.rlwy.net:25944/railway"
	} else {
		log.Println("DATABASE_URL tidak ditemukan, merakit DSN dari .env (mode Lokal)")
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		sslmode := os.Getenv("DB_SSLMODE")

		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
			host, user, password, dbname, port, sslmode)
	}
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connection successfully opened")
}
