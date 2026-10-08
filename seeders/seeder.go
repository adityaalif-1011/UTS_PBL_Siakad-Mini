package seeders

import (
	"fmt"
	"math/rand"

	"golang.org/x/crypto/bcrypt"

	"siakad-mini/config"
	"siakad-mini/models"
)

func RunSeeder() {
	var count int64
	config.DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		fmt.Println("Seeder dilewati, data sudah ada.")
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte("admin12345"), bcrypt.DefaultCost)
	admin := models.User{Email: "admin@siakad.ac.id", Password: string(hashed), Role: "admin"}
	config.DB.Create(&admin)

	prodis := []string{"Sistem Informasi", "Teknik Informatika", "Ilmu Komputer"}
	for i := 1; i <= 20; i++ {
		nim := fmt.Sprintf("18722100%04d", i)
		email := fmt.Sprintf("mahasiswa%d@siakad.ac.id", i)
		pw, _ := bcrypt.GenerateFromPassword([]byte(nim), bcrypt.DefaultCost)
		user := models.User{Email: email, Password: string(pw), Role: "mahasiswa"}
		config.DB.Create(&user)
		student := models.Student{
			UserID:      user.ID,
			NIM:         nim,
			Nama:        fmt.Sprintf("Mahasiswa %d", i),
			Prodi:       prodis[i%len(prodis)],
			Angkatan:    2022 + (i % 3),
			IPKTerakhir: 2.0 + rand.Float64()*2.0,
		}
		config.DB.Create(&student)
	}

	for i := 1; i <= 10; i++ {
		course := models.Course{
			KodeMK:   fmt.Sprintf("MK%03d", i),
			NamaMK:   fmt.Sprintf("Mata Kuliah %d", i),
			SKS:      2 + (i % 3),
			Semester: 1 + (i % 8),
			Kuota:    30,
		}
		config.DB.Create(&course)
	}

	fmt.Println("Seeder selesai: 1 admin, 20 mahasiswa, 10 mata kuliah")
}
