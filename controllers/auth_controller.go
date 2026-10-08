package controllers

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"

	"siakad-mini/config"
	"siakad-mini/helpers"
	"siakad-mini/models"
)

var jsonFieldNames = map[string]string{
	"Email":         "email",
	"Password":      "password",
	"NIM":           "nim",
	"Nama":          "nama",
	"Prodi":         "prodi",
	"Angkatan":      "angkatan",
	"IPKTerakhir":   "ipk_terakhir",
	"CourseID":      "course_id",
	"TahunAkademik": "tahun_akademik",
}

func mapBindErrors(err error) map[string][]string {
	result := map[string][]string{}
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		result["error"] = []string{err.Error()}
		return result
	}
	for _, fe := range verrs {
		field := fe.Field()
		key := strings.ToLower(field)
		if j, ok := jsonFieldNames[field]; ok {
			key = j
		}
		result[key] = append(result[key], bindMessage(field, fe.Tag(), fe.Param()))
	}
	return result
}

func bindMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return field + " wajib diisi"
	case "email":
		return "Format email tidak valid"
	case "min":
		if field == "Password" {
			return "Password minimal " + param + " karakter"
		}
		return field + " minimal " + param
	case "max":
		return field + " maksimal " + param
	case "len":
		return field + " harus " + param + " digit"
	case "numeric":
		return field + " harus berupa angka"
	default:
		return field + " tidak valid"
	}
}

var (
	loginAttempts = make(map[string][]time.Time)
	mu            sync.Mutex
)

func isRateLimited(email string) bool {
	mu.Lock()
	defer mu.Unlock()
	now := time.Now()
	attempts := loginAttempts[email]
	valid := []time.Time{}
	for _, t := range attempts {
		if now.Sub(t) < time.Minute {
			valid = append(valid, t)
		}
	}
	loginAttempts[email] = valid
	return len(valid) >= 5
}

func recordAttempt(email string) {
	mu.Lock()
	defer mu.Unlock()
	loginAttempts[email] = append(loginAttempts[email], time.Now())
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.ValidationError(c, mapBindErrors(err))
		return
	}

	if isRateLimited(input.Email) {
		helpers.Error(c, 429, "Terlalu banyak percobaan login, coba lagi nanti", nil)
		return
	}

	var user models.User
	if err := config.DB.Preload("Student").Where("email = ?", input.Email).First(&user).Error; err != nil {
		recordAttempt(input.Email)
		helpers.Error(c, 401, "Kredensial salah", nil)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		recordAttempt(input.Email)
		helpers.Error(c, 401, "Kredensial salah", nil)
		return
	}

	if user.Role == "mahasiswa" && (user.Student == nil || user.Student.DeletedAt.Valid) {
		helpers.Error(c, 401, "Akun tidak aktif", nil)
		return
	}

	token, expiresIn, err := helpers.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		helpers.Error(c, 500, "Gagal membuat token", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "Login berhasil",
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   expiresIn,
		"data": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func Me(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := config.DB.Preload("Student").First(&user, userID).Error; err != nil {
		helpers.Error(c, 401, "User tidak ditemukan", nil)
		return
	}

	data := gin.H{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	}

	if user.Role == "mahasiswa" && user.Student != nil {
		data["student"] = gin.H{
			"nim":      user.Student.NIM,
			"nama":     user.Student.Nama,
			"prodi":    user.Student.Prodi,
			"angkatan": user.Student.Angkatan,
		}
	}

	helpers.Success(c, 200, "Profil berhasil diambil", data, nil)
}
