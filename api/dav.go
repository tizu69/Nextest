package api

import (
	"context"
	"encoding/xml"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"g.tizu.dev/Nextest/config"
	"github.com/djherbis/times"
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

	user := a.RequireAuth(w, r)
	if user == "" {
		w.WriteHeader(401)
		return
	}
	r = r.WithContext(context.WithValue(r.Context(), ctxAuthedUser{}, user))

	// Nextcloud client sends a HEAD request to /remote.php/dav/~/ on log-in,
	// which the go webdav server will politely 405 Method Not Allowed on, as
	// it doesn't support HEAD on directories.
	// TODO: is this a spec violation on Go or Nextcloud's part?
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
	slog.Info("opening file", "path", path, "flag", flag, "perm", perm, "user", ctxAuthedUserGet(ctx), "name", name)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, flag, perm)
	return &davFile{f}, err
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
	slog.Info("stat", "path", path, "user", ctxAuthedUserGet(ctx), "name", name)
	if err != nil {
		return nil, err
	}
	return os.Stat(path)
}

type davFile struct {
	*os.File
}

var (
	_ webdav.File            = (*davFile)(nil)
	_ webdav.DeadPropsHolder = (*davFile)(nil)
)

func (f *davFile) DeadProps() (map[xml.Name]webdav.Property, error) {
	var (
		xmlName    xml.Name
		property   webdav.Property
		properties = make(map[xml.Name]webdav.Property)
	)

	stat, err1 := f.Stat()
	times, err2 := times.StatFile(f.File)
	if err := errors.Join(err1, err2); err != nil {
		return nil, err
	}

	xmlName.Space = "DAV:"
	xmlName.Local = "creationdate"
	property.XMLName = xmlName
	property.InnerXML = []byte(times.BirthTime().Format(time.RFC3339))
	properties[xmlName] = property

	_ = stat
	return properties, nil
}

func (f *davFile) Patch([]webdav.Proppatch) ([]webdav.Propstat, error) {
	return nil, nil
}
