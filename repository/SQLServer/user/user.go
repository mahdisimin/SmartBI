package SQLServer

import (
	"database/sql"
	"intelligentBI/entity"
	"intelligentBI/repository/SQLServer"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/jmoiron/sqlx"
)

type User struct {
}

func (s User) IsPhoneNumberExists(phoneNumber string) (bool, error) {
	var userID int64
	db, err := SQLServer.Connect()
	if err != nil {
		return false, err
	}
	defer db.Close()
	row := db.QueryRow("SELECT id FROM APP.[USER] WHERE PhoneNumber = ?", phoneNumber)
	if err := row.Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s User) PersistUser(_userName, _phonNumber, _Password string) (int64, error) {
	var userID int64
	var db *sqlx.DB
	if dbTemp, err := SQLServer.Connect(); err != nil {
		return 0, err
	} else {
		db = dbTemp
	}
	defer db.Close()

	row := db.QueryRow("INSERT INTO APP.[USER] (UserName,PhoneNumber,Password) values (?,?,?) select ID = convert(bigint, SCOPE_IDENTITY())", _userName, _phonNumber, _Password)
	if err := row.Scan(&userID); err != nil {
		return 0, err
	}
	return userID, nil
}

func (s User) GetPasswordByPhoneNumber(phoneNumber string) (string, error) {
	var password string
	db, err := SQLServer.Connect()
	if err != nil {
		return "", err
	}
	defer db.Close()
	row := db.QueryRow("SELECT Password FROM APP.[USER] WHERE PhoneNumber = ?", phoneNumber)
	if err := row.Scan(&password); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return password, nil
}

func (s User) GetUserIDByPhoneNumber(phoneNumber string) (int64, error) {
	var userId int64
	db, err := SQLServer.Connect()
	if err != nil {
		return 0, err
	}
	defer db.Close()
	row := db.QueryRow("SELECT id FROM APP.[USER] WHERE PhoneNumber = ?", phoneNumber)
	if err := row.Scan(&userId); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return userId, nil
}

func (s User) GetUserByUserID(userID int64) (entity.User, error) {
	var user entity.User
	var userLinkList []entity.WebAppList
	db, err := SQLServer.Connect()
	if err != nil {
		return user, err
	}
	defer db.Close()

	errTemp := db.Get(&user, "SELECT UserName , PhoneNumber , Password FROM APP.[USER] WHERE ID = ?", userID)
	if errTemp != nil {
		return user, errTemp
	}

	userLinkListTemp, errTemp := s.GetUserLinkListByUserID(userID)
	if errTemp != nil {
		return user, errTemp
	} else {
		userLinkList = append(userLinkList, userLinkListTemp...)
	}

	user.WebAppList = userLinkList

	return user, nil
}

func (s User) GetUserLinkListByUserID(userID int64) ([]entity.WebAppList, error) {
	var WebAppsTempDB []entity.WebAppListDB
	var WebApps []entity.WebAppList

	db, err := SQLServer.Connect()
	if err != nil {
		return WebApps, err
	}
	defer db.Close()

	errTemp := db.Select(&WebAppsTempDB, "SELECT WebAppName ,WebAppLink AS WebAppURL  FROM APP.GetUserListByID WHERE UserID = ?", userID)
	if errTemp != nil {
		return WebApps, errTemp
	}

	for _, WebApp := range WebAppsTempDB {
		WebAppName := ""
		if WebApp.WebAppName.Valid {
			WebAppName = WebApp.WebAppName.String
		}
		WebAppURL := ""
		if WebApp.WebAppURL.Valid {
			WebAppURL = WebApp.WebAppURL.String
		}
		WebApps = append(WebApps, entity.WebAppList{WebAppName, WebAppURL})
	}

	return WebApps, nil
}
