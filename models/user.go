package models

import (
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/lib/pq"
)

type User struct {
	ID              uint          `gorm:"primaryKey"`
	Username        string        `gorm:"size:255;unique;not null"`
	PasswordHash    []byte        `gorm:"type:bytea"`
	ImageURL        string        `gorm:"type:text"`
	LikedCategoryID pq.Int64Array `gorm:"type:bigint[]"`
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
