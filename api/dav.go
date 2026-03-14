package api

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"g.tizu.dev/Nextest/config"
	"github.com/go-chi/chi/v5"
	"golang.org/x/net/webdav"
)

func init() {
	chi.RegisterMethod("PROPFIND")
	chi.RegisterMethod("PROPPATCH")
	chi.RegisterMethod("MKOL")
	chi.RegisterMethod("COPY")
	chi.RegisterMethod("MOVE")
	chi.RegisterMethod("LOCK")
	chi.RegisterMethod("UNLOCK")
}

func (a *API) RouteWebDAV(w http.ResponseWriter, r *http.Request) {
	if a.dav == nil {
		a.dav = &webdav.Handler{
			Prefix:     "/remote.php/dav/files/~",
			FileSystem: NewDavFS(a.Config),
			LockSystem: webdav.NewMemLS(),
			Logger: func(r *http.Request, err error) {
				if err != nil {
					slog.Error("webdav", "err", err)
				}
			},
		}
	}

	// GNOME sends OPTIONS to verify the server is WebDAV-compatible, without
	// sending credentials. That would 401, so we handle it here.
	if r.URL.Path == "/remote.php/webdav/" && r.Method == "OPTIONS" {
		w.WriteHeader(204)
		w.Header().Set("Allow", "OPTIONS, LOCK, PUT, MKCOL")
		w.Header().Set("DAV", "1, 2")
		w.Header().Set("MS-Author-Via", "DAV")
		return
	}

	// Nextcloud client sends a HEAD request to /remote.php/dav/~/ on log-in,
	// which the go webdav server will politely 405 Method Not Allowed on, as
	// it doesn't support HEAD on directories.
	// TODO: is this a spec violation on Go or Nextcloud's part?
	if r.URL.Path == "/remote.php/dav/files/~/" && r.Method == "HEAD" {
		w.WriteHeader(200)
		return
	}

	user := a.RequireAuth(w, r)
	if user == "" {
		w.WriteHeader(401)
		return
	}
	r = r.WithContext(context.WithValue(r.Context(), ctxAuthedUser{}, user))

	a.dav.ServeHTTP(w, r)
}

type DavFS struct {
	Config *config.Config
}

func NewDavFS(cfg *config.Config) *DavFS {
	return &DavFS{cfg}
}

type ctxAuthedUser struct{}

func ctxAuthedUserGet(ctx context.Context) string {
	user, _ := ctx.Value(ctxAuthedUser{}).(string)
	return user
}

func (f *DavFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	path, _, err := f.Config.Mount.Real(name, ctxAuthedUserGet(ctx))
	if err != nil {
		return err
	}
	return os.Mkdir(path, perm)
}

func (fs *DavFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	path, _, err := fs.Config.Mount.Real(name, ctxAuthedUserGet(ctx))
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, flag, perm)
	return &davFile{f}, err
}

func (fs *DavFS) RemoveAll(ctx context.Context, name string) error {
	path, _, err := fs.Config.Mount.Real(name, ctxAuthedUserGet(ctx))
	if err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func (fs *DavFS) Rename(ctx context.Context, oldName, newName string) error {
	oldPath, _, err := fs.Config.Mount.Real(oldName, ctxAuthedUserGet(ctx))
	if err != nil {
		return err
	}
	newPath, _, err := fs.Config.Mount.Real(newName, ctxAuthedUserGet(ctx))
	if err != nil {
		return err
	}
	return os.Rename(oldPath, newPath)
}

func (fs *DavFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	path, _, err := fs.Config.Mount.Real(name, ctxAuthedUserGet(ctx))
	if err != nil {
		return nil, err
	}
	return os.Stat(path)
}
