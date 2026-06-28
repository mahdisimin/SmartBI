package SQLServer

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"intelligentBI/pkg"
	"net/url"
)

func Connect() (*sqlx.DB, error) {
	var sqlDB *sqlx.DB
	pass := url.QueryEscape(pkg.Password)
	connString := fmt.Sprintf("sqlserver://%s:%s@%s?database=%s", pkg.UserName, pass, pkg.URL, pkg.Database)
	if dbtemp, err := sqlx.Open("sqlserver", connString); err != nil {
		return nil, err
	} else {
		sqlDB = dbtemp
	}
	return sqlDB, nil
}
