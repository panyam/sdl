package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	// "github.com/go-session/cookie"

	"github.com/alexedwards/scs/v2"
	oa "github.com/panyam/oneauth"
	oa2 "github.com/panyam/oneauth/oauth2"
	gotl "github.com/panyam/templar"
	"golang.org/x/oauth2"
)

var SUPERUSERS = map[string]bool{
	"sri.panyam@gmail.com": true,
}

type LCContext struct {
	Templates *gotl.TemplateGroup
}

type LCApp struct {
	// TODO - turn this over to slicer to manage clients
	Auth    *oa.OneAuth
	Session *scs.SessionManager

	mux     *http.ServeMux
	BaseUrl string
	Context LCContext
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

	templates := gotl.NewTemplateGroup()
	templates.Loader = (&gotl.LoaderList{}).AddLoader(gotl.NewFileSystemLoader("./web/templates"))
	templates.AddFuncs(gotl.DefaultFuncMap())
	templates.AddFuncs(template.FuncMap{
		"Ctx": func() *LCContext {
			return &app.Context
		},
		"AsHtmlAttribs": func(m map[string]string) template.HTML {
			return `a = 'b' c = 'd'`
		},
		"Ago": func(t time.Time) string {
			diff := time.Since(t)

			if years := int64(diff.Hours() / (365 * 24)); years > 0 {
				return fmt.Sprintf("%d years ago", years)
			}

			if months := int64(diff.Hours() / (30 * 24)); months > 0 {
				return fmt.Sprintf("%d months ago", months)
			}

			if weeks := int64(diff.Hours() / (7 * 24)); weeks > 0 {
				return fmt.Sprintf("%d weeks ago", weeks)
			}

			if days := int64(diff.Hours() / (24)); days > 0 {
				return fmt.Sprintf("%d days ago", days)
			}

			if hours := int64(diff.Hours()); hours > 0 {
				return fmt.Sprintf("%d hours ago", hours)
			}

			if minutes := int64(diff.Minutes()); minutes > 0 {
				return fmt.Sprintf("%d minutes ago", minutes)
			}

			if diff.Seconds() > 0 {
				return fmt.Sprintf("%d seconds ago", int64(diff.Seconds()))
			}
			return "just now"
		},
		"Indented": func(nspaces int, code string) (formatted string) {
			lines := (strings.Split(strings.TrimSpace(code), "\n"))
			return strings.Join(lines, "<br/>")
		},
	})
	app.Context = LCContext{
		Templates: templates,
	}
	return
}

func (n *LCApp) Handler() http.Handler {
	n.mux = http.NewServeMux()

	// This single file serving is only for dev
	// n.mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./static/webapp/icons/favicon.png") })
	// n.mux.Handle("/auth/", http.StripPrefix("/auth", n.Auth.Handler()))
	n.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// For handling drawings
	n.mux.Handle("/api/drawings/", http.StripPrefix("/api/drawings", NewDrawingApi("./content").Handler()))

	// n.mux.HandleFunc("/cases/bitly", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "Did this work?") })
	n.mux.Handle("/", &site)
	// n.RegisterCaseStudy("/cases/bitly", "./casestudies/bitly")

	return n.mux
	// return n.Session.LoadAndSave(n.mux)
}

func (n *LCApp) RegisterCaseStudy(path, folder string) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	tostrip := path[:len(path)-1]
	cs := NewDrawingApi(folder)
	cs.Templates = n.Context.Templates
	n.mux.Handle(path, http.StripPrefix(tostrip, cs.Handler()))
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
