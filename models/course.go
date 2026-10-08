package models

type Course struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	KodeMK      string       `gorm:"unique;not null" json:"kode_mk"`
	NamaMK      string       `gorm:"not null" json:"nama_mk"`
	SKS         int          `gorm:"not null" json:"sks"`
	Semester    int          `gorm:"not null" json:"semester"`
	Kuota       int          `gorm:"not null" json:"kuota"`
	Enrollments []Enrollment `gorm:"foreignKey:CourseID" json:"-"`
}
