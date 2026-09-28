package echowebframework

// End to end against the real database: real repositories, real session
// store, real routes and middleware, a real HTTP client with a cookie jar.
// Creates two throwaway users (plus a Synops grant for one of them) and
// deletes them afterwards; sessions go with them (ON DELETE CASCADE).
// Skipped when SQL Server is unavailable.

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"intelligentBI/pkg"
	"intelligentBI/repository/SQLServer"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v5"
)

func TestE2E_SessionAuthAgainstDatabase(t *testing.T) {
	db, err := SQLServer.NewDB()
	if err != nil {
		t.Skipf("SQL Server unavailable: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	e := echo.New()
	NewHandler(db, pkg.SessionConfig{TTL: time.Hour}).RegisterRoutes(e)
	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)

	suffix := time.Now().UnixNano()
	alicePhone, bobPhone := fmt.Sprintf("e2eA%d", suffix), fmt.Sprintf("e2eB%d", suffix)
	t.Cleanup(func() {
		db.Exec("DELETE FROM APP.User_WebAppLink WHERE UserID IN (SELECT ID FROM APP.[USER] WHERE PhoneNumber IN (@p1, @p2))", alicePhone, bobPhone)
		db.Exec("DELETE FROM APP.[USER] WHERE PhoneNumber IN (@p1, @p2)", alicePhone, bobPhone)
	})

	alice, bob := newE2EClient(t, srv.URL), newE2EClient(t, srv.URL)
	aliceID := alice.register(alicePhone)
	bobID := bob.register(bobPhone)
	grantSynops(t, db, aliceID)

	// Not logged in yet.
	alice.expect(http.MethodGet, "/user/user_profile/me", "", http.StatusUnauthorized)
	alice.expect(http.MethodGet, "/export/synops", "", http.StatusUnauthorized)

	// Wrong password: 401 and no session.
	alice.expect(http.MethodPost, "/user/login", `{"phone_number":"`+alicePhone+`","password":"wrong"}`, http.StatusUnauthorized)
	alice.expect(http.MethodGet, "/user/user_profile/me", "", http.StatusUnauthorized)

	alice.login(alicePhone)
	bob.login(bobPhone)

	if body := alice.expect(http.MethodGet, "/user/user_profile/me", "", http.StatusOK); !strings.Contains(body, alicePhone) {
		t.Errorf("/me returned someone else's profile: %s", body)
	}
	alice.expect(http.MethodGet, fmt.Sprintf("/user/user_profile/%d", aliceID), "", http.StatusOK)
	alice.expect(http.MethodGet, fmt.Sprintf("/user/user_profile/%d", bobID), "", http.StatusForbidden)

	// Alice is granted Synops; Bob is logged in but not granted.
	if body := alice.expect(http.MethodGet, "/export/synops", "", http.StatusOK); !strings.Contains(body, `"kpis"`) {
		t.Errorf("export body is not dashboard data: %.200s", body)
	}
	bob.expect(http.MethodGet, "/export/synops", "", http.StatusForbidden)

	// Logout ends the session server-side.
	alice.expect(http.MethodPost, "/user/logout", "", http.StatusNoContent)
	alice.expect(http.MethodGet, "/user/user_profile/me", "", http.StatusUnauthorized)
	var aliceSessions int
	db.QueryRow("SELECT COUNT(*) FROM APP.UserSession WHERE UserID = @p1", aliceID).Scan(&aliceSessions)
	if aliceSessions != 0 {
		t.Errorf("logout left %d session row(s) for alice", aliceSessions)
	}
}

type e2eClient struct {
	t    *testing.T
	base string
	http *http.Client
}

func newE2EClient(t *testing.T, base string) *e2eClient {
	jar, _ := cookiejar.New(nil)
	return &e2eClient{t: t, base: base, http: &http.Client{Jar: jar}}
}

func (c *e2eClient) expect(method, path, body string, want int) string {
	c.t.Helper()
	req, _ := http.NewRequest(method, c.base+path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != want {
		c.t.Fatalf("%s %s: want %d, got %d %s", method, path, want, resp.StatusCode, data)
	}
	return string(data)
}

func (c *e2eClient) register(phone string) int64 {
	c.t.Helper()
	body := c.expect(http.MethodPost, "/user/register", `{"user_name":"e2e","phone_number":"`+phone+`","password":"secret"}`, http.StatusOK)
	var id int64
	if _, err := fmt.Sscanf(body, `{"user_id":%d}`, &id); err != nil || id <= 0 {
		c.t.Fatalf("register response %q: %v", body, err)
	}
	return id
}

func (c *e2eClient) login(phone string) {
	c.t.Helper()
	c.expect(http.MethodPost, "/user/login", `{"phone_number":"`+phone+`","password":"secret"}`, http.StatusOK)
}

func grantSynops(t *testing.T, db *sqlx.DB, userID int64) {
	t.Helper()
	res, err := db.Exec(`INSERT INTO APP.User_WebAppLink (UserID, LinkID)
		SELECT @p1, ID FROM APP.WebAppLinkList WHERE Link = '/dashboards/synops'`, userID)
	if err != nil {
		t.Fatal(err)
	}
	if n, _ := res.RowsAffected(); n != 1 {
		t.Skip("Synops dashboard link not registered (sample_data/synops_dashboard_access.sql)")
	}
}
