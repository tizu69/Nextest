package config

import "testing"

func TestMounts(t *testing.T) {
	m := Mounts{
		"/":         Mount{Real: "/mnt"},
		"/personal": Mount{Real: "/home", UserLocal: true},
		"/bin":      Mount{Real: "/usr/bin"},
	}

	cases := []struct {
		Path string
		Want string
	}{
		// Global paths
		{"/", "/mnt"},
		{"/foo", "/mnt/foo"},
		{"/foo/bar", "/mnt/foo/bar"},

		// Odd slashes
		{"/", "/mnt"}, // only slashes
		{"//", "/mnt"},
		{"/foo//bar", "/mnt/foo/bar"}, // mid-path doubles
		{"/foo///bar", "/mnt/foo/bar"},
		{"/foo/", "/mnt/foo"}, // trailing slashes
		{"/foo//", "/mnt/foo"},
		{"foo", "/mnt/foo"}, // no leading slash
		{"foo/bar", "/mnt/foo/bar"},

		// Mount points inside other mount points
		{"/bin", "/usr/bin"},
		{"/bin/foo", "/usr/bin/foo"},

		// User-local paths
		{"/personal", "/home/__exampleuser__"},
		{"/personal/foo", "/home/__exampleuser__/foo"},

		// Paths with ..
		{"/personal/..", "/mnt"},
		{"/personal/foo/../..", "/mnt"},
		{"/../../../../..", "/mnt"},
	}

	for _, c := range cases {
		got, _, err := m.Real(c.Path, "__exampleuser__")
		if err != nil {
			t.Errorf("Translate(%q) (err) = %v; want %q", c.Path, err, c.Want)
		} else if got != c.Want {
			t.Errorf("Translate(%q) = %q; want %q", c.Path, got, c.Want)
		}
	}
}
