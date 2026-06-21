package echowebframework

import (
	echowebframework "intelligentBI/delivery/echo/handler"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func Router() error {
	e := echo.New()

	e.GET("/healthcheck", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	userGroup := e.Group("/user")
	userGroup.POST("/register", echowebframework.UserRegisterHandler)
	userGroup.POST("/login", echowebframework.UserLoginHandler)
	userGroup.GET("/user_profile/:id", echowebframework.UserProfileHandler)

	if err := e.Start(":8091"); err != nil {
		return err
	}
	return nil

}
