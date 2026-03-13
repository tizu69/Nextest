package api

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"g.tizu.dev/Nextest/template"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type pendAuth struct {
	authing map[string]string
	done    map[string]string
	mu      sync.Mutex
}

func newPendAuth() *pendAuth {
	return &pendAuth{
		authing: make(map[string]string), done: make(map[string]string),
	}
}

type loginReq struct {
	Poll  loginReqPoll `json:"poll"`
	Login string       `json:"login"`
}
type loginReqPoll struct {
	Token    string `json:"token"`
	Endpoint string `json:"endpoint"`
}

func (a *API) RouteLogin(w http.ResponseWriter, r *http.Request) {
	token, ua := uuid.New().String(), r.Header.Get("User-Agent")
	a.pendAuth.mu.Lock()
	a.pendAuth.authing[token] = ua
	a.pendAuth.mu.Unlock()
	time.AfterFunc(20*time.Minute, func() {
		a.pendAuth.mu.Lock()
		delete(a.pendAuth.authing, token)
		a.pendAuth.mu.Unlock()
	})

	url := *r.URL
	url.Scheme = requestScheme(r)
	url.Host = r.Host
	render.JSON(w, r, loginReq{
		Poll: loginReqPoll{
			Token:    token,
			Endpoint: url.JoinPath("/" + token + "/poll").String(),
		},
		Login: url.JoinPath("/" + token).String(),
	})
}

func requestScheme(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-Proto"); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

type loginFlowModel struct {
	UA    string
	Token string
	User  string
	Error string
}

func (a *API) RouteLoginFlow(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	a.pendAuth.mu.Lock()
	ua, ok := a.pendAuth.authing[token]
	a.pendAuth.mu.Unlock()
	if !ok {
		renderHtml(w, r, "error.gohtml", "This login session has expired. Please try again.")
		return
	}

	renderHtml(w, r, "login.gohtml", loginFlowModel{
		UA: ua, Token: token,
	})
}

func (a *API) RouteLoginFlowTry(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	a.pendAuth.mu.Lock()
	ua, ok := a.pendAuth.authing[token]
	a.pendAuth.mu.Unlock()
	if !ok {
		renderHtml(w, r, "error.gohtml", "This login session has expired. Please try again.")
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	if username == "" || password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	if !a.Config.Accounts.Validate(username, password) {
		renderHtml(w, r, "login.gohtml", loginFlowModel{
			UA: ua, Token: token, Error: "invalid username or password", User: username,
		})
		return
	}

	a.pendAuth.mu.Lock()
	a.pendAuth.done[token] = username
	delete(a.pendAuth.authing, token)
	a.pendAuth.mu.Unlock()
	time.AfterFunc(1*time.Minute, func() {
		a.pendAuth.mu.Lock()
		delete(a.pendAuth.done, token)
		a.pendAuth.mu.Unlock()
	})

	http.Redirect(w, r, "/index.php/login/v2/done", http.StatusFound)
}

func (a *API) RouteLoginFlowDone(w http.ResponseWriter, r *http.Request) {
	renderHtml(w, r, "logindone.gohtml", nil)
}

type loginFlowPollResp struct {
	Server      string `json:"server"`
	LoginName   string `json:"loginName"`
	AppPassword string `json:"appPassword"`
}

func (a *API) RouteLoginFlowPoll(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	a.pendAuth.mu.Lock()
	username, ok := a.pendAuth.done[token]
	_, authing := a.pendAuth.authing[token]
	a.pendAuth.mu.Unlock()
	if !ok {
		status := http.StatusNotFound
		if !authing {
			status = http.StatusBadRequest
		}
		http.Error(w, "[]", status)
		return
	}

	url := *r.URL
	url.Scheme = requestScheme(r)
	url.Host = r.Host
	url.Path = ""
	url.RawQuery = ""

	render.JSON(w, r, loginFlowPollResp{
		Server:      url.String(),
		LoginName:   username,
		AppPassword: a.Config.Accounts[username],
	})
}

func renderHtml(w http.ResponseWriter, r *http.Request, name string, data any) {
	templ, err := template.Render(name, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render.HTML(w, r, templ)
}
