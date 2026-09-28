package echowebframework

import "github.com/labstack/echo/v5"

// RegisterRoutes mounts the API's user and export routes on e, with
// RequireSession on everything that needs a login.
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	userGroup := e.Group("/user")
	userGroup.POST("/register", h.UserRegisterHandler)
	userGroup.POST("/login", h.UserLoginHandler)
	userGroup.POST("/logout", h.UserLogoutHandler)
	// "/me" is a static route, so it takes precedence over "/:id".
	userGroup.GET("/user_profile/me", h.UserMeHandler, h.RequireSession)
	userGroup.GET("/user_profile/:id", h.UserProfileHandler, h.RequireSession)

	exportGroup := e.Group("/export", h.RequireSession)
	exportGroup.GET("/:product", h.ExportHandler)
}
