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
		UserName:    "TestUser",
		PhoneNumber: "09201008700",
		Password:    "123456",
	}
	_, err := userServ.Register(Req)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSqlServer_Login(t *testing.T) {
	userServ := &service.UserService{
		Repository: SqlServer{},
	}
	req := service.UserLoginRequest{
		PhoneNumber: "09201008700",
		Password:    "123456",
	}
	_, err := userServ.Login(req)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSqlServer_GetUserByUserID(t *testing.T) {
	s := SqlServer{}
	user, err := s.GetUserByUserID(2)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", user)
}

func TestSqlServer_GetUserLinkListByUserID(t *testing.T) {
	s := SqlServer{}
	Links, err := s.GetUserLinkListByUserID(2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(Links)
}
