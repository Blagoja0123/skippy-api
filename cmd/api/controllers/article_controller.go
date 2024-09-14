package controller

import (
	"net/http"
	"strconv"

	"github.com/Blagoja0123/skippy/cmd/api/service"
	"github.com/Blagoja0123/skippy/models"
	"github.com/labstack/echo/v4"
)

type ArticleController struct {
	service *service.ArticleService
}

func NewArticleController(service *service.ArticleService) *ArticleController {
	return &ArticleController{
		service: service,
	}
}

func (ac *ArticleController) Get(ctx echo.Context) error {

	params := make(map[string]string)

	if ctx.QueryParam("category_id") != "" {
		params["category_id"] = ctx.QueryParam("category_id")
	}
	if ctx.QueryParam("source") != "" {
		params["source"] = ctx.QueryParam("source")
	}
	if ctx.QueryParam("within_last") != "" {
		params["within_last"] = ctx.QueryParam("within_last")
	}
	if ctx.QueryParam("limit") != "" {
		params["limit"] = ctx.QueryParam("limit")
	}

	res, err := ac.service.Get(ctx.Request().Context(), params)

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

func (ac *ArticleController) GetByID(ctx echo.Context) error {

	id, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {
		return ctx.JSON(http.StatusBadRequest, "Invalid ID")
	}

	article, err := ac.service.GetByID(ctx.Request().Context(), id)

	if err != nil {
		return ctx.JSON(http.StatusNotFound, err)
	}

	return ctx.JSON(http.StatusOK, article)
}

func (ac *ArticleController) Create(ctx echo.Context) error {
	article := new(models.Article)

	if err := ctx.Bind(article); err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	if err := ac.service.Create(ctx.Request().Context(), article); err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	return ctx.JSON(http.StatusCreated, article)
}

func (ac *ArticleController) BulkDelete(ctx echo.Context) error {

	source := ctx.Param("source")

	if err := ac.service.BulkDelete(ctx.Request().Context(), source); err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
