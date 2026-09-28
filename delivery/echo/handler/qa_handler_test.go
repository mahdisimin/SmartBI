package echowebframework

// QA: HTTP-level tests of the Echo handlers with in-memory fakes behind the
// real service layer. Each test asserts the behaviour an API client (the
// dashboard) relies on: status codes, error bodies, response shape.

import (
	"database/sql"
	"encoding/json"
	"errors"
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

type qaUserRepo struct {
	users  map[string]entity.User // by phone
	nextID int64
}

func (r *qaUserRepo) IsPhoneNumberExists(p string) (bool, error) { _, ok := r.users[p]; return ok, nil }
func (r *qaUserRepo) PersistUser(n, p, pw string) (int64, error) {
	r.nextID++
	r.users[p] = entity.User{ID: r.nextID, UserName: n, PhoneNumber: p, Password: pw}
	return r.nextID, nil
}
func (r *qaUserRepo) GetPasswordByPhoneNumber(p string) (string, error) {
	return r.users[p].Password, nil
}
func (r *qaUserRepo) GetUserIDByPhoneNumber(p string) (int64, error) { return r.users[p].ID, nil }
func (r *qaUserRepo) GetUserByUserID(id int64) (entity.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return entity.User{}, sql.ErrNoRows // what sqlx.Get returns for a missing row
}

// HasDashboardAccess implements export.AccessRepo: user 1 (alice) is granted
// the Synops dashboard, nobody else is.
func (r *qaUserRepo) HasDashboardAccess(userID int64, link string) (bool, error) {
	return userID == 1 && link == "/dashboards/synops", nil
}

type qaExportRepo struct {
	events []entity.ActivityEvent
	err    error
}

func (r qaExportRepo) GetActivityEvents(pkg.ProductList, time.Time, time.Time) ([]entity.ActivityEvent, error) {
	return r.events, r.err
}

// qaServer mounts the real routes (RegisterRoutes, so RequireSession is
// applied exactly as in production) over in-memory fakes.
func qaServer(userRepo *qaUserRepo, exp qaExportRepo) *echo.Echo {
	exportService := export.NewExportService(exp)
	exportService.Access = userRepo
	h := &Handler{
		UserService:    service.UserService{Repository: userRepo},
		ExportService:  exportService,
		SessionService: service.SessionService{Repository: authSessions{}, TTL: time.Hour},
	}
	e := echo.New()
	h.RegisterRoutes(e)
	return e
}

func qaDo(e *echo.Echo, method, path, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func newQARepo() *qaUserRepo {
	r := &qaUserRepo{users: map[string]entity.User{}}
	hash, err := pkg.HashPassword("secret")
	if err != nil {
		panic(err)
	}
	r.PersistUser("alice", "09120000001", hash)
	return r
}

// qaSession logs alice (user 1, granted Synops) in and returns her session cookie.
func qaSession(t *testing.T, e *echo.Echo) *http.Cookie {
	t.Helper()
	return login(t, e, "09120000001")
}

func TestQA_Profile_InvalidID_Returns400(t *testing.T) {
	e := qaServer(newQARepo(), qaExportRepo{})
	rec := qaDo(e, http.MethodGet, "/user/user_profile/abc", "", qaSession(t, e))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("GET /user/user_profile/abc: want 400, got %d body=%s", rec.Code, rec.Body)
	}
}

func TestQA_Profile_OtherUserID_Forbidden(t *testing.T) {
	e := qaServer(newQARepo(), qaExportRepo{})
	rec := qaDo(e, http.MethodGet, "/user/user_profile/999", "", qaSession(t, e))
	if rec.Code != http.StatusForbidden {
		t.Errorf("GET another user's profile with a valid session: want 403, got %d body=%s", rec.Code, rec.Body)
	}
}

func TestQA_Profile_OwnID_OK(t *testing.T) {
	e := qaServer(newQARepo(), qaExportRepo{})
	rec := qaDo(e, http.MethodGet, "/user/user_profile/1", "", qaSession(t, e))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"user_name":"alice"`) {
		t.Errorf("own profile: %d %s", rec.Code, rec.Body)
	}
}

func TestQA_ProtectedRoutes_RequireSession(t *testing.T) {
	e := qaServer(newQARepo(), qaExportRepo{})
	forged := &http.Cookie{Name: SessionCookieName, Value: "forged-token"}
	for _, path := range []string{"/user/user_profile/1", "/user/user_profile/me", "/export/synops"} {
		if rec := qaDo(e, http.MethodGet, path, ""); rec.Code != http.StatusUnauthorized {
			t.Errorf("GET %s without session: want 401, got %d", path, rec.Code)
		}
		if rec := qaDo(e, http.MethodGet, path, "", forged); rec.Code != http.StatusUnauthorized {
			t.Errorf("GET %s with forged cookie: want 401, got %d", path, rec.Code)
		}
	}
}

func TestQA_Profile_NoAuthRequired(t *testing.T) {
	// Any caller can read any user's profile (name + phone) by id — no token,
	// no session. Documented as a security finding.
	rec := qaDo(qaServer(newQARepo(), qaExportRepo{}), http.MethodGet, "/user/user_profile/1", "")
	if rec.Code == http.StatusOK && strings.Contains(rec.Body.String(), "09120000001") {
		t.Errorf("IDOR: unauthenticated request got another user's phone: %s", rec.Body)
	}
}

func TestQA_Register_RejectsBlankFields(t *testing.T) {
	for _, body := range []string{
		`{}`,
		`{"user_name":"","phone_number":"","password":""}`,
		`{"user_name":"bob","phone_number":"09120000002","password":""}`,
		`{"user_name":"bob","phone_number":"   ","password":"x"}`,
	} {
		rec := qaDo(qaServer(newQARepo(), qaExportRepo{}), http.MethodPost, "/user/register", body)
		if rec.Code == http.StatusOK {
			t.Errorf("register %s: want 400, got 200 (account created with blank fields) body=%s", body, rec.Body)
		}
	}
}

func TestQA_Register_MalformedJSON_Returns400(t *testing.T) {
	rec := qaDo(qaServer(newQARepo(), qaExportRepo{}), http.MethodPost, "/user/register", `{"user_name":`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

func TestQA_Register_ResponseFieldCasing(t *testing.T) {
	rec := qaDo(qaServer(newQARepo(), qaExportRepo{}), http.MethodPost, "/user/register",
		`{"user_name":"bob","phone_number":"09120000002","password":"pw"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("register: %d %s", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"user_id"`) {
		t.Errorf("API contract: register returns %s — PascalCase \"UserId\" while login returns snake_case \"user_id\"", strings.TrimSpace(rec.Body.String()))
	}
}

func TestQA_Login_ErrorsDoNotRevealAccountExistence(t *testing.T) {
	e := qaServer(newQARepo(), qaExportRepo{})
	unknown := qaDo(e, http.MethodPost, "/user/login", `{"phone_number":"09129999999","password":"x"}`)
	wrongPw := qaDo(e, http.MethodPost, "/user/login", `{"phone_number":"09120000001","password":"x"}`)
	if unknown.Body.String() != wrongPw.Body.String() {
		t.Errorf("user enumeration: unknown phone -> %s | wrong password -> %s",
			strings.TrimSpace(unknown.Body.String()), strings.TrimSpace(wrongPw.Body.String()))
	}
	if wrongPw.Code != http.StatusUnauthorized {
		t.Errorf("wrong credentials: want 401, got %d", wrongPw.Code)
	}
}

func TestQA_Login_Success(t *testing.T) {
	rec := qaDo(qaServer(newQARepo(), qaExportRepo{}), http.MethodPost, "/user/login", `{"phone_number":"09120000001","password":"secret"}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"user_id":1`) {
		t.Errorf("login: %d %s", rec.Code, rec.Body)
	}
}

func TestQA_Export_UnknownProduct_Returns400(t *testing.T) {
	e := qaServer(newQARepo(), qaExportRepo{})
	rec := qaDo(e, http.MethodGet, "/export/nope", "", qaSession(t, e))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

func TestQA_Export_RepoError_DoesNotLeakInternals(t *testing.T) {
	e := qaServer(newQARepo(), qaExportRepo{err: errors.New("mssql: Login failed for user 'sa'")})
	rec := qaDo(e, http.MethodGet, "/export/synops", "", qaSession(t, e))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "mssql") {
		t.Errorf("internal DB error leaked to client: %s", strings.TrimSpace(rec.Body.String()))
	}
}

func TestQA_Export_EmptyData_ArraysNotNull(t *testing.T) {
	e := qaServer(newQARepo(), qaExportRepo{})
	rec := qaDo(e, http.MethodGet, "/export/synops", "", qaSession(t, e))
	if rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	var m map[string]any
	json.Unmarshal(rec.Body.Bytes(), &m)
	for _, k := range []string{"daily_trend", "weekly_trend", "monthly_trend", "top_modules", "method_breakdown", "users"} {
		if m[k] == nil {
			t.Errorf("%s is null on empty data — frontend calls .map() on it", k)
		}
	}
}

func TestQA_Export_InvalidUserFilterSilentlyIgnored(t *testing.T) {
	uid := int64(5)
	exp := qaExportRepo{events: []entity.ActivityEvent{{OccurredAt: time.Now(), UserID: &uid, Status: 200, Module: "M", Method: "GET"}}}
	e := qaServer(newQARepo(), exp)
	rec := qaDo(e, http.MethodGet, "/export/synops?users=abc", "", qaSession(t, e))
	if rec.Code == http.StatusOK && strings.Contains(rec.Body.String(), `"total_events":1`) {
		t.Errorf("users=abc: invalid filter silently dropped -> unfiltered data returned with 200 (want 400)")
	}
}

func TestQA_Export_UserWithoutGrant_Forbidden(t *testing.T) {
	repo := newQARepo()
	hash, _ := pkg.HashPassword("secret")
	repo.PersistUser("bob", "09120000002", hash) // id 2, no dashboard grant
	e := qaServer(repo, qaExportRepo{})
	rec := qaDo(e, http.MethodGet, "/export/synops", "", login(t, e, "09120000002"))
	if rec.Code != http.StatusForbidden {
		t.Errorf("export by a logged-in user without the Synops grant: want 403, got %d body=%s", rec.Code, rec.Body)
	}
}
