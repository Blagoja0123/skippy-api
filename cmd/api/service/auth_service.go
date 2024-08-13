package service

import (
	"context"
	"errors"

	"github.com/Blagoja0123/skippy/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

type AuthReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		db: db,
	}
}

func (as *AuthService) Register(ctx context.Context, body *models.User) error {
	return as.db.WithContext(ctx).Model(&models.User{}).Create(body).Error
}

func (as *AuthService) Login(ctx context.Context, body *AuthReq) error {

	var dbUser models.User

	if err := as.db.WithContext(ctx).Model(&models.User{}).Where("username = ?", body.Username).First(&dbUser).Error; err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(body.Password)); err != nil {
		return errors.New("Invalid password, please try again")
	}

	return nil
}
