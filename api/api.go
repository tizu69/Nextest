package api

import (
	"log/slog"
	"net/http"

	"g.tizu.dev/Nextest/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/net/webdav"
)

type API struct {
	Config   *config.Config
	dav      *webdav.Handler
	pendAuth *pendAuth
}

func NewAPI(cfg *config.Config) *API {
	a := &API{Config: cfg}
	a.dav = &webdav.Handler{
		Prefix:     "/remote.php/dav",
		FileSystem: NewDavFS(a.Config),
		LockSystem: webdav.NewMemLS(),
	}
	a.pendAuth = newPendAuth()
	return a
}

// Routes returns a handler that serves all API endpoints (ServeMux).
func (a *API) Routes() http.Handler {
	mux := chi.NewMux()
	mux.Use(middleware.Logger)
	mux.Use(middleware.Compress(5))
	mux.Use(func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			prw := newPeekRW(w)
			next.ServeHTTP(prw, r)
			slog.Info("responded", "with", string(prw.Peek()))
		}
		return http.HandlerFunc(fn)
	})
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Recoverer)

	mux.Get("/status.php", a.RouteGetStatus)
	mux.HandleFunc("/remote.php/dav", a.RouteWebDAV)
	mux.HandleFunc("/remote.php/dav/", a.RouteWebDAV)
	mux.Post("/index.php/login/v2", a.RouteLogin)
	mux.Get("/index.php/login/v2/{token}", a.RouteLoginFlow)
	mux.Post("/index.php/login/v2/{token}", a.RouteLoginFlowTry)
	mux.Get("/index.php/login/v2/done", a.RouteLoginFlowDone)
	mux.Post("/index.php/login/v2/{token}/poll", a.RouteLoginFlowPoll)
	mux.Get("/ocs/v2.php/cloud/user", a.RouteGetUser)
	return mux
}

func (a *API) RequireAuth(w http.ResponseWriter, r *http.Request) string {
	name, pass, ok := r.BasicAuth()
	if !ok {
		return ""
	}
	if !a.Config.Accounts.Validate(name, pass) {
		return ""
	}
	return name
}
