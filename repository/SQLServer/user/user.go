package SQLServer

import (
	"database/sql"
	"fmt"
	"intelligentBI/entity"
	"intelligentBI/repository/SQLServer"

	"github.com/jmoiron/sqlx"
)

type User struct {
	DB *sqlx.DB
}

func NewUser(db *sqlx.DB) User {
	return User{DB: db}
}

func (s User) IsPhoneNumberExists(phoneNumber string) (bool, error) {
	var userID int64
	row := s.DB.QueryRow("SELECT id FROM APP.[USER] WHERE PhoneNumber = @p1", phoneNumber)
	if err := row.Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// PersistUser inserts a user and returns its new ID. A phone number that is
// already registered (UX_USER_PhoneNumber, including a concurrent insert that
// won the race) returns entity.ErrPhoneNumberExists.
//
// The ID comes from OUTPUT in the same statement: with a separate
// SCOPE_IDENTITY() SELECT in the batch, a failed INSERT still lets the SELECT
// run and return NULL, masking the real error behind a confusing Scan error.
func (s User) PersistUser(_userName, _phonNumber, _Password string) (int64, error) {
	var userID int64
	row := s.DB.QueryRow("INSERT INTO APP.[USER] (UserName,PhoneNumber,Password) OUTPUT INSERTED.ID values (@p1,@p2,@p3)", _userName, _phonNumber, _Password)
	if err := row.Scan(&userID); err != nil {
		if SQLServer.IsDuplicateKey(err) {
			return 0, entity.ErrPhoneNumberExists
		}
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return userID, nil
}

func (s User) GetPasswordByPhoneNumber(phoneNumber string) (string, error) {
	var password string
	row := s.DB.QueryRow("SELECT Password FROM APP.[USER] WHERE PhoneNumber = @p1", phoneNumber)
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
	row := s.DB.QueryRow("SELECT id FROM APP.[USER] WHERE PhoneNumber = @p1", phoneNumber)
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

	// A missing user returns sql.ErrNoRows (wrapped), so callers can tell
	// "not found" apart from a database failure with errors.Is.
	errTemp := s.DB.Get(&user, "SELECT UserName , PhoneNumber , Password FROM APP.[USER] WHERE ID = @p1", userID)
	if errTemp != nil {
		return user, fmt.Errorf("get user %d: %w", userID, errTemp)
	}

	userLinkListTemp, errTemp := s.GetUserLinkListByUserID(userID)
	if errTemp != nil {
		return user, errTemp
	} else {
		userLinkList = userLinkListTemp
	}

	user.WebAppList = userLinkList

	return user, nil
}

// GetUserLinkListByUserID returns the web apps a user can access — an empty
// (non-nil) list when there are none.
func (s User) GetUserLinkListByUserID(userID int64) ([]entity.WebAppList, error) {
	var WebAppsTempDB []entity.WebAppListDB
	WebApps := []entity.WebAppList{}

	errTemp := s.DB.Select(&WebAppsTempDB, "SELECT WebAppName ,WebAppLink AS WebAppURL  FROM APP.GetUserListByID WHERE UserID = @p1", userID)
	if errTemp != nil {
		return WebApps, errTemp
	}

	for _, WebApp := range WebAppsTempDB {
		// APP.GetUserListByID LEFT JOINs the links, so a user without any
		// yields one row with NULL name and link — that is "no links", not a
		// link.
		if !WebApp.WebAppName.Valid && !WebApp.WebAppURL.Valid {
			continue
		}
		WebAppName := ""
		if WebApp.WebAppName.Valid {
			WebAppName = WebApp.WebAppName.String
		}
		WebAppURL := ""
		if WebApp.WebAppURL.Valid {
			WebAppURL = WebApp.WebAppURL.String
		}
		WebApps = append(WebApps, entity.WebAppList{WebAppName: WebAppName, WebAppURL: WebAppURL})
	}

	return WebApps, nil
}

// HasDashboardAccess reports whether the user has been granted the web-app
// link (APP.User_WebAppLink -> APP.WebAppLinkList.Link) of a dashboard, e.g.
// "/dashboards/synops".
func (s User) HasDashboardAccess(userID int64, link string) (bool, error) {
	var granted bool
	err := s.DB.QueryRow(`SELECT CASE WHEN EXISTS (
			SELECT 1 FROM APP.User_WebAppLink uw
			JOIN APP.WebAppLinkList w ON w.ID = uw.LinkID
			WHERE uw.UserID = @p1 AND w.Link = @p2
		) THEN CAST(1 AS BIT) ELSE CAST(0 AS BIT) END`, userID, link).Scan(&granted)
	if err != nil {
		return false, fmt.Errorf("check dashboard access: %w", err)
	}
	return granted, nil
}
