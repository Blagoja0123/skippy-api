package service

import (
	"context"
	"fmt"

	"github.com/Blagoja0123/skippy/models"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

func (us *UserService) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User

	if err := us.db.WithContext(ctx).Select("id", "username", "image_url", "liked_category_id").First(&user, id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (us *UserService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User

	if err := us.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	fmt.Println("USER ID:", user.ID)
	return &user, nil
}

func (us *UserService) Create(ctx context.Context, body *models.User) error {
	return us.db.WithContext(ctx).Model(&models.User{}).Create(body).Error
}

func (us *UserService) Update(ctx context.Context, body *models.User) error {
	return us.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", body.ID).Updates(body).Error
}

func (us *UserService) Delete(ctx context.Context, id int) error {
	return us.db.WithContext(ctx).Model(&models.User{}).Delete(&models.User{}, id).Error
}

func (us *UserService) UpdateLike(ctx context.Context, cat int64, id uint) error {
	query := `
	UPDATE users 
	SET liked_category_id = COALESCE(liked_category_id, '{}') || ?
	WHERE id = ?
	`
	return us.db.WithContext(ctx).Model(&models.User{}).Exec(query, pq.Int64Array{cat}, id).Error
}

func (us *UserService) UpdateDislike(ctx context.Context, cat int64, id uint) error {
	query := `
	UPDATE users 
	SET liked_category_id = ARRAY_REMOVE(liked_category_id, ?)
	WHERE id = ?
	`
	return us.db.WithContext(ctx).Model(&models.User{}).Exec(query, cat, id).Error
}
