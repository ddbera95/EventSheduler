package data

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

var DB *gorm.DB

func Init() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	err = db.AutoMigrate(&Trigger{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	err = db.AutoMigrate(&Event{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_email ON users(email);")

	DB = db
}
