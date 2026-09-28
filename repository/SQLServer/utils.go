package SQLServer

import (
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/jmoiron/sqlx"

	"intelligentBI/pkg"
	"net/url"
)

// Connection pool limits. A single *sqlx.DB is opened once per process and
// shared by every repository, so connections are reused instead of opening a
// new TCP connection per query (which exhausts ephemeral ports under load).
const (
	maxOpenConns    = 10
	maxIdleConns    = 10
	connMaxLifetime = 30 * time.Minute
	connMaxIdleTime = 5 * time.Minute
)

// NewDB opens the shared SQL Server connection pool and verifies it with a
// ping. Connection settings come from the environment (see pkg.LoadDBConfig).
// Call it once at process start and pass the result to repositories; the
// caller owns closing it.
func NewDB() (*sqlx.DB, error) {
	cfg, err := pkg.LoadDBConfig()
	if err != nil {
		return nil, err
	}
	connURL := url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		RawQuery: url.Values{"database": {cfg.Database}}.Encode(),
	}
	db, err := sqlx.Open("sqlserver", connURL.String())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
