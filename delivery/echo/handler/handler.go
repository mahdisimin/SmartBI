package echowebframework

import (
	"intelligentBI/pkg"
	sessionrepo "intelligentBI/repository/SQLServer/session"
	synopsrepo "intelligentBI/repository/SQLServer/synops"
	user "intelligentBI/repository/SQLServer/user"
	"intelligentBI/service"
	"intelligentBI/service/export"

	"github.com/jmoiron/sqlx"
)

// Handler holds the services every HTTP handler needs. It is built once at
// startup from the shared DB pool, so handlers never open their own
// connections.
type Handler struct {
	UserService    service.UserService
	ExportService  *export.ExportService
	SessionService service.SessionService
	// CookieSecure sets the Secure flag on the session cookie.
	CookieSecure bool
}

func NewHandler(db *sqlx.DB, sessionCfg pkg.SessionConfig) *Handler {
	userRepo := user.NewUser(db)

	exportService := export.NewExportService(synopsrepo.NewUserActivity(db))
	exportService.Access = userRepo

	return &Handler{
		UserService:   service.UserService{Repository: userRepo},
		ExportService: exportService,
		SessionService: service.SessionService{
			Repository: sessionrepo.NewSession(db),
			TTL:        sessionCfg.TTL,
		},
		CookieSecure: sessionCfg.CookieSecure,
	}
}
