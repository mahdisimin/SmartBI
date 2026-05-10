package SQLServer

import (
	"intelligentBI/service"
	"testing"
)

func TestSqlServer_PersistUser(t *testing.T) {
	userServ := &service.UserService{
		Repository: SqlServer{},
	}
	Req := service.UserRegisterRequest{
		UserName:    "Mahdi",
		PhoneNumber: "09201008697",
		Password:    "123456",
	}
	_, err := userServ.Register(Req)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSqlServer_Connect(t *testing.T) {
	db, err := connect()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
}

//func TestSqlServer_IsPhoneNumberExists(t *testing.T) {
//
//}
