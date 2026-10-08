package main

import (
	"fmt"
	"siakad-mini/config"
	"siakad-mini/routes"
	"siakad-mini/seeders"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Gagal load .env: " + err.Error())
	}
	config.ConnectDB()
	seeders.RunSeeder()
	fmt.Println("🚀 Server SIAKAD Mini jalan di http://localhost:8080")
	r := routes.SetupRouter()
	r.Run(":8080")
}
