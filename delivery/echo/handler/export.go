package echowebframework

import (
	"log"
	"net/http"

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

	repo := synopsrepo.UserActivity{}
	exportServ := export.NewExportService(repo)

	response, err := exportServ.Export(export.ExportRequest{Product: product})
	if err != nil {
		log.Printf("error on HTTP request , export error : %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, response)
}
