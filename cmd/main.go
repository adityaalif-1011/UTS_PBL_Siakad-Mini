package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"siakad-mini/config"
	"siakad-mini/routes"
	"siakad-mini/seeders"
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
