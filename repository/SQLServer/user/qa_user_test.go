package SQLServer

// QA: user repository + service against the real SMARTBI database.

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"intelligentBI/service"
)

func qaPhone() string { return fmt.Sprintf("qa%d", time.Now().UnixNano()%1e12) }

func qaCleanup(t *testing.T, phone string) {
	db := requireDB(t)
	t.Cleanup(func() { db.Exec("DELETE FROM APP.[USER] WHERE PhoneNumber = @p1", phone) })
}

// Two simultaneous registrations with the same phone number: the service
// checks-then-inserts with no unique constraint, so both can succeed.
func TestQA_Register_ConcurrentSamePhone_CreatesDuplicates(t *testing.T) {
	db := requireDB(t)
	phone := qaPhone()
	qaCleanup(t, phone)
	svc := service.UserService{Repository: NewUser(db)}

	const n = 8
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			svc.Register(service.UserRegisterRequest{UserName: "qa", PhoneNumber: phone, Password: "pw"})
		}()
	}
	close(start)
	wg.Wait()

	var rows int
	db.QueryRow("SELECT COUNT(*) FROM APP.[USER] WHERE PhoneNumber = @p1", phone).Scan(&rows)
	if rows != 1 {
		t.Errorf("race: %d concurrent registers with one phone created %d accounts (want 1)", n, rows)
	}
}

// BUG-09: a failed INSERT must surface its real SQL error, not a Scan error
// from a trailing SCOPE_IDENTITY() SELECT. Called on the repository directly,
// bypassing service validation.
func TestQA_PersistUser_FailedInsertReturnsRealError(t *testing.T) {
	db := requireDB(t)
	phone := qaPhone()
	qaCleanup(t, phone)
	_, err := NewUser(db).PersistUser(strings.Repeat("n", 501), phone, "pw")
	if err == nil {
		t.Fatal("501-char name accepted by nvarchar(500)?")
	}
	if strings.Contains(err.Error(), "Scan error") || !strings.Contains(strings.ToLower(err.Error()), "truncated") {
		t.Errorf("real insert error masked: %v", err)
	}
}

func TestQA_Register_OversizedName_ErrorLeaksDB(t *testing.T) {
	db := requireDB(t)
	phone := qaPhone()
	qaCleanup(t, phone)
	svc := service.UserService{Repository: NewUser(db)}
	_, err := svc.Register(service.UserRegisterRequest{UserName: strings.Repeat("n", 501), PhoneNumber: phone, Password: "pw"})
	if err == nil {
		t.Fatal("501-char name accepted by nvarchar(500)?")
	}
	t.Logf("error returned to client: %v", err)
	if strings.Contains(err.Error(), "mssql") || strings.Contains(err.Error(), "truncated") {
		t.Errorf("no input length validation; raw DB error returned to client: %v", err)
	}
}

func TestQA_Profile_UserWithoutLinks(t *testing.T) {
	db := requireDB(t)
	phone := qaPhone()
	qaCleanup(t, phone)
	svc := service.UserService{Repository: NewUser(db)}
	reg, err := svc.Register(service.UserRegisterRequest{UserName: "qa", PhoneNumber: phone, Password: "pw"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := svc.Profile(service.UserProfileRequest{UserID: reg.UserId})
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range p.UserLinkList {
		if l.WebAppName == "" && l.WebAppURL == "" {
			t.Errorf("profile of user with no links returns placeholder link %+v (LEFT JOIN null row) — want empty list", p.UserLinkList)
			break
		}
	}
}

func TestQA_Login_RoundTrip(t *testing.T) {
	db := requireDB(t)
	phone := qaPhone()
	qaCleanup(t, phone)
	svc := service.UserService{Repository: NewUser(db)}
	reg, err := svc.Register(service.UserRegisterRequest{UserName: "کاربر تست", PhoneNumber: phone, Password: "p@ss' OR 1=1--"})
	if err != nil {
		t.Fatal(err)
	}
	login, err := svc.Login(service.UserLoginRequest{PhoneNumber: phone, Password: "p@ss' OR 1=1--"})
	if err != nil || login.UserId != reg.UserId {
		t.Fatalf("login round trip: %+v %v", login, err)
	}
	if _, err := svc.Login(service.UserLoginRequest{PhoneNumber: phone, Password: "' OR '1'='1"}); err == nil {
		t.Fatal("SQL-injection style password logged in")
	}
	p, _ := svc.Profile(service.UserProfileRequest{UserID: reg.UserId})
	if p.UserName != "کاربر تست" {
		t.Errorf("unicode name not preserved: %q", p.UserName)
	}
}
