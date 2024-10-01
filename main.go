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
	"github.com/Blagoja0123/skippy/cmd/api/scraper"
	"github.com/Blagoja0123/skippy/cmd/api/service"
	"github.com/Blagoja0123/skippy/custom_middleware"
	"github.com/Blagoja0123/skippy/models"
	"github.com/Blagoja0123/skippy/storage"
	"github.com/joho/godotenv"

	// echojwt "github.com/labstack/echo-jwt/v4"
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
	e.Use(middleware.CORS())
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", hello)
	e.GET("/health", handlers.HealthCheckHandler)

	authService := service.NewAuthService(db)
	userService := service.NewUserService(db)
	authController := controller.NewAuthController(authService, userService)

	authRoute := e.Group("/auth")
	authRoute.POST("/register", authController.Register)
	authRoute.POST("/login", authController.Login)

	articleService := service.NewArticleService(db)
	articleController := controller.NewArticleController(articleService)

	userController := controller.NewUserController(userService, articleService)

	userRoute := e.Group("/users", custom_middleware.AuthMiddleware)
	userRoute.GET("", userController.GetByID)
	userRoute.PATCH("", userController.Update)
	userRoute.GET("/feed", userController.GetFeed)
	userRoute.PATCH("/likes/add", userController.UpdateLiked)
	userRoute.PATCH("/likes/remove", userController.UpdateDisliked)

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

	loc, _ := time.LoadLocation("Europe/Skopje")

	cronPV := cron.NewWithLocation(loc)
	cronPV.AddFunc("@every 4h", func() {
		sectionsGD := []string{"politics", "business", "sport", "film", "technology", "science"}
		queryParams := map[string]string{
			"page-size":   "50",
			"show-fields": "body",
			"show-tags":   "keyword",
		}
		gp := provider.NewGuardianProvider("https://content.guardianapis.com/search", sectionsGD, queryParams, os.Getenv("GUARDIAN_KEY"), categoryService, articleService)

		sectionsNYT := []string{"business", "politics", "sports", "movies", "technology", "science"}

		nyt := provider.NewNYTimesProvider("https://api.nytimes.com/svc/topstories/v2/", sectionsNYT, nil, os.Getenv("NYT_KEY"), categoryService, articleService)

		providers := []provider.Provider{gp, nyt}

		for _, pv := range providers {
			articles, err := pv.GetArticles(context.Background())

			if err != nil {
				log.Fatal(err)
			}

			log.Printf("Wrote: %d\n", len(articles))
		}
	})
	cronPV.Start()

	cronSC := cron.NewWithLocation(loc)
	cronSC.AddFunc("@every 4h", func() {

		var articles []models.Article
		espn := scraper.NewESPNScraper("https://www.espn.com/f1", []string{"f1", "nba", "nfl", "soccer", "mma", "mlb", "nhl", "nascar", "boxing"})
		fb := scraper.NewForbesBSNScraper("https://www.forbes.com/", []string{"law", "manufacturing", "energy", "policy", "retail", "fintech", "investing", "markets"})
		ft := scraper.NewForbesTechScraper("https://www.forbes.com/", []string{"cloud", "ai", "big-data", "cybersecurity", "consumer-tech"})

		collectors := []scraper.Scraper{
			espn,
			fb,
			ft,
		}

		for _, scr := range collectors {
			data := scr.Collect()
			articles = append(articles, data...)
		}

		for _, article := range articles {
			articleService.Create(context.Background(), &article)
		}
		fmt.Println("Wrote scraper articles")
	})
	cronSC.Start()

	e.Logger.Fatal(e.Start(":8000"))
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
