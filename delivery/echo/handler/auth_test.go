package echowebframework

// Session auth, end to end through the real routes (RegisterRoutes, so the
// RequireSession wiring itself is under test) with in-memory repositories.

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"intelligentBI/entity"
	"intelligentBI/pkg"
	"intelligentBI/service"
	"intelligentBI/service/export"

	"github.com/labstack/echo/v5"
)

type authUser struct {
	entity.User
	dashboards []string
}

// authUserRepo implements service.Repo and export.AccessRepo.
type authUserRepo struct{ users []authUser }

func (r *authUserRepo) byPhone(p string) *authUser {
	for i := range r.users {
		if r.users[i].PhoneNumber == p {
			return &r.users[i]
		}
	}
	return nil
}
func (r *authUserRepo) IsPhoneNumberExists(p string) (bool, error) { return r.byPhone(p) != nil, nil }
func (r *authUserRepo) PersistUser(n, p, pw string) (int64, error) {
	id := int64(len(r.users) + 1)
	r.users = append(r.users, authUser{User: entity.User{ID: id, UserName: n, PhoneNumber: p, Password: pw}})
	return id, nil
}
func (r *authUserRepo) GetPasswordByPhoneNumber(p string) (string, error) {
	if u := r.byPhone(p); u != nil {
		return u.Password, nil
	}
	return "", nil
}
func (r *authUserRepo) GetUserIDByPhoneNumber(p string) (int64, error) { return r.byPhone(p).ID, nil }
func (r *authUserRepo) GetUserByUserID(id int64) (entity.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u.User, nil
		}
	}
	return entity.User{}, sql.ErrNoRows
}
func (r *authUserRepo) HasDashboardAccess(userID int64, link string) (bool, error) {
	for _, u := range r.users {
		if u.ID == userID {
			for _, d := range u.dashboards {
				if d == link {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

type authSessions map[string]int64

func (m authSessions) CreateSession(h []byte, userID int64, _ time.Time) error {
	m[string(h)] = userID
	return nil
}
func (m authSessions) GetSessionUserID(h []byte, _ time.Time) (int64, error) {
	if id, ok := m[string(h)]; ok {
		return id, nil
	}
	return 0, entity.ErrSessionNotFound
}
func (m authSessions) DeleteSession(h []byte) error          { delete(m, string(h)); return nil }
func (m authSessions) DeleteExpiredSessions(time.Time) error { return nil }

type authExportRepo struct{}

func (authExportRepo) GetActivityEvents(pkg.ProductList, time.Time, time.Time) ([]entity.ActivityEvent, error) {
	return nil, nil
}

// newAuthServer has two users: alice (id 1, granted Synops) and bob (id 2,
// no dashboards). Both use password "secret".
func newAuthServer(t *testing.T) *echo.Echo {
	t.Helper()
	hash, err := pkg.HashPassword("secret")
	if err != nil {
		t.Fatal(err)
	}
	users := &authUserRepo{users: []authUser{
		{User: entity.User{ID: 1, UserName: "alice", PhoneNumber: "09120000001", Password: hash}, dashboards: []string{"/dashboards/synops"}},
		{User: entity.User{ID: 2, UserName: "bob", PhoneNumber: "09120000002", Password: hash}},
	}}
	exportService := export.NewExportService(authExportRepo{})
	exportService.Access = users

	h := &Handler{
		UserService:    service.UserService{Repository: users},
		ExportService:  exportService,
		SessionService: service.SessionService{Repository: authSessions{}, TTL: time.Hour},
	}
	e := echo.New()
	h.RegisterRoutes(e)
	return e
}

func authDo(e *echo.Echo, method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func sessionCookie(rec *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == SessionCookieName {
			return c
		}
	}
	return nil
}

func login(t *testing.T, e *echo.Echo, phone string) *http.Cookie {
	t.Helper()
	rec := authDo(e, http.MethodPost, "/user/login", `{"phone_number":"`+phone+`","password":"secret"}`, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("login %s: %d %s", phone, rec.Code, rec.Body)
	}
	c := sessionCookie(rec)
	if c == nil {
		t.Fatal("login did not set the session cookie")
	}
	return c
}

func TestAuth_LoginSetsSecureSessionCookie(t *testing.T) {
	e := newAuthServer(t)

	rec := authDo(e, http.MethodPost, "/user/login", `{"phone_number":"09120000001","password":"wrong"}`, nil)
	if rec.Code != http.StatusUnauthorized || sessionCookie(rec) != nil {
		t.Fatalf("wrong password: want 401 and no cookie, got %d cookie=%v", rec.Code, sessionCookie(rec))
	}

	c := login(t, e, "09120000001")
	if !c.HttpOnly || c.SameSite != http.SameSiteLaxMode || c.Path != "/" || c.MaxAge <= 0 || c.Value == "" {
		t.Errorf("unexpected cookie attributes: %+v", c)
	}
}

func TestAuth_ProfileRequiresOwnSession(t *testing.T) {
	e := newAuthServer(t)
	alice := login(t, e, "09120000001")

	for _, tc := range []struct {
		name   string
		path   string
		cookie *http.Cookie
		want   int
	}{
		{"me, no session", "/user/user_profile/me", nil, http.StatusUnauthorized},
		{"me, forged session", "/user/user_profile/me", &http.Cookie{Name: SessionCookieName, Value: "forged"}, http.StatusUnauthorized},
		{"me, logged in", "/user/user_profile/me", alice, http.StatusOK},
		{"own id", "/user/user_profile/1", alice, http.StatusOK},
		{"other user's id (IDOR)", "/user/user_profile/2", alice, http.StatusForbidden},
		{"id, no session", "/user/user_profile/1", nil, http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := authDo(e, http.MethodGet, tc.path, "", tc.cookie)
			if rec.Code != tc.want {
				t.Fatalf("want %d, got %d %s", tc.want, rec.Code, rec.Body)
			}
			if rec.Code == http.StatusOK && !strings.Contains(rec.Body.String(), `"user_name":"alice"`) {
				t.Errorf("expected alice's own profile, got %s", rec.Body)
			}
			if rec.Code != http.StatusOK && strings.Contains(rec.Body.String(), "0912") {
				t.Errorf("phone number leaked in error response: %s", rec.Body)
			}
		})
	}
}

func TestAuth_ExportRequiresSessionAndProductGrant(t *testing.T) {
	e := newAuthServer(t)
	alice := login(t, e, "09120000001")
	bob := login(t, e, "09120000002")

	if rec := authDo(e, http.MethodGet, "/export/synops", "", nil); rec.Code != http.StatusUnauthorized {
		t.Errorf("no session: want 401, got %d", rec.Code)
	}
	if rec := authDo(e, http.MethodGet, "/export/synops", "", bob); rec.Code != http.StatusForbidden {
		t.Errorf("logged in without the Synops grant: want 403, got %d %s", rec.Code, rec.Body)
	}
	if rec := authDo(e, http.MethodGet, "/export/synops", "", alice); rec.Code != http.StatusOK {
		t.Errorf("granted user: want 200, got %d %s", rec.Code, rec.Body)
	}
}

func TestAuth_LogoutRevokesSession(t *testing.T) {
	e := newAuthServer(t)
	alice := login(t, e, "09120000001")

	rec := authDo(e, http.MethodPost, "/user/logout", "", alice)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout: want 204, got %d", rec.Code)
	}
	if c := sessionCookie(rec); c == nil || c.MaxAge >= 0 || c.Value != "" {
		t.Errorf("logout must clear the cookie, got %+v", c)
	}
	// The old token is dead server-side, even if a client kept it.
	if rec := authDo(e, http.MethodGet, "/user/user_profile/me", "", alice); rec.Code != http.StatusUnauthorized {
		t.Errorf("after logout: want 401, got %d", rec.Code)
	}
	// Logging out again, or without a session, is harmless.
	if rec := authDo(e, http.MethodPost, "/user/logout", "", nil); rec.Code != http.StatusNoContent {
		t.Errorf("logout without session: want 204, got %d", rec.Code)
	}
}
