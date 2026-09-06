package echowebframework

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"intelligentBI/pkg"
	synopsrepo "intelligentBI/repository/SQLServer/synops"
	"intelligentBI/service/export"

	"github.com/labstack/echo/v5"
)

func ExportHandler(c *echo.Context) error {
	productName := c.Param("product")
	if productName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "product is required")
	}

	product, err := pkg.ParseProductList(productName)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	filters := export.Filters{
		Days:    splitCSV(c.QueryParam("days")),
		Modules: splitCSV(c.QueryParam("modules")),
		Methods: splitCSV(c.QueryParam("methods")),
		UserIDs: splitInt64CSV(c.QueryParam("users")),
	}

	repo := synopsrepo.UserActivity{}
	exportServ := export.NewExportService(repo)

	response, err := exportServ.Export(export.ExportRequest{Product: product, Filters: filters})
	if err != nil {
		log.Printf("error on HTTP request , export error : %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
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

func splitInt64CSV(value string) []int64 {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if v, err := strconv.ParseInt(p, 10, 64); err == nil {
			out = append(out, v)
		}
	}
	return out
}
