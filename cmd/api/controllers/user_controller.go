package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Blagoja0123/skippy/cmd/api/service"
	"github.com/Blagoja0123/skippy/models"
	"github.com/labstack/echo/v4"
)

type UserController struct {
	service    *service.UserService
	artService *service.ArticleService
}

func NewUserController(service *service.UserService, artService *service.ArticleService) *UserController {
	return &UserController{
		service:    service,
		artService: artService,
	}
}

func (uc *UserController) GetByID(ctx echo.Context) error {

	id := ctx.Get("userID")
	userId := uint(id.(float64))

	log.Println(userId)
	user, err := uc.service.GetByID(ctx.Request().Context(), userId)

	if err != nil {
		return ctx.JSON(http.StatusNotFound, err)
	}

	return ctx.JSON(http.StatusOK, user)
}

func (uc *UserController) GetFeed(ctx echo.Context) error {

	id := ctx.Get("userID")

	userId, _ := id.(float64)
	log.Println(userId)
	user, err := uc.service.GetByID(ctx.Request().Context(), uint(userId))

	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	categories := ""

	for _, cat := range user.LikedCategoryID {
		categories += strconv.Itoa(int(cat)) + ","
	}

	if categories == "" {
		return ctx.JSON(http.StatusInternalServerError, errors.New("user has no liked categories"))
	}
	if len(user.LikedCategoryID) == 0 {
		return ctx.JSON(http.StatusInternalServerError, errors.New("user has no liked categories"))
	}
	categories = strings.TrimRight(categories, ", ")

	articles, err := uc.artService.Get(ctx.Request().Context(), map[string]string{
		"category_id": categories,
	})

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	data := map[string]interface{}{
		"status": "OK",
		"total":  len(articles),
		"data":   articles,
	}

	return ctx.JSON(http.StatusOK, data)
}

func (uc *UserController) Update(ctx echo.Context) error {

	var user models.User
	if err := ctx.Bind(&user); err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}
	if err := uc.service.Update(ctx.Request().Context(), &user); err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	data := map[string]interface{}{
		"status": "OK",
	}

	return ctx.JSON(http.StatusOK, data)
}

func (uc *UserController) UpdateLiked(ctx echo.Context) error {

	var cat struct {
		ID            uint  `json:"id"`
		LikedCategory int64 `json:"liked_category"`
	}
	id := ctx.Get("userID")

	userId, ok := id.(float64)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid user ID"})
	}

	if err := ctx.Bind(&cat); err != nil {
		log.Println("Error binding")
		return ctx.JSON(http.StatusBadRequest, err)
	}
	if err := uc.service.UpdateLike(ctx.Request().Context(), cat.LikedCategory, uint(userId)); err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	data := map[string]interface{}{
		"status": "OK",
	}

	return ctx.JSON(http.StatusOK, data)
}

func (uc *UserController) UpdateDisliked(ctx echo.Context) error {

	var cat struct {
		ID            uint  `json:"id"`
		LikedCategory int64 `json:"liked_category"`
	}
	id := ctx.Get("userID")

	userId, _ := id.(float64)

	if err := ctx.Bind(&cat); err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}
	if err := uc.service.UpdateDislike(ctx.Request().Context(), cat.LikedCategory, uint(userId)); err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	data := map[string]interface{}{
		"status": "OK",
	}

	return ctx.JSON(http.StatusOK, data)
}
