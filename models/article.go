package models

import (
	"log"
	"time"

	"gorm.io/gorm"
)

type Article struct {
	ID         uint   `gorm:"primaryKey"`
	Title      string `gorm:"size:255;unique;not null"`
	Content    string `gorm:"type:text"`
	Source     string `gorm:"size:255"`
	ImageURL   string `gorm:"type:text"`
	Origin     string `gorm:"type:text"`
	CategoryID uint
	Category   Category
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func MigrateArticle(db *gorm.DB) {
	err := db.AutoMigrate(&Article{})
	if err != nil {
		log.Fatal(err)
	}
}
