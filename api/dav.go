package api

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"g.tizu.dev/Nextest/config"
	"golang.org/x/net/webdav"
)

func (a *API) RouteWebDAV(w http.ResponseWriter, r *http.Request) {
	user := a.RequireAuth(w, r)
	if user == "" {
		w.WriteHeader(401)
		return
	}
	r = r.WithContext(context.WithValue(r.Context(), ctxAuthedUser{}, user))

	// Nextcloud client sends a HEAD request to /remote.php/dav/~/ on log-in,
	// which the go webdav server will politely 405 Method Not Allowed on.
	if r.URL.Path == "/remote.php/dav/files/~/" && r.Method == "HEAD" {
		w.WriteHeader(200)
		return
	}

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
	path, err := f.Config.Mount.Real(name, ctxAuthedUserGet(ctx))
	if err != nil {
		return err
	}
	return os.Mkdir(path, perm)
}

func (fs *DavFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	path, err := fs.Config.Mount.Real(name, ctxAuthedUserGet(ctx))
	if err != nil {
		return nil, err
	}
	slog.Info("opening file", "path", path, "flag", flag, "perm", perm, "user", ctxAuthedUserGet(ctx), "name", name)
	return os.OpenFile(path, flag, perm)
}

func (fs *DavFS) RemoveAll(ctx context.Context, name string) error {
	path, err := fs.Config.Mount.Real(name, ctxAuthedUserGet(ctx))
	if err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func (fs *DavFS) Rename(ctx context.Context, oldName, newName string) error {
	oldPath, err := fs.Config.Mount.Real(oldName, ctxAuthedUserGet(ctx))
	if err != nil {
		return err
	}
	newPath, err := fs.Config.Mount.Real(newName, ctxAuthedUserGet(ctx))
	if err != nil {
		return err
	}
	return os.Rename(oldPath, newPath)
}

func (fs *DavFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	path, err := fs.Config.Mount.Real(name, ctxAuthedUserGet(ctx))
	if err != nil {
		return nil, err
	}
	slog.Info("stat", "path", path, "user", ctxAuthedUserGet(ctx), "name", name)
	return os.Stat(path)
}
