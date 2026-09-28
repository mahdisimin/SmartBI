package echowebframework

import (
	echowebframework "intelligentBI/delivery/echo/handler"
	"intelligentBI/pkg"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func Router(db *sqlx.DB) error {
	sessionCfg, err := pkg.LoadSessionConfig()
	if err != nil {
		return err
	}
	h := echowebframework.NewHandler(db, sessionCfg)
	e := echo.New()

	e.GET("/healthcheck", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: pkg.CORSAllowedOrigins(),
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// The session cookie is sent cross-origin (dashboard on :5173, API on
		// :8091); browsers only include it — and expose the response — when
		// the server allows credentials. Safe because origins are an explicit
		// allow-list, never "*".
		AllowCredentials: true,
	}))

	h.RegisterRoutes(e)

	if err := e.Start(":8091"); err != nil {
		return err
	}
	return nil

}
