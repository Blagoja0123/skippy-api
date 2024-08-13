package models

import (
	"log"

	"gorm.io/gorm"
)

type ArticleCategories struct {
	ArticleID  uint `gorm:"primaryKey"`
	CategoryID uint `gorm:"primaryKey"`
}

func MigrateArticleCategories(db *gorm.DB) {
	err := db.AutoMigrate(&ArticleCategories{})
	if err != nil {
		log.Fatal(err)
	}
}
