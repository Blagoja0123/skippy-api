package controller

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Blagoja0123/skippy/cmd/api/service"
	"github.com/Blagoja0123/skippy/models"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	service     *service.AuthService
	userService *service.UserService
}

func NewAuthController(service *service.AuthService, userService *service.UserService) *AuthController {
	return &AuthController{
		service:     service,
		userService: userService,
	}
}

func (ac *AuthController) Register(ctx echo.Context) error {

	var req struct {
		Username  string        `json:"username"`
		Password  string        `json:"password"`
		LikedCats pq.Int64Array `json:"liked_categories"`
		ImageURL  string        `json:"image_url"`
	}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}
	fmt.Println(req.ImageURL)
	// fmt.Print("Type of LikedCats: ", req.LikedCats)
	// return ctx.JSON(http.StatusBadRequest, req.LikedCats)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	user := &models.User{
		Username:        req.Username,
		PasswordHash:    hashedPassword,
		LikedCategoryID: req.LikedCats,
		ImageURL:        req.ImageURL,
	}
	fmt.Println("Liked categories: ", user.LikedCategoryID)
	if err := ac.service.Register(ctx.Request().Context(), user); err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	return ctx.JSON(http.StatusOK, map[string]string{
		"status": "OK",
	})
}

func (ac *AuthController) Login(ctx echo.Context) error {

	var req service.AuthReq

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	if err := ac.service.Login(ctx.Request().Context(), &req); err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	user, err := ac.userService.GetByUsername(ctx.Request().Context(), req.Username)

	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(), // 30-day expiry for refresh token
	})

	signedRefresh, err := refreshToken.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"status":        "OK",
		"bearer_token":  signed,
		"refresh_token": signedRefresh,
		"expires":       time.Now().Add(time.Hour * 24).Unix(),
	})

}
