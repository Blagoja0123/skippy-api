package service

import (
	"context"

	"github.com/Blagoja0123/skippy/models"
	"gorm.io/gorm"
)

type CategoryService struct {
	db *gorm.DB
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{
		db: db,
	}
}

func (cs *CategoryService) GetCategories(ctx context.Context) ([]models.Category, error) {

	var categories []models.Category

	if err := cs.db.Model(&models.Category{}).Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

func (cs *CategoryService) GetByName(ctx context.Context, name string) (*models.Category, error) {
	var category models.Category

	if err := cs.db.WithContext(ctx).Model(&models.Category{}).Where("name LIKE ?", "%"+name+"%").First(&category).Error; err != nil {
		return nil, err
	}

	return &category, nil
}

func (cs *CategoryService) AddCategory(ctx context.Context, category *models.Category) error {

	return cs.db.WithContext(ctx).Model(&models.Category{}).Create(&category).Error
}

func (cs *CategoryService) UpdateCategory(ctx context.Context, body *models.Category) error {
	return cs.db.WithContext(ctx).Model(&models.Category{}).Save(body).Error
}

func (cs *CategoryService) DeleteCategory(ctx context.Context, id uint) error {

	return cs.db.WithContext(ctx).Model(&models.Category{}).Delete(&models.Category{}, id).Error
}
