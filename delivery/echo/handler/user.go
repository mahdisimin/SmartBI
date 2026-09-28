package echowebframework

import (
	"intelligentBI/service"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
)

func (h *Handler) UserRegisterHandler(c *echo.Context) error {
	var userRegReq service.UserRegisterRequest
	var userRegResp service.UserRegisterResponse
	var userServ = h.UserService

	if err := c.Bind(&userRegReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	userRegRespTemp, err := userServ.Register(userRegReq)
	if err != nil {
		return httpError("register", err)
	} else {
		userRegResp = userRegRespTemp
	}
	return c.JSON(http.StatusOK, userRegResp)
}

// UserLoginHandler checks the credentials and, on success, starts a session:
// the response sets the HttpOnly session cookie.
func (h *Handler) UserLoginHandler(c *echo.Context) error {
	var userLoginReq service.UserLoginRequest
	var userLoginResp service.UserLoginResponse
	var userServ = h.UserService

	if err := c.Bind(&userLoginReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	userLoginRespTemp, err := userServ.Login(userLoginReq)
	if err != nil {
		return httpError("login", err)
	} else {
		userLoginResp = userLoginRespTemp
	}

	token, expiresAt, err := h.SessionService.Create(userLoginResp.UserId)
	if err != nil {
		return httpError("login: create session", err)
	}
	h.setSessionCookie(c, token, expiresAt)

	return c.JSON(http.StatusOK, userLoginResp)
}

// UserLogoutHandler ends the current session, if any, and clears the cookie.
// It always succeeds, so a client can call it without checking its state.
func (h *Handler) UserLogoutHandler(c *echo.Context) error {
	if cookie, err := c.Cookie(SessionCookieName); err == nil {
		if err := h.SessionService.Revoke(cookie.Value); err != nil {
			log.Printf("error on HTTP request , logout error : %v", err)
		}
	}
	h.clearSessionCookie(c)
	return c.NoContent(http.StatusNoContent)
}

// UserMeHandler returns the logged-in user's own profile. Behind
// RequireSession: 401 means "not logged in", which is how the frontend checks
// its session state.
func (h *Handler) UserMeHandler(c *echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return httpError("user profile", service.ErrUnauthenticated)
	}
	return h.profile(c, userID)
}

// UserProfileHandler returns a profile by ID, but only the caller's own:
// another user's ID gets 403. Prefer GET /user/user_profile/me.
func (h *Handler) UserProfileHandler(c *echo.Context) error {
	userIdStr := c.Param("id")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil || userId <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "user id must be a positive integer")
	}

	currentID, ok := currentUserID(c)
	if !ok {
		return httpError("user profile", service.ErrUnauthenticated)
	}
	if userId != currentID {
		return echo.NewHTTPError(http.StatusForbidden, "cannot access another user's profile")
	}
	return h.profile(c, userId)
}

func (h *Handler) profile(c *echo.Context, userID int64) error {
	userProfileResp, err := h.UserService.Profile(service.UserProfileRequest{UserID: userID})
	if err != nil {
		return httpError("user profile", err)
	}
	return c.JSON(http.StatusOK, userProfileResp)
}
