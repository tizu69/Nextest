package api

import (
	"net/http"

	"github.com/go-chi/render"
)

type respGetStatus struct {
	Installed       bool   `json:"installed" xml:"installed"`
	Maintenance     bool   `json:"maintenance" xml:"maintenance"`
	NeedsDbUpgrade  bool   `json:"needsDbUpgrade" xml:"needs-db-upgrade"`
	Version         string `json:"version" xml:"version"`
	Versionstring   string `json:"versionstring" xml:"versionstring"`
	Edition         string `json:"edition" xml:"edition"`
	Productname     string `json:"productname" xml:"productname"`
	ExtendedSupport bool   `json:"extendedSupport" xml:"extended-support"`
}

func (a *API) RouteGetStatus(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, respGetStatus{
		Installed:       true,
		Maintenance:     false,
		NeedsDbUpgrade:  false,
		Version:         "0.0.1",
		Versionstring:   "0.0.1",
		Edition:         "",
		Productname:     "Nextest",
		ExtendedSupport: false,
	})
}
