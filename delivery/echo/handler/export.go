package echowebframework

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"intelligentBI/pkg"
	"intelligentBI/service"
	"intelligentBI/service/export"

	"github.com/labstack/echo/v5"
)

// ExportHandler returns a product's dashboard data. Behind RequireSession;
// the export service also checks the user has been granted the product.
func (h *Handler) ExportHandler(c *echo.Context) error {
	productName := c.Param("product")
	if productName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "product is required")
	}

	product, err := pkg.ParseProductList(productName)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	days := splitCSV(c.QueryParam("days"))
	for _, day := range days {
		if _, err := time.Parse("2006-01-02", day); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("days: %q is not a YYYY-MM-DD date", day))
		}
	}
	userIDs, err := splitInt64CSV(c.QueryParam("users"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "users: "+err.Error())
	}

	filters := export.Filters{
		Days:    days,
		Modules: splitCSV(c.QueryParam("modules")),
		Methods: splitCSV(c.QueryParam("methods")),
		UserIDs: userIDs,
	}

	userID, ok := currentUserID(c)
	if !ok {
		return httpError("export", service.ErrUnauthenticated)
	}

	response, err := h.ExportService.Export(export.ExportRequest{UserID: userID, Product: product, Filters: filters})
	if err != nil {
		return httpError("export", err)
	}

	return c.JSON(http.StatusOK, response)
}

// splitCSV parses a comma-separated query param into a trimmed, non-empty
// value list ("" -> nil, meaning "no filter on this dimension").
func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// splitInt64CSV parses a comma-separated list of integer IDs. Any entry that
// is not an integer is an error — silently dropping it would widen the filter
// and return data the caller did not ask for.
func splitInt64CSV(value string) ([]int64, error) {
	parts := splitCSV(value)
	if parts == nil {
		return nil, nil
	}
	out := make([]int64, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%q is not an integer id", p)
		}
		out = append(out, v)
	}
	return out, nil
}
