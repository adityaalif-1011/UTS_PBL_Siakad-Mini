package controllers

import (
	"github.com/gin-gonic/gin"

	"siakad-mini/config"
	"siakad-mini/helpers"
	"siakad-mini/models"
)

func GetCourses(c *gin.Context) {
	semester := c.Query("semester")
	search := c.Query("search")
	available := c.Query("available")

	var courses []models.Course
	q := config.DB.Model(&models.Course{})
	if semester != "" {
		q = q.Where("semester = ?", semester)
	}
	if search != "" {
		q = q.Where("kode_mk ILIKE ? OR nama_mk ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	q.Find(&courses)

	result := []gin.H{}
	for _, course := range courses {
		var terisi int64
		config.DB.Model(&models.Enrollment{}).Where("course_id = ?", course.ID).Count(&terisi)
		sisa := course.Kuota - int(terisi)
		if available == "true" && sisa <= 0 {
			continue
		}
		result = append(result, gin.H{
			"id":         course.ID,
			"kode_mk":    course.KodeMK,
			"nama_mk":    course.NamaMK,
			"sks":        course.SKS,
			"semester":   course.Semester,
			"kuota":      course.Kuota,
			"terisi":     terisi,
			"sisa_kuota": sisa,
		})
	}

	helpers.Success(c, 200, "Data mata kuliah berhasil diambil", result, nil)
}
