package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	controller "github.com/Blagoja0123/skippy/cmd/api/controllers"
	"github.com/Blagoja0123/skippy/cmd/api/handlers"
	"github.com/Blagoja0123/skippy/cmd/api/provider"
	"github.com/Blagoja0123/skippy/cmd/api/scraper"
	"github.com/Blagoja0123/skippy/cmd/api/service"
	"github.com/Blagoja0123/skippy/models"
	"github.com/Blagoja0123/skippy/storage"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"

	// echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/robfig/cron"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ctx.Response().Header().Add("User", "Authorization")
		authHeader := ctx.Request().Header.Get("Authorization")
		// fmt.Println(authHeader)
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized access"})
		}

		authHeaderSplit := strings.Split(authHeader, " ")
		accessToken := authHeaderSplit[1]

		token, err := jwt.Parse(accessToken, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}

			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})

		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		// Validate the token claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx.Set("userID", claims["id"])
			ctx.Set("username", claims["username"])
		} else {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized access"})
		}

		return next(ctx)
	}
}

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
	// e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	// 	AllowOrigins: []string{"http://localhost:4321/", "http://localhost:4321"},
	// 	AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
	// 	AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	// }))

	e.Use(middleware.CORS())
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// e.Use(echojwt.WithConfig(echojwt.Config{
	// 	SigningKey: []byte(os.Getenv("JWT_SECRET_KEY")),
	// }))

	// Routes
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

	userRoute := e.Group("/users", AuthMiddleware)
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

	// for _, route := range e.Routes() {
	// 	log.Printf("%s %s\n", route.Method, route.Path)
	// }

	loc, _ := time.LoadLocation("Europe/Skopje")

	cronGuardian := cron.NewWithLocation(loc)
	cronGuardian.AddFunc("@every 4h", func() {
		sections := []string{"politics", "business", "sport", "film", "technology", "science"}
		queryParams := map[string]string{
			"page-size":   "50",
			"show-fields": "body",
			"show-tags":   "keyword",
		}
		gp := provider.NewGuardianProvider("https://content.guardianapis.com/search", sections, queryParams, os.Getenv("GUARDIAN_KEY"), categoryService, articleService)
		_, err := gp.GetArticles(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Wrote Guardian articles")
	})
	cronGuardian.Start()

	cronNYT := cron.NewWithLocation(loc)
	cronNYT.AddFunc("@every 4h", func() {
		sections := []string{"business", "politics", "sports", "movies", "technology", "science"}

		nyt := provider.NewNYTimesProvider("https://api.nytimes.com/svc/topstories/v2/", sections, nil, os.Getenv("NYT_KEY"), categoryService, articleService)

		_, err := nyt.GetArticles(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Wrote NYT articles")
	})
	cronNYT.Start()
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
