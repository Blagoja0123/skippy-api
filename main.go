package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	controller "github.com/Blagoja0123/skippy/cmd/api/controllers"
	"github.com/Blagoja0123/skippy/cmd/api/handlers"
	"github.com/Blagoja0123/skippy/cmd/api/provider"
	"github.com/Blagoja0123/skippy/cmd/api/service"
	"github.com/Blagoja0123/skippy/models"
	"github.com/Blagoja0123/skippy/storage"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/robfig/cron"
)

func main() {
	// Echo instance

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal(err)
	}
	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASSWORD"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}

	db, err := storage.NewConnection(config)

	if err != nil {
		fmt.Println("DB connection error!")
		log.Fatal(err)
	}

	models.MigrateUser(db)
	models.MigrateCategories(db)
	models.MigrateArticle(db)
	models.MigrateArticleCategories(db)

	e := echo.New()

	// Middleware
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", hello)
	e.GET("/health", handlers.HealthCheckHandler)

	authService := service.NewAuthService(db)
	authController := controller.NewAuthController(authService)

	userRoute := e.Group("/auth")
	userRoute.POST("/register", authController.Register)
	// userRoute.GET("/", controller.GetUsers)

	articleService := service.NewArticleService(db)
	articleController := controller.NewArticleController(articleService)

	articleRoute := e.Group("/articles")
	articleRoute.GET("", articleController.Get)
	articleRoute.GET("/:id", articleController.GetByID)
	articleRoute.POST("", articleController.Create)
	articleRoute.DELETE("/:source", articleController.BulkDelete)

	categoryService := service.NewCategoryService(db)
	categoryController := controller.NewCategoryController(*categoryService)

	categoryRoute := e.Group("/categories")
	categoryRoute.POST("", categoryController.Create)
	categoryRoute.GET("", categoryController.Get)
	categoryRoute.DELETE("", categoryController.Delete)

	for _, route := range e.Routes() {
		log.Printf("%s %s\n", route.Method, route.Path)
	}

	loc, _ := time.LoadLocation("Europe/Skopje")

	cronGuardian := cron.NewWithLocation(loc)
	cronGuardian.AddFunc("@every 6h", func() {
		sections := []string{"politics", "business", "sport", "film", "technology", "science"}
		queryParams := map[string]string{
			"page-size":   "50",
			"show-fields": "body",
			"show-tags":   "keyword",
		}
		gp := provider.NewGuardianProvider("https://content.guardianapis.com/search", sections, queryParams, os.Getenv("GUARDIAN_KEY"), categoryService, articleService)
		data, _ := gp.GetArticles(context.Background())
		log.Print(data)
		log.Println()
	})
	cronGuardian.Start()

	cronNYT := cron.NewWithLocation(loc)
	cronNYT.AddFunc("@every 6h", func() {
		sections := []string{"business", "politics", "sports", "movies", "technology", "science"}

		nyt := provider.NewNYTimesProvider("https://api.nytimes.com/svc/topstories/v2/", sections, nil, os.Getenv("NYT_KEY"), categoryService, articleService)

		data, err := nyt.GetArticles(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		log.Print(data)
		log.Println()
	})
	cronNYT.Start()

	e.Logger.Fatal(e.Start(":8000"))
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
