package SQLServer

import (
	"errors"
	external "intelligentBI/external/synops"
	"intelligentBI/pkg"
	"intelligentBI/repository/SQLServer"

	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/jmoiron/sqlx"
)

type Synops struct {
	resourseName pkg.SynOpsAPIList
}

type loginHistory struct {
}

func (s Synops) Persistdata(data any) error {
	switch s.resourseName {
	case pkg.LoginHistory:
		return s.LoginHostory(data)
	}
	return nil
}

func (s Synops) LoginHostory(data any) error {
	var db *sqlx.DB
	input, ok := data.(external.LoginHistoryRes)
	if dbTemp, err := SQLServer.Connect(); err != nil {
		return err
	} else {
		db = dbTemp
	}
	defer db.Close()

	if !ok {
		return errors.New("input is not loginHistoryRes")
	}

	tvp := mssql.TVP{
		TypeName: "synops.LoginHistoryType",
		Value:    input.Body.Data,
	}

	if _, err := db.Exec("EXEC synops.AddLoginHistory  @LoginHistory=@p1", tvp); err != nil {
		return err
	}
	return nil
}
