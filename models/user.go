package models

import (
	"log"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID              uint   `gorm:"primaryKey"`
	Username        string `gorm:"size:255;unique;not null"`
	PasswordHash    []byte `gorm:"type:bytea"`
	LikedCategoryID *uint
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Migration function for User model
func MigrateUser(db *gorm.DB) {
	err := db.AutoMigrate(&User{})
	if err != nil {
		log.Fatal(err)
	}
}
