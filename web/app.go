package web

import (
	"net/http"
	// "github.com/go-session/cookie"

	"github.com/alexedwards/scs/v2"
	oa "github.com/panyam/oneauth"
	oa2 "github.com/panyam/oneauth/oauth2"
	"golang.org/x/oauth2"
)

var SUPERUSERS = map[string]bool{
	"sri.panyam@gmail.com": true,
}

type LCApp struct {
	// TODO - turn this over to slicer to manage clients
	Auth    *oa.OneAuth
	Session *scs.SessionManager

	mux     *http.ServeMux
	BaseUrl string
	// Template *template.Template
}

func NewWebApp() (app *LCApp, err error) {
	session := scs.New() //scs.NewCookieManager("u46IpCV9y5Vlur8YvODJEhgOY8m9JVE4"),
	// session.Store = NewMemoryStore(0)

	oneauth := oa.New("LeetCoach")
	oneauth.Session = session
	oneauth.Middleware.SessionGetter = func(r *http.Request, key string) any {
		return session.GetString(r.Context(), key)
	}
	oneauth.AddAuth("/google", oa2.NewGoogleOAuth2("", "", "", oneauth.SaveUserAndRedirect).Handler())
	oneauth.AddAuth("/github", oa2.NewGithubOAuth2("", "", "", oneauth.SaveUserAndRedirect).Handler())

	app = &LCApp{
		Session: session,
		Auth:    oneauth,
	}

	// TODO - setup oneauth.UserStore
	oneauth.UserStore = app

	// TODO - use godotenv and move configs to .env files instead
	/*
		if os.Getenv("LEETCOACH_ENV") == "dev" {
			n.authConfigs = DEV_CONFIGS
		}
	*/
	return
}

func (n *LCApp) Handler() http.Handler {
	n.mux = http.NewServeMux()

	// This single file serving is only for dev
	// n.mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./static/webapp/icons/favicon.png") })
	// n.mux.Handle("/auth/", http.StripPrefix("/auth", n.Auth.Handler()))
	n.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// TODO - turn this into a handle that will dynamically create case studies based on path and contents
	// n.mux.Handle("/casestudies/bitly", NewCaseStudy("../casestudies/bitly").Handler())
	// n.mux.HandleFunc("/cases/bitly", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "Did this work?") })
	n.mux.Handle("/cases/bitly/", http.StripPrefix("/cases/bitly", NewCaseStudy("./casestudies/bitly").Handler()))

	return n.mux
	// return n.Session.LoadAndSave(n.mux)
}

type LCAuthUser struct {
	id string
}

func (n *LCAuthUser) Id() string {
	return n.id
}

func (n *LCApp) GetUserByID(userId string) (oa.User, error) {
	var user LCAuthUser
	var err error
	// user.User, err = n.ClientMgr.GetAuthService().GetUserByID(userId)
	return &user, err
}

func (n *LCApp) EnsureAuthUser(authtype string, provider string, token *oauth2.Token, userInfo map[string]any) (oa.User, error) {
	var user LCAuthUser
	var err error
	// user.User, err = n.ClientMgr.GetAuthService().EnsureAuthUser(authtype, provider, token, userInfo)
	return &user, err
}
