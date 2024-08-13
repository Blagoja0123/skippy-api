package controller

// import (
// 	"log"
// 	"net/http"
// 	"os"
// 	"time"

// 	"github.com/Blagoja0123/skippy/cmd/api/provider"
// 	"github.com/Blagoja0123/skippy/cmd/api/service"
// 	"github.com/Blagoja0123/skippy/storage"
// 	"github.com/labstack/echo/v4"
// )

// func GetGuardian(ctx echo.Context) error {
// 	sections := []string{"politics", "business", "sport", "film", "technology", "science"}

// 	queryParams := map[string]string{
// 		"page-size":   "1",
// 		"show-fields": "body",
// 		"show-tags":   "keyword",
// 	}

// 	gp := provider.NewGuardianProvider("https://content.guardianapis.com/search", sections, queryParams, os.Getenv("GUARDIAN_KEY"))
// 	start := time.Now()
// 	data, err := gp.GetArticles(ctx.Request().Context())
// 	end := time.Now()
// 	log.Printf("Guardian article writing took: %d ms", end.UnixMilli()-start.UnixMilli())
// 	if err != nil {
// 		return ctx.JSON(http.StatusInternalServerError, err)
// 	}

// 	return ctx.JSON(http.StatusOK, data)
// }

// func GetNYT(ctx echo.Context) error {
// 	sections := []string{"business", "politics", "sports", "movies", "technology", "science"}

// 	nyt := provider.NewNYTimesProvider("https://api.nytimes.com/svc/topstories/v2/", sections, nil, os.Getenv("NYT_KEY"))

// 	data, err := nyt.GetArticles(ctx.Request().Context())

// 	if err != nil {
// 		return ctx.JSON(http.StatusInternalServerError, err)
// 	}

// 	return ctx.JSON(http.StatusOK, data)
// }

// func GetNewsData(ctx echo.Context) error {
// 	sections := []string{"business", "politics", "sports", "technology"}

// 	queryParams := map[string]string{
// 		"language":  "en",
// 		"page_size": "2",
// 	}

// 	np := provider.NewNewsProvider(
// 		"https://api.currentsapi.services/v1/latest-news",
// 		sections,
// 		queryParams,
// 		os.Getenv("CURRENTS_KEY"),
// 		*service.NewCategoryService(storage.DB()),
// 		*service.NewArticleService(storage.DB()),
// 	)

// 	data, err := np.GetArticles(ctx.Request().Context())

// 	if err != nil {
// 		return ctx.JSON(http.StatusInternalServerError, err)
// 	}

// 	return ctx.JSON(http.StatusOK, data)
// }
