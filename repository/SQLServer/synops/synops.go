package SQLServer

import (
	"errors"
	external "intelligentBI/external/synops/ver1"
	"intelligentBI/pkg"

	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/jmoiron/sqlx"
)

type Synops struct {
	DB           *sqlx.DB
	resourseName pkg.SynOpsAPIList
}

type loginHistory struct {
}

func (s Synops) Persistdata(data any) error {
	switch s.resourseName {
	case pkg.LoginHistory:
		return s.LoginHistory(data)
	}
	return nil
}

func (s Synops) LoginHistory(data any) error {
	input, ok := data.(external.LoginHistoryRes)
	if !ok {
		return errors.New("input is not loginHistoryRes")
	}

	tvp := mssql.TVP{
		TypeName: "synops.LoginHistoryType",
		Value:    input.Body.Data,
	}

	if _, err := s.DB.Exec("EXEC synops.AddLoginHistory  @LoginHistory=@p1", tvp); err != nil {
		return err
	}
	return nil
}
