package SQLServer

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/jmoiron/sqlx"
)

const (
	userName = "m.simin"
	password = "123456"
	uRL      = "10.1.52.13:1433"
	database = "BRS_PreUpdate"
)

type SqlServer struct {
}

func connect() (*sqlx.DB, error) {
	var sqlDB *sqlx.DB
	pass := url.QueryEscape(password)
	connString := fmt.Sprintf("sqlserver://%s:%s@%s?database=%s", userName, pass, uRL, database)
	if dbtemp, err := sqlx.Open("mssql", connString); err != nil {
		return nil, err
	} else {
		sqlDB = dbtemp
	}
	return sqlDB, nil
}

func (s SqlServer) IsPhoneNumberExists(phoneNumber string) (bool, error) {
	var userID int64
	db, err := connect()
	if err != nil {
		return false, err
	}
	defer db.Close()
	row := db.QueryRow("SELECT * FROM APP.[USER] WHERE PhoneNumber = ?", phoneNumber)
	if err := row.Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s SqlServer) PersistUser(_userName, _phonNumber, _Password string) (int64, error) {
	var userID int64
	var result sql.Result
	var db *sqlx.DB
	if dbTemp, err := connect(); err != nil {
		return 0, err
	} else {
		db = dbTemp
	}
	defer db.Close()

	if resultTemp, err := db.Exec("INSERT INTO APP.[USER] (UserName,PhoneNumber,Password) values (?,?,?)", _userName, _phonNumber, _Password); err != nil {
		return 0, err
	} else {
		result = resultTemp
	}
	if userIDTemp, err := result.LastInsertId(); err != nil {
		return 0, err
	} else {
		userID = userIDTemp
	}
	return userID, nil
}
