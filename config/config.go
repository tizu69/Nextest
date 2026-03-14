package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Accounts Accounts `toml:"accounts"`
	Mount    Mounts   `toml:"mount"`
}

func Load(path string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

type Accounts map[string]string

func (a Accounts) Validate(name, passwrd string) bool {
	pw, ok := a[name]
	return ok && pw == passwrd
}

type Mounts map[string]Mount

type Mount struct {
	Real string `toml:"real"`
	// UserLocal mount points are unique to each user. In other words, if "/foo"
	// is a user-local mount point, then the real path, say "/mnt/foo", will
	// contain a directory for each user, e.g. alice's "/foo/bar" will be mapped
	// to the real path "/mnt/foo/alice/bar".
	UserLocal bool `toml:"user-local"`
}

func (m Mounts) Real(path, user string) (string, Mount, error) {
	if path == "" {
		return "", Mount{}, os.ErrNotExist
	}
	path = filepath.Clean(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	parts := strings.Split(path, "/")
	for i := len(parts); i > 0; i-- {
		key := strings.Join(parts[:i], "/")
		if key == "" {
			key = "/"
		}
		mount, ok := m[key]
		if !ok {
			continue
		}

		base := mount.Real
		if mount.UserLocal {
			if user == "" {
				return "", Mount{}, os.ErrNotExist
			}
			base = filepath.Join(base, user)
		}

		rest := strings.Join(parts[i:], "/")
		if rest == "" {
			return base, mount, nil
		}
		return filepath.Join(base, rest), mount, nil
	}
	return "", Mount{}, os.ErrNotExist
}
