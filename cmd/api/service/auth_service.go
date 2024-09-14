package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Blagoja0123/skippy/models"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

type AuthReq struct {
	Username  string        `json:"username"`
	Password  string        `json:"password"`
	LikedCats pq.Int64Array `json:"liked_categories"`
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		db: db,
	}
}

func (as *AuthService) Register(ctx context.Context, body *models.User) error {
	int64Array := make(pq.Int64Array, len(body.LikedCategoryID))
	for i, v := range body.LikedCategoryID {
		int64Array[i] = int64(v)
	}
	body.LikedCategoryID = int64Array
	fmt.Printf("Type of cats: %T\n", body.LikedCategoryID)

	return as.db.WithContext(ctx).Model(&models.User{}).Create(&body).Error
}

func (as *AuthService) Login(ctx context.Context, body *AuthReq) error {

	var dbUser models.User

	if err := as.db.WithContext(ctx).Model(&models.User{}).Where("username = ?", body.Username).First(&dbUser).Error; err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(body.Password)); err != nil {
		return errors.New("invalid password, please try again")
	}

	return nil
}
