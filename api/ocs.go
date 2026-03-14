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

type ocsResp struct {
	Meta ocsMeta `json:"meta" xml:"meta"`
	Data any     `json:"data" xml:"data"`
}
type ocsResponse struct {
	OCS ocsResp `json:"ocs" xml:"ocs"`
}

type ocsMeta struct {
	Status string `json:"status" xml:"status"`
	Code   int    `json:"statuscode" xml:"statuscode"`
	Msg    string `json:"message" xml:"message"`
}

type ocsScope string

const (
	ocsScopePrivate   ocsScope = "v2-private"
	ocsScopeLocal     ocsScope = "v2-local"
	ocsScopeFederated ocsScope = "v2-federated"
	ocsScopePublished ocsScope = "v2-published"
)

type ocsBackendCapabilities struct {
	SetDisplayName bool `json:"setDisplayName" xml:"setDisplayName"`
	SetPassword    bool `json:"setPassword" xml:"setPassword"`
}

type ocsQuota struct {
	Free     int64 `json:"free" xml:"free"`
	Used     int64 `json:"used" xml:"used"`
	Total    int64 `json:"total" xml:"total"`
	Relative int64 `json:"relative" xml:"relative"`
	Quota    int64 `json:"quota" xml:"quota"`
}

type ocsUser struct {
	Enabled             bool                   `json:"enabled" xml:"enabled"`
	ID                  string                 `json:"id" xml:"id"`
	FirstLoginTimestamp int64                  `json:"firstLoginTimestamp" xml:"firstLoginTimestamp"`
	LastLoginTimestamp  int64                  `json:"lastLoginTimestamp" xml:"lastLoginTimestamp"`
	LastLogin           int64                  `json:"lastLogin" xml:"lastLogin"`
	Backend             string                 `json:"backend" xml:"backend"`
	Subadmin            []string               `json:"subadmin" xml:"subadmin"`
	Quota               ocsQuota               `json:"quota" xml:"quota"`
	Manager             string                 `json:"manager" xml:"manager"`
	AvatarScope         ocsScope               `json:"avatarScope" xml:"avatarScope"`
	Email               string                 `json:"email" xml:"email"`
	EmailScope          ocsScope               `json:"emailScope" xml:"emailScope"`
	AdditionalMail      []string               `json:"additional_mail" xml:"additional_mail"`
	AdditionalMailScope []string               `json:"additional_mailScope" xml:"additional_mailScope"`
	Displayname         string                 `json:"displayname" xml:"displayname"`
	DisplayName         string                 `json:"display-name" xml:"display-name"`
	DisplaynameScope    ocsScope               `json:"displaynameScope" xml:"displaynameScope"`
	Phone               string                 `json:"phone" xml:"phone"`
	PhoneScope          ocsScope               `json:"phoneScope" xml:"phoneScope"`
	Address             string                 `json:"address" xml:"address"`
	AddressScope        ocsScope               `json:"addressScope" xml:"addressScope"`
	Website             string                 `json:"website" xml:"website"`
	WebsiteScope        ocsScope               `json:"websiteScope" xml:"websiteScope"`
	Twitter             string                 `json:"twitter" xml:"twitter"`
	TwitterScope        ocsScope               `json:"twitterScope" xml:"twitterScope"`
	Fediverse           string                 `json:"fediverse" xml:"fediverse"`
	FediverseScope      ocsScope               `json:"fediverseScope" xml:"fediverseScope"`
	Organisation        string                 `json:"organisation" xml:"organisation"`
	OrganisationScope   ocsScope               `json:"organisationScope" xml:"organisationScope"`
	Role                string                 `json:"role" xml:"role"`
	RoleScope           ocsScope               `json:"roleScope" xml:"roleScope"`
	Headline            string                 `json:"headline" xml:"headline"`
	HeadlineScope       ocsScope               `json:"headlineScope" xml:"headlineScope"`
	Biography           string                 `json:"biography" xml:"biography"`
	BiographyScope      ocsScope               `json:"biographyScope" xml:"biographyScope"`
	ProfileEnabled      string                 `json:"profile_enabled" xml:"profile_enabled"`
	ProfileEnabledScope ocsScope               `json:"profile_enabledScope" xml:"profile_enabledScope"`
	Pronouns            string                 `json:"pronouns" xml:"pronouns"`
	PronounsScope       ocsScope               `json:"pronounsScope" xml:"pronounsScope"`
	Groups              []string               `json:"groups" xml:"groups"`
	Language            string                 `json:"language" xml:"language"`
	Locale              string                 `json:"locale" xml:"locale"`
	NotifyEmail         *string                `json:"notify_email" xml:"notify_email"`
	BackendCapabilities ocsBackendCapabilities `json:"backendCapabilities" xml:"backendCapabilities"`
	Bluesky             string                 `json:"bluesky" xml:"bluesky"`
	BlueskyScope        ocsScope               `json:"blueskyScope" xml:"blueskyScope"`
	Timezone            string                 `json:"timezone" xml:"timezone"`
	StorageLocation     string                 `json:"storageLocation" xml:"storageLocation"`
}

func (a *API) RouteGetUser(w http.ResponseWriter, r *http.Request) {
	user := a.RequireAuth(w, r)
	if user == "" {
		render.Status(r, 401)
		render.JSON(w, r, ocsResponse{
			ocsResp{
				Meta: ocsMeta{Status: "unauthorized", Msg: "Unauthorized", Code: 401},
			},
		})
		return
	}

	render.JSON(w, r, ocsResponse{
		ocsResp{
			Meta: ocsMeta{Status: "ok", Msg: "OK", Code: 200},
			Data: ocsUser{
				AdditionalMail:      []string{},
				Backend:             "nextest",
				BackendCapabilities: ocsBackendCapabilities{},
				Displayname:         user,
				DisplayName:         user,
				Groups:              []string{},
				ID:                  "~", // this seems to be used for the webdav path only
				Language:            "en",
				Locale:              "us",
				Quota:               ocsQuota{},
				Subadmin:            []string{},
			},
		},
	})
}

type ocsCapabilities struct {
	Version      ocsCapabilitiesVersion `json:"version" xml:"version"`
	Capabilities map[string]any         `json:"capabilities" xml:"capabilities"`
}

type ocsCapabilitiesVersion struct {
	Major           int    `json:"major" xml:"major"`
	Minor           int    `json:"minor" xml:"minor"`
	Micro           int    `json:"micro" xml:"micro"`
	String          string `json:"string" xml:"string"`
	Edition         string `json:"edition" xml:"edition"`
	ExtendedSupport bool   `json:"extendedSupport" xml:"extendedSupport"`
}

func (a *API) RouteGetCapabilities(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, ocsResponse{
		ocsResp{
			Meta: ocsMeta{Status: "ok", Msg: "OK", Code: 200},
			Data: ocsCapabilities{
				Version: ocsCapabilitiesVersion{
					Major:           0,
					Minor:           0,
					Micro:           1,
					String:          "0.0.1",
					Edition:         "",
					ExtendedSupport: false,
				},
				Capabilities: map[string]any{
					"core": map[string]any{
						"pollinterval":        60,
						"webdav-root":         "remote.php/webdav",
						"reference-api":       true,
						"reference-regex":     "(\\s|\\n|^)(https?:\\/\\/)([-A-Z0-9+_.]+(?::[0-9]+)?(?:\\/[-A-Z0-9+&@#%?=~_|!:,.;()]*)*)(\\s|\\n|$)",
						"mod-rewrite-working": false,
					},
				},
			},
		},
	})
}

func (a *API) RoutePredefinedStatuses(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, ocsResponse{
		ocsResp{
			Meta: ocsMeta{Status: "ok", Msg: "OK", Code: 200},
			Data: []int{},
		},
	})
}
