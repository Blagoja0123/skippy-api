package custom_middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
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
