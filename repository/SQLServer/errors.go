package SQLServer

import (
	"errors"

	mssql "github.com/denisenkom/go-mssqldb"
)

// SQL Server duplicate-key errors: 2627 = unique constraint, 2601 = unique index.
const (
	errDuplicateKeyConstraint = 2627
	errDuplicateKeyIndex      = 2601
)

// IsDuplicateKey reports whether err is a SQL Server unique-key violation.
// Inserts keyed on a natural unique key (EventID, Kafka topic/partition/offset,
// phone number) use it to detect a row that already exists.
func IsDuplicateKey(err error) bool {
	var sqlErr mssql.Error
	if !errors.As(err, &sqlErr) {
		return false
	}
	return sqlErr.Number == errDuplicateKeyConstraint || sqlErr.Number == errDuplicateKeyIndex
}

// permanentDataErrors are SQL Server errors raised because the data itself is
// unacceptable (too long, out of range, wrong type, NULL into NOT NULL, ...).
// Retrying the same row can never succeed, unlike connection/timeout errors.
var permanentDataErrors = map[int32]bool{
	220:  true, // arithmetic overflow for data type
	241:  true, // conversion failed: date/time from string
	242:  true, // conversion produced out-of-range datetime
	245:  true, // conversion failed: value to int
	515:  true, // cannot insert NULL into NOT NULL column
	547:  true, // CHECK / FOREIGN KEY constraint violation
	2628: true, // string or binary data would be truncated (SQL 2019+)
	8114: true, // error converting data type
	8115: true, // arithmetic overflow converting expression
	8152: true, // string or binary data would be truncated
	8169: true, // conversion failed: string to uniqueidentifier
}

// PermanentError marks an error that will fail identically on every retry, so
// a caller that retries transient failures (e.g. the Kafka worker) should
// stop and route the input elsewhere instead. Callers detect it through the
// Permanent() method, without importing this package.
type PermanentError struct {
	Err error
}

func (e PermanentError) Error() string   { return e.Err.Error() }
func (e PermanentError) Unwrap() error   { return e.Err }
func (e PermanentError) Permanent() bool { return true }

// ClassifyError wraps err in PermanentError when it is a data error SQL Server
// will raise on every retry; anything else (connection loss, timeouts,
// deadlocks, ...) is returned unchanged and treated as transient.
func ClassifyError(err error) error {
	var sqlErr mssql.Error
	if errors.As(err, &sqlErr) && permanentDataErrors[sqlErr.Number] {
		return PermanentError{Err: err}
	}
	return err
}
