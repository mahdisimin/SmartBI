package SQLServer

import (
	"errors"
	"fmt"
	"intelligentBI/entity"
	"intelligentBI/service"
	"strings"
	"testing"
	"time"
)

// testPhone returns a phone number unique to this run and deletes the user
// holding it when the test ends, so the tests are re-runnable.
func testPhone(t *testing.T) string {
	t.Helper()
	db := requireDB(t)
	phone := fmt.Sprintf("test%d", time.Now().UnixNano())
	t.Cleanup(func() { db.Exec("DELETE FROM APP.[USER] WHERE PhoneNumber = @p1", phone) })
	return phone
}

func TestSqlServer_PersistUser(t *testing.T) {
	userServ := &service.UserService{
		Repository: NewUser(requireDB(t)),
	}
	Req := service.UserRegisterRequest{
		UserName:    "TestUser",
		PhoneNumber: testPhone(t),
		Password:    "123456",
	}
	_, err := userServ.Register(Req)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSqlServer_Login(t *testing.T) {
	userServ := &service.UserService{
		Repository: NewUser(requireDB(t)),
	}
	phone := testPhone(t)
	if _, err := userServ.Register(service.UserRegisterRequest{UserName: "TestUser", PhoneNumber: phone, Password: "123456"}); err != nil {
		t.Fatal(err)
	}
	req := service.UserLoginRequest{
		PhoneNumber: phone,
		Password:    "123456",
	}
	_, err := userServ.Login(req)
	if err != nil {
		t.Fatal(err)
	}
}

// A second insert with the same phone number hits UX_USER_PhoneNumber and is
// reported as entity.ErrPhoneNumberExists, not a raw driver error.
func TestSqlServer_PersistUser_DuplicatePhone(t *testing.T) {
	repo := NewUser(requireDB(t))
	phone := testPhone(t)
	if _, err := repo.PersistUser("TestUser", phone, "x"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.PersistUser("TestUser", phone, "x"); !errors.Is(err, entity.ErrPhoneNumberExists) {
		t.Fatalf("want entity.ErrPhoneNumberExists, got %v", err)
	}
}

// A failed insert (here: a name longer than NVARCHAR(500), bypassing service
// validation) must surface SQL Server's own error, not a Scan-NULL error from
// a trailing SCOPE_IDENTITY() SELECT.
func TestSqlServer_PersistUser_SurfacesRealError(t *testing.T) {
	repo := NewUser(requireDB(t))
	_, err := repo.PersistUser(strings.Repeat("n", 501), testPhone(t), "x")
	if err == nil {
		t.Fatal("expected an error for a 501-character name")
	}
	if strings.Contains(err.Error(), "converting NULL") {
		t.Fatalf("real error masked: %v", err)
	}
}

// testUserID creates a throwaway user (deleted when the test ends) and
// returns its ID, so tests do not depend on pre-existing rows.
func testUserID(t *testing.T) int64 {
	t.Helper()
	id, err := NewUser(requireDB(t)).PersistUser("TestUser", testPhone(t), "x")
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestSqlServer_GetUserByUserID(t *testing.T) {
	s := NewUser(requireDB(t))
	user, err := s.GetUserByUserID(testUserID(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", user)
}

func TestSqlServer_GetUserLinkListByUserID(t *testing.T) {
	s := NewUser(requireDB(t))
	Links, err := s.GetUserLinkListByUserID(testUserID(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(Links)
}

// HasDashboardAccess follows APP.User_WebAppLink grants. Grants the Synops
// dashboard link to a throwaway user, then removes the grant and the user.
func TestSqlServer_HasDashboardAccess(t *testing.T) {
	db := requireDB(t)
	repo := NewUser(db)
	phone := testPhone(t)

	var linkID int64
	if err := db.QueryRow("SELECT ID FROM APP.WebAppLinkList WHERE Link = '/dashboards/synops'").Scan(&linkID); err != nil {
		t.Skipf("Synops dashboard link not registered (sample_data/synops_dashboard_access.sql): %v", err)
	}
	userID, err := repo.PersistUser("TestUser", phone, "x")
	if err != nil {
		t.Fatal(err)
	}

	if ok, err := repo.HasDashboardAccess(userID, "/dashboards/synops"); err != nil || ok {
		t.Fatalf("before grant: got %v, %v; want false", ok, err)
	}

	if _, err := db.Exec("INSERT INTO APP.User_WebAppLink (UserID, LinkID) VALUES (@p1, @p2)", userID, linkID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM APP.User_WebAppLink WHERE UserID = @p1", userID) })

	if ok, err := repo.HasDashboardAccess(userID, "/dashboards/synops"); err != nil || !ok {
		t.Fatalf("after grant: got %v, %v; want true", ok, err)
	}
	if ok, err := repo.HasDashboardAccess(userID, "/dashboards/other"); err != nil || ok {
		t.Fatalf("other dashboard: got %v, %v; want false", ok, err)
	}
}
