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

func (cc *CategoryController) Get(ctx echo.Context) error {

	res, err := cc.service.GetCategories(ctx.Request().Context())

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	data := map[string]interface{}{
		"status": "OK",
		"total":  len(res),
		"data":   res,
	}

	return ctx.JSON(http.StatusOK, data)
}

func (cc *CategoryController) Update(ctx echo.Context) error {

	var category models.Category

	if err := ctx.Bind(&category); err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	if err := cc.service.UpdateCategory(ctx.Request().Context(), &category); err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	data := map[string]interface{}{
		"status": "OK",
	}

	return ctx.JSON(http.StatusOK, data)
}

func (cc *CategoryController) Delete(ctx echo.Context) error {

	var category models.Category

	if err := ctx.Bind(&category); err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	if err := cc.service.DeleteCategory(ctx.Request().Context(), category.ID); err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	data := map[string]interface{}{
		"status": "OK",
	}

	return ctx.JSON(http.StatusOK, data)
}
