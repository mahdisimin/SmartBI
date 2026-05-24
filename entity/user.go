package entity

import "database/sql"

type User struct {
	ID          int64
	UserName    string `db:"UserName"`
	PhoneNumber string `db:"PhoneNumber"`
	Password    string `db:"Password"`
	Avatar      string
	WebAppList  []WebAppList
}

type WebAppList struct {
	WebAppName string
	WebAppURL  string
}

type WebAppListDB struct {
	WebAppName sql.NullString `db:"WebAppName"`
	WebAppURL  sql.NullString `db:"WebAppURL"`
}
