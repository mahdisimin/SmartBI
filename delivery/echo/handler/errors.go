package echowebframework

import (
	"errors"
	"log"
	"net/http"

	"intelligentBI/service"
	"intelligentBI/service/export"

	"github.com/labstack/echo/v5"
)

// httpError maps a service error to an HTTP error. Only errors that describe
// the caller's own request reach the client verbatim; anything else (database
// or driver failures) is logged with `op` for context and answered with a
// generic 500, so internals never leak.
func httpError(op string, err error) error {
	var validationErr service.ValidationError
	switch {
	case errors.As(err, &validationErr):
		return echo.NewHTTPError(http.StatusBadRequest, validationErr.Error())
	case errors.Is(err, service.ErrUnauthenticated):
		return echo.NewHTTPError(http.StatusUnauthorized, service.ErrUnauthenticated.Error())
	case errors.Is(err, export.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, export.ErrForbidden.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		return echo.NewHTTPError(http.StatusUnauthorized, service.ErrInvalidCredentials.Error())
	case errors.Is(err, service.ErrPhoneNumberExists):
		return echo.NewHTTPError(http.StatusConflict, service.ErrPhoneNumberExists.Error())
	case errors.Is(err, service.ErrUserNotFound):
		return echo.NewHTTPError(http.StatusNotFound, service.ErrUserNotFound.Error())
	default:
		log.Printf("error on HTTP request , %s error : %v", op, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}
