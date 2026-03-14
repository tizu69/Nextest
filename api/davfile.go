package api

import (
	"encoding/xml"
	"errors"
	"hash/fnv"
	"os"
	"strconv"
	"time"

	"github.com/djherbis/times"
	"golang.org/x/net/webdav"
)

type davFile struct {
	*os.File
}

var (
	_ webdav.File            = (*davFile)(nil)
	_ webdav.DeadPropsHolder = (*davFile)(nil)
)

// DeadProps provides WebDAV props, see [the Nextcloud docs].
//
// [the Nextcloud docs]: https://github.com/nextcloud/documentation/blob/master/developer_manual/client_apis/WebDAV/basic.rst
func (f *davFile) DeadProps() (map[xml.Name]webdav.Property, error) {
	stat, err1 := f.Stat()
	times, err2 := times.StatFile(f.File)
	if err := errors.Join(err1, err2); err != nil {
		return nil, err
	}

	h := fnv.New64a()
	h.Write([]byte(f.Name()))
	fileid := h.Sum64()

	b := NewDeadPropBuilder()
	b.DAV("creationdate", times.BirthTime().Format(time.RFC3339))
	b.OC("permissions", "MGDNVWCK")
	b.OC("owner-id", "~")
	b.OC("owner-display-name", "~")
	b.OC("id", strconv.Itoa(int(fileid)))
	b.OC("fileid", strconv.Itoa(int(fileid)))
	b.OC("size", strconv.Itoa(int(stat.Size())))

	return b.Build(), nil
}

func (f *davFile) Patch([]webdav.Proppatch) ([]webdav.Propstat, error) {
	return nil, nil
}

// DeadPropBuilder is a helper for building [webdav.DeadPropsHolder]s.
type DeadPropBuilder struct {
	m map[xml.Name]webdav.Property
}

func NewDeadPropBuilder() *DeadPropBuilder {
	return &DeadPropBuilder{
		m: make(map[xml.Name]webdav.Property),
	}
}

func (b *DeadPropBuilder) Raw(ns, name, innerXML string) {
	n := xml.Name{Space: ns, Local: name}
	b.m[n] = webdav.Property{XMLName: n, InnerXML: []byte(innerXML)}
}

func (b *DeadPropBuilder) DAV(name, value string) {
	b.Raw("DAV:", name, value)
}

func (b *DeadPropBuilder) OC(name, value string) {
	b.Raw("http://owncloud.org/ns", name, value)
}

func (b *DeadPropBuilder) NC(name, value string) {
	b.Raw("http://nextcloud.org/ns", name, value)
}

func (b *DeadPropBuilder) OCS(name, value string) {
	b.Raw("http://open-collaboration-services.org/ns", name, value)
}

func (b *DeadPropBuilder) OCM(name, value string) {
	b.Raw("http://open-cloud-mesh.org/ns", name, value)
}

func (b *DeadPropBuilder) Build() map[xml.Name]webdav.Property {
	return b.m
}
