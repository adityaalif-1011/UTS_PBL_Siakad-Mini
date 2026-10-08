package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"siakad-mini/config"
	"siakad-mini/helpers"
	"siakad-mini/models"
)

type EnrollmentInput struct {
	CourseID      uint   `json:"course_id" binding:"required"`
	TahunAkademik string `json:"tahun_akademik" binding:"required"`
}

var tahunAkademikRegex = regexp.MustCompile(`^\d{4}/\d{4}-(Ganjil|Genap)$`)

func CreateEnrollment(c *gin.Context) {
	userID := c.GetUint("user_id")

	var input EnrollmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.Error(c, 422, "Validasi gagal", gin.H{"error": err.Error()})
		return
	}

	if !tahunAkademikRegex.MatchString(input.TahunAkademik) {
		helpers.Error(c, 422, "Validasi gagal", gin.H{"tahun_akademik": "Format harus 2026/2027-Ganjil"})
		return
	}

	var student models.Student
	if err := config.DB.Where("user_id = ?", userID).First(&student).Error; err != nil {
		helpers.Error(c, 403, "Hanya mahasiswa yang dapat mengambil mata kuliah", nil)
		return
	}

	var enrollment models.Enrollment
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var course models.Course
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&course, input.CourseID).Error; err != nil {
			return errors.New("course_not_found")
		}

		var dup int64
		tx.Model(&models.Enrollment{}).
			Where("student_id = ? AND course_id = ? AND tahun_akademik = ?", student.ID, input.CourseID, input.TahunAkademik).
			Count(&dup)
		if dup > 0 {
			return errors.New("duplicate")
		}

		var terisi int64
		tx.Model(&models.Enrollment{}).Where("course_id = ?", input.CourseID).Count(&terisi)
		if int(terisi) >= course.Kuota {
			return errors.New("quota_full")
		}

		var existing []models.Enrollment
		tx.Preload("Course").Where("student_id = ? AND tahun_akademik = ?", student.ID, input.TahunAkademik).Find(&existing)
		totalSKS := 0
		for _, e := range existing {
			totalSKS += e.Course.SKS
		}
		batasSKS := hitungBatasSKS(student.IPKTerakhir)
		if totalSKS+course.SKS > batasSKS {
			return fmt.Errorf("sks_exceeded:%d", batasSKS-totalSKS)
		}

		enrollment = models.Enrollment{
			StudentID:     student.ID,
			CourseID:      input.CourseID,
			TahunAkademik: input.TahunAkademik,
		}
		return tx.Create(&enrollment).Error
	})
	if err != nil {
		msg := err.Error()
		switch {
		case msg == "duplicate":
			helpers.Error(c, 409, "Mata kuliah sudah pernah diambil pada tahun akademik ini", nil)
		case msg == "quota_full":
			helpers.Error(c, 422, "Kuota mata kuliah penuh", nil)
		case msg == "course_not_found":
			helpers.Error(c, 404, "Mata kuliah tidak ditemukan", nil)
		case len(msg) >= 13 && msg[:13] == "sks_exceeded:":
			helpers.Error(c, 422, fmt.Sprintf("Total SKS melebihi batas. Sisa SKS tersedia: %s", msg[13:]), nil)
		default:
			helpers.Error(c, 500, "Gagal mengambil mata kuliah", nil)
		}
		return
	}

	helpers.Success(c, http.StatusCreated, "Mata kuliah berhasil diambil", enrollment, nil)
}

func DeleteEnrollment(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetUint("user_id")

	var student models.Student
	config.DB.Where("user_id = ?", userID).First(&student)

	var enrollment models.Enrollment
	if err := config.DB.First(&enrollment, id).Error; err != nil {
		helpers.Error(c, 404, "Enrollment tidak ditemukan", nil)
		return
	}

	if enrollment.StudentID != student.ID {
		helpers.Error(c, 403, "Akses ditolak", nil)
		return
	}

	config.DB.Delete(&enrollment)

	c.JSON(http.StatusNoContent, nil)
}
