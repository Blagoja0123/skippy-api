package controller

import (
	"net/http"

	"github.com/Blagoja0123/skippy/cmd/api/service"
	"github.com/Blagoja0123/skippy/models"
	"github.com/labstack/echo/v4"
)

type CategoryController struct {
	service service.CategoryService
}

func NewCategoryController(service service.CategoryService) *CategoryController {
	return &CategoryController{
		service: service,
	}
}

func (cc *CategoryController) Create(ctx echo.Context) error {

	category := new(models.Category)

	if err := ctx.Bind(category); err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	if err := cc.service.AddCategory(ctx.Request().Context(), category); err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	return ctx.JSON(http.StatusCreated, category)

}
