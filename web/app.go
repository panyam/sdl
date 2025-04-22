package web

import (
	"log"
	"net/http"

	// "github.com/go-session/cookie"

	"github.com/alexedwards/scs/v2"
	svc "github.com/panyam/leetcoach/services"
	views "github.com/panyam/leetcoach/web/views"
	oa "github.com/panyam/oneauth"
	oa2 "github.com/panyam/oneauth/oauth2"
)

type LCContext struct {
	// Templates *templar.TemplateGroup
}

type LCApp struct {
	// TODO - turn this over to slicer to manage clients
	ClientMgr *svc.ClientMgr
	Api       *LCApi
	Auth      *oa.OneAuth
	Views     *views.LCViews
	Session   *scs.SessionManager

	mux     *http.ServeMux
	BaseUrl string
	Context LCContext
}

func NewWebApp(grpcAddress string, ClientMgr *svc.ClientMgr) (app *LCApp, err error) {
	session := scs.New() //scs.NewCookieManager("u46IpCV9y5Vlur8YvODJEhgOY8m9JVE4"),
	// session.Store = NewMemoryStore(0)

	oneauth := oa.New("LeetCoach")
	oneauth.Session = session
	oneauth.UsernameField = "email"
	oneauth.Middleware.SessionGetter = func(r *http.Request, key string) any {
		return session.GetString(r.Context(), key)
	}
	oneauth.AddAuth("/google", oa2.NewGoogleOAuth2("", "", "", oneauth.SaveUserAndRedirect).Handler())
	oneauth.AddAuth("/github", oa2.NewGithubOAuth2("", "", "", oneauth.SaveUserAndRedirect).Handler())

	app = &LCApp{
		ClientMgr: ClientMgr,
		Session:   session,
		Auth:      oneauth,
		Api:       NewLCApi(grpcAddress, &oneauth.Middleware, ClientMgr),
		Views:     views.NewLCViews(&oneauth.Middleware, ClientMgr),
	}
	oneauth.ValidateUsernamePassword = app.ValidateUsernamePassword

	// TODO - setup oneauth.UserStore
	oneauth.UserStore = app

	// TODO - use godotenv and move configs to .env files instead
	/*
		if os.Getenv("LEETCOACH_ENV") == "dev" {
			n.authConfigs = DEV_CONFIGS
		}
	*/

	/*
		templates := templar.NewTemplateGroup()
		templates.Loader = (&templar.LoaderList{}).AddLoader(templar.NewFileSystemLoader("./web/templates"))
		templates.AddFuncs(gotl.DefaultFuncMap())
		templates.AddFuncs(template.FuncMap{
			"Ctx": func() *LCContext {
				return &app.Context
			},
			"AsHtmlAttribs": func(m map[string]string) template.HTML {
				return `a = 'b' c = 'd'`
			},
			"Indented": func(nspaces int, code string) (formatted string) {
				// TBD
				lines := (strings.Split(strings.TrimSpace(code), "\n"))
				return strings.Join(lines, "<br/>")
			},
		})
		app.Context = LCContext{
			Templates: templates,
		}
	*/
	return
}

func (n *LCApp) Handler() http.Handler {
	n.mux = http.NewServeMux()

	// This single file serving is only for dev
	// n.mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./static/webapp/icons/favicon.png") })
	// n.mux.Handle("/auth/", http.StripPrefix("/auth", n.Auth.Handler()))
	n.mux.Handle("/auth/", http.StripPrefix("/auth", n.Auth.Handler()))
	n.mux.Handle("/api/", http.StripPrefix("/api", n.Api.Handler()))
	n.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	n.mux.Handle("/", n.Views.Handler())

	// For handling drawings
	// n.mux.Handle("/api/drawings/", http.StripPrefix("/api/drawings", NewDrawingApi("./content").Handler()))

	return n.Session.LoadAndSave(n.mux)
}

func (n *LCApp) GetUser(r *http.Request) *svc.User {
	userId := n.Auth.Middleware.GetLoggedInUserId(r)
	if userId == "" {
		return nil
	}

	log.Println("LoggedInUser: ", userId)
	userdsc := n.ClientMgr.GetUserDSClient()
	var user svc.User
	err := userdsc.GetByID(userId, &user)
	if err != nil {
		log.Print("Could not find user: ", userId)
		return nil
	}
	// TODO - Also validate the user so it cant just be "set"
	return &user
}

/*
func (n *LCApp) RegisterCaseStudy(path, folder string) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	tostrip := path[:len(path)-1]
	cs := NewDrawingApi(folder)
	// cs.Templates = n.Context.Templates
	n.mux.Handle(path, http.StripPrefix(tostrip, cs.Handler()))
}
*/
