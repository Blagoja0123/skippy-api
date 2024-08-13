package service

import (
	"context"
	"errors"

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

func (cs *CategoryService) GetCategories() (interface{}, error) {

	var categories []*models.Category

	if err := cs.db.Model(&models.Category{}).Find(&categories).Error; err != nil {
		return nil, err
	}

	res := map[string]interface{}{
		"data": categories,
	}

	return res, nil
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

func (cs *CategoryService) UpdateCategory(cat models.Category) (interface{}, error) {

	result := cs.db.Model(&models.Category{}).Where("id = ?", cat.ID).Updates(map[string]interface{}{
		"name": cat.Name,
	})

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("Category not found")
	}

	return cat, nil
}

func (cs *CategoryService) DeleteCategory(id uint) (interface{}, error) {

	result := cs.db.Delete(&models.Category{}, id)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("Category not found")
	}

	data := map[string]interface{}{
		"data": true,
	}

	return data, nil
}
