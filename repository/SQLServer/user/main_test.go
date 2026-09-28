package SQLServer

import (
	"log"
	"os"
	"testing"

	"intelligentBI/repository/SQLServer"

	"github.com/jmoiron/sqlx"
)

// testDB is the shared connection pool for this package's integration tests.
// It is nil when SQL Server is unreachable; tests call requireDB to skip.
var testDB *sqlx.DB

func TestMain(m *testing.M) {
	db, err := SQLServer.NewDB()
	if err != nil {
		log.Printf("SQL Server unavailable, DB integration tests will be skipped: %v", err)
	} else {
		testDB = db
	}

	code := m.Run()
	if testDB != nil {
		testDB.Close()
	}
	os.Exit(code)
}

func requireDB(t *testing.T) *sqlx.DB {
	t.Helper()
	if testDB == nil {
		t.Skip("SQL Server unavailable")
	}
	return testDB
}
