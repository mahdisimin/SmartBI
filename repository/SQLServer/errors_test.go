package SQLServer

import (
	"errors"
	"fmt"
	"testing"

	mssql "github.com/denisenkom/go-mssqldb"
)

func TestClassifyError(t *testing.T) {
	for _, tc := range []struct {
		name      string
		err       error
		permanent bool
	}{
		{"truncation", mssql.Error{Number: 8152}, true},
		{"truncation 2019+", mssql.Error{Number: 2628}, true},
		{"bad uniqueidentifier", mssql.Error{Number: 8169}, true},
		{"smallint overflow", mssql.Error{Number: 220}, true},
		{"wrapped data error", fmt.Errorf("insert: %w", mssql.Error{Number: 8115}), true},
		{"deadlock victim", mssql.Error{Number: 1205}, false},
		{"connection error", errors.New("dial tcp [::1]:1433: connectex: refused"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var p interface{ Permanent() bool }
			got := errors.As(ClassifyError(tc.err), &p) && p.Permanent()
			if got != tc.permanent {
				t.Errorf("permanent = %v, want %v", got, tc.permanent)
			}
		})
	}
}

func TestIsDuplicateKey(t *testing.T) {
	if !IsDuplicateKey(mssql.Error{Number: 2627}) || !IsDuplicateKey(fmt.Errorf("x: %w", mssql.Error{Number: 2601})) {
		t.Error("2627/2601 must be duplicate keys")
	}
	if IsDuplicateKey(mssql.Error{Number: 8152}) || IsDuplicateKey(errors.New("2627")) {
		t.Error("only SQL Server errors 2627/2601 are duplicate keys")
	}
}
