package export

import (
	"fmt"
	"time"

	"intelligentBI/entity"
	"intelligentBI/pkg"
)

// Repo is implemented by the repository layer. Given a product/system, it
// resolves which table holds that system's data and returns the rows for the
// given time range, already flattened into entity.DataSet — this service
// never needs to know a specific product's table name or column layout.
type Repo interface {
	GetActivityData(product pkg.ProductList, from, to time.Time) (entity.DataSet, error)
}

// exportLookbackMonths is how far back an export goes.
const exportLookbackMonths = 2

// ExportService returns a product's recent activity data for the caller to
// consume directly — it does not persist or write the result anywhere.
type ExportService struct {
	Repository Repo
}

func NewExportService(repo Repo) *ExportService {
	return &ExportService{Repository: repo}
}

type ExportRequest struct {
	Product pkg.ProductList
}

type ExportResponse struct {
	RowCount int              `json:"row_count"`
	Data     []map[string]any `json:"data"`
}

func (e ExportService) Export(request ExportRequest) (response ExportResponse, err error) {
	to := time.Now()
	from := to.AddDate(0, -exportLookbackMonths, 0)

	data, err := e.Repository.GetActivityData(request.Product, from, to)
	if err != nil {
		return response, fmt.Errorf("fetch activity data: %w", err)
	}

	rows := make([]map[string]any, 0, len(data.Rows))
	for _, row := range data.Rows {
		record := make(map[string]any, len(data.Columns))
		for i, column := range data.Columns {
			if i < len(row) {
				record[column] = row[i]
			}
		}
		rows = append(rows, record)
	}

	return ExportResponse{
		RowCount: len(rows),
		Data:     rows,
	}, nil
}
