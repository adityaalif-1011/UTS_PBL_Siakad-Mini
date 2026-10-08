package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"siakad-mini/config"
	"siakad-mini/helpers"
	"siakad-mini/models"
)

func hitungBatasSKS(ipk float64) int {
	if ipk >= 3.00 {
		return 24
	}
	if ipk >= 2.50 {
		return 21
	}
	return 18
}

func GetStudents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if perPage > 50 {
		perPage = 50
	}
	if page < 1 {
		page = 1
	}
	prodi := c.Query("prodi")
	angkatan := c.Query("angkatan")
	search := c.Query("search")
	sort := c.DefaultQuery("sort", "nama")

	q := config.DB.Model(&models.Student{})
	if prodi != "" {
		q = q.Where("prodi = ?", prodi)
	}
	if angkatan != "" {
		q = q.Where("angkatan = ?", angkatan)
	}
	if search != "" {
		q = q.Where("nim ILIKE ? OR nama ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	switch sort {
	case "-ipk_terakhir":
		q = q.Order("ipk_terakhir DESC")
	default:
		q = q.Order("nama ASC")
	}

	var total int64
	q.Session(&gorm.Session{}).Count(&total)

	var students []models.Student
	offset := (page - 1) * perPage
	q.Offset(offset).Limit(perPage).Find(&students)

	lastPage := int((total + int64(perPage) - 1) / int64(perPage))

	helpers.Success(c, 200, "Data mahasiswa berhasil diambil", students, gin.H{
		"current_page": page,
		"per_page":     perPage,
		"total":        total,
		"last_page":    lastPage,
	})
}

type CreateStudentInput struct {
	NIM         string  `json:"nim" binding:"required,len=12,numeric"`
	Nama        string  `json:"nama" binding:"required"`
	Email       string  `json:"email" binding:"required,email"`
	Prodi       string  `json:"prodi" binding:"required"`
	Angkatan    int     `json:"angkatan" binding:"required,min=1000,max=9999"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

func CreateStudent(c *gin.Context) {
	var input CreateStudentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.ValidationError(c, mapBindErrors(err))
		return
	}

	if input.Angkatan > time.Now().Year() {
		helpers.ValidationError(c, map[string][]string{"angkatan": {"Tidak boleh melebihi tahun berjalan"}})
		return
	}

	if input.IPKTerakhir < 0 || input.IPKTerakhir > 4 {
		helpers.ValidationError(c, map[string][]string{"ipk_terakhir": {"Harus antara 0.00 - 4.00"}})
		return
	}

	var count int64
	config.DB.Model(&models.Student{}).Where("nim = ?", input.NIM).Count(&count)
	if count > 0 {
		helpers.ValidationError(c, map[string][]string{"nim": {"NIM sudah terdaftar"}})
		return
	}

	config.DB.Model(&models.User{}).Where("email = ?", input.Email).Count(&count)
	if count > 0 {
		helpers.ValidationError(c, map[string][]string{"email": {"Email sudah terdaftar"}})
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.NIM), bcrypt.DefaultCost)

	var student models.Student
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		user := models.User{
			Email:    input.Email,
			Password: string(hashed),
			Role:     "mahasiswa",
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		student = models.Student{
			UserID:      user.ID,
			NIM:         input.NIM,
			Nama:        input.Nama,
			Prodi:       input.Prodi,
			Angkatan:    input.Angkatan,
			IPKTerakhir: input.IPKTerakhir,
		}
		return tx.Create(&student).Error
	})
	if err != nil {
		helpers.Error(c, 500, "Gagal menambah mahasiswa", nil)
		return
	}

	helpers.Success(c, http.StatusCreated, "Mahasiswa berhasil ditambahkan", student, nil)
}

func GetStudentByID(c *gin.Context) {
	id := c.Param("id")
	role := c.GetString("role")
	userID := c.GetUint("user_id")

	var student models.Student
	if err := config.DB.Preload("Enrollments.Course").First(&student, id).Error; err != nil {
		helpers.Error(c, 404, "Mahasiswa tidak ditemukan", nil)
		return
	}

	if role == "mahasiswa" && student.UserID != userID {
		helpers.Error(c, 403, "Akses ditolak", nil)
		return
	}

	enrollments := []gin.H{}
	totalSKS := 0
	for _, e := range student.Enrollments {
		enrollments = append(enrollments, gin.H{
			"id":             e.ID,
			"course_id":      e.CourseID,
			"kode_mk":        e.Course.KodeMK,
			"nama_mk":        e.Course.NamaMK,
			"sks":            e.Course.SKS,
			"tahun_akademik": e.TahunAkademik,
		})
		totalSKS += e.Course.SKS
	}

	batasSKS := hitungBatasSKS(student.IPKTerakhir)

	helpers.Success(c, 200, "Detail mahasiswa berhasil diambil", gin.H{
		"id":           student.ID,
		"nim":          student.NIM,
		"nama":         student.Nama,
		"prodi":        student.Prodi,
		"angkatan":     student.Angkatan,
		"ipk_terakhir": student.IPKTerakhir,
		"courses":      enrollments,
		"total_sks":    totalSKS,
		"batas_sks":    batasSKS,
	}, nil)
}

type UpdateStudentInput struct {
	Nama        string  `json:"nama"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan" binding:"omitempty,min=1000,max=9999"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

func UpdateStudent(c *gin.Context) {
	var student models.Student
	if err := config.DB.First(&student, c.Param("id")).Error; err != nil {
		helpers.Error(c, 404, "Mahasiswa tidak ditemukan", nil)
		return
	}

	var input UpdateStudentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.ValidationError(c, mapBindErrors(err))
		return
	}

	if input.Angkatan != 0 && input.Angkatan > time.Now().Year() {
		helpers.ValidationError(c, map[string][]string{"angkatan": {"Tidak boleh melebihi tahun berjalan"}})
		return
	}

	if input.IPKTerakhir < 0 || input.IPKTerakhir > 4 {
		helpers.ValidationError(c, map[string][]string{"ipk_terakhir": {"Harus antara 0.00 - 4.00"}})
		return
	}

	if input.Nama != "" {
		student.Nama = input.Nama
	}
	if input.Prodi != "" {
		student.Prodi = input.Prodi
	}
	if input.Angkatan != 0 {
		student.Angkatan = input.Angkatan
	}
	if input.IPKTerakhir >= 0 && input.IPKTerakhir <= 4 {
		student.IPKTerakhir = input.IPKTerakhir
	}

	config.DB.Save(&student)

	helpers.Success(c, 200, "Data mahasiswa berhasil diperbarui", student, nil)
}

func DeleteStudent(c *gin.Context) {
	var student models.Student
	if err := config.DB.First(&student, c.Param("id")).Error; err != nil {
		helpers.Error(c, 404, "Mahasiswa tidak ditemukan", nil)
		return
	}

	config.DB.Delete(&student)

	c.JSON(http.StatusNoContent, nil)
}
