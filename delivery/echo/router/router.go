package echowebframework

import (
	echowebframework "intelligentBI/delivery/echo/handler"
	"net/http"

	"github.com/labstack/echo/v5"
)

func Router() error {
	e := echo.New()

	e.GET("/healthcheck", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	userGroup := e.Group("/user")
	userGroup.POST("/register", echowebframework.UserRegisterHandler)
	userGroup.POST("/login", echowebframework.UserLoginHandler)
	userGroup.GET("/user_profile/:id", echowebframework.UserProfileHandler)

	if err := e.Start(":8080"); err != nil {
		return err
	}
	return nil

}
