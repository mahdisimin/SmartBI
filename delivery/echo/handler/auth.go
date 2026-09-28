package echowebframework

import (
	"net/http"
	"time"

	"intelligentBI/service"

	"github.com/labstack/echo/v5"
)

// SessionCookieName is the cookie carrying the opaque login-session token.
// It is HttpOnly: the frontend cannot read it, and learns whether it is
// logged in from GET /user/user_profile/me (200 vs 401).
const SessionCookieName = "smartbi_session"

// userIDContextKey is where RequireSession stores the authenticated user ID.
const userIDContextKey = "auth.userID"

// RequireSession rejects requests without a valid session cookie with 401,
// and stores the session's user ID in the context for handlers.
func (h *Handler) RequireSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		cookie, err := c.Cookie(SessionCookieName)
		if err != nil {
			return httpError("authenticate", service.ErrUnauthenticated)
		}
		userID, err := h.SessionService.Authenticate(cookie.Value)
		if err != nil {
			return httpError("authenticate", err)
		}
		c.Set(userIDContextKey, userID)
		return next(c)
	}
}

// currentUserID returns the user RequireSession authenticated. Handlers
// behind RequireSession use it; ok is false when the middleware did not run,
// and handlers must then refuse the request (fail closed).
func currentUserID(c *echo.Context) (int64, bool) {
	userID, ok := c.Get(userIDContextKey).(int64)
	return userID, ok && userID > 0
}

func (h *Handler) setSessionCookie(c *echo.Context, token string, expiresAt time.Time) {
	c.SetCookie(&http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   h.CookieSecure,
		// Lax: sent on same-site requests (the dashboard and API on the same
		// site, any port) and top-level navigations, never on cross-site
		// POSTs — which also blocks CSRF against logout.
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearSessionCookie(c *echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}
