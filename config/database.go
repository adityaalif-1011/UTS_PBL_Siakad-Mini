package config

import (
	"fmt"
	"os"

	"siakad-mini/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Gagal koneksi database: " + err.Error())
	}

	err = db.AutoMigrate(&models.User{}, &models.Student{}, &models.Course{}, &models.Enrollment{})
	if err != nil {
		panic("Gagal migrate: " + err.Error())
	}

	DB = db

	fmt.Println("✅ Database terkoneksi & migration selesai")
}
