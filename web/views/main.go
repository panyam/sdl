package views

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	gotl "github.com/panyam/goutils/template"
	svc "github.com/panyam/leetcoach/services"
	oa "github.com/panyam/oneauth"
	tmplr "github.com/panyam/templar"
)

type ViewContext struct {
	AuthMiddleware *oa.Middleware
	ClientMgr      *svc.ClientMgr
	Ctx            context.Context
	Templates      *tmplr.TemplateGroup
}

type ViewMaker func() View

type Copyable interface {
	Copy() View
}

func Copier[V Copyable](v V) ViewMaker {
	return v.Copy
}

type LCViews struct {
	mux     *http.ServeMux
	Context *ViewContext
}

type View interface {
	Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool)
}

func (b *LCViews) ViewRenderer(view ViewMaker, template string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b.RenderView(view(), template, r, w)
	}
}

func (b *LCViews) RenderView(view View, template string, r *http.Request, w http.ResponseWriter) {
	if template == "" {
		t := reflect.TypeOf(view)
		e := t.Elem()
		template = e.Name()
	}
	err, finished := view.Load(r, w, b.Context)
	if !finished {
		if err != nil {
			log.Println("Error: ", err)
			fmt.Fprint(w, "Error rendering: ", err.Error())
		} else {
			tmpl, err := b.Context.Templates.Loader.Load(template, "")
			if err != nil {
				log.Println("Template Load Error: ", template, err)
				fmt.Fprint(w, "Error rendering: ", err.Error())
			} else {
				b.Context.Templates.RenderHtmlTemplate(w, tmpl[0], template, view, nil)
			}
		}
	}
}

func (b *LCViews) HandleError(err error, w io.Writer) {
	if err != nil {
		fmt.Fprint(w, "Error rendering: ", err.Error())
	}
}

func (n *LCViews) Handler() http.Handler {
	return n.mux
}

func NewLCViews(middleware *oa.Middleware, clients *svc.ClientMgr) *LCViews {
	out := LCViews{
		mux: http.NewServeMux(),
	}

	templates := tmplr.NewTemplateGroup()
	templates.Loader = (&tmplr.LoaderList{}).AddLoader(tmplr.NewFileSystemLoader("./web/views/templates"))
	templates.AddFuncs(gotl.DefaultFuncMap())
	templates.AddFuncs(template.FuncMap{
		"Ctx": func() *ViewContext {
			return out.Context
		},
		"UserInfo": func(userId string) map[string]any {
			// Just a hacky cache
			return map[string]any{
				"FullName":  "XXXX YYY",
				"Name":      "XXXX",
				"AvatarUrl": "/avatar/url",
			}
		},
		"Indented": func(nspaces int, code string) (formatted string) {
			lines := (strings.Split(strings.TrimSpace(code), "\n"))
			return strings.Join(lines, "<br/>")
		},
	})
	out.Context = &ViewContext{
		AuthMiddleware: middleware,
		ClientMgr:      clients,
		Templates:      templates,
	}

	// setup routes
	out.setupRoutes()
	return &out
}
