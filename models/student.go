package models

import (
	"time"

	"gorm.io/gorm"
)

type Student struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"unique;not null" json:"user_id"`
	NIM         string         `gorm:"unique;not null;size:12" json:"nim"`
	Nama        string         `gorm:"not null" json:"nama"`
	Prodi       string         `gorm:"not null" json:"prodi"`
	Angkatan    int            `gorm:"not null" json:"angkatan"`
	IPKTerakhir float64        `gorm:"column:ipk_terakhir;default:0" json:"ipk_terakhir"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Enrollments []Enrollment   `gorm:"foreignKey:StudentID" json:"enrollments,omitempty"`
}
