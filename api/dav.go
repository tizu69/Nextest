package api

import (
	"context"
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
	return os.Stat(path)
}
