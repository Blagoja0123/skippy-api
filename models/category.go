package models

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// Category represents the category table
type Category struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:255"`
	ParentID  *uint  // Foreign key to Category
	CreatedAt time.Time
	UpdatedAt time.Time
}

func MigrateCategories(db *gorm.DB) {
	err := db.AutoMigrate(&Category{})
	if err != nil {
		log.Fatal(err)
	}
}
