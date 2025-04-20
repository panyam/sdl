package views

import (
	"net/http"
)

type LoginPage struct {
	Header          Header
	CallbackURL     string
	CsrfToken       string
	EnableUserLogin bool
}

type RegisterPage struct {
	Header         Header
	CallbackURL    string
	CsrfToken      string
	Name           string
	Email          string
	Password       string
	VerifyPassword string
	Errors         map[string]string
}

func (p *LoginPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	err, finished = p.Header.Load(r, w, vc)
	p.CallbackURL = r.URL.Query().Get("callbackURL")
	return
}

func (p *RegisterPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	err, finished = p.Header.Load(r, w, vc)
	p.CallbackURL = r.URL.Query().Get("callbackURL")
	return
}

func (g *LoginPage) Copy() View    { return &LoginPage{} }
func (g *RegisterPage) Copy() View { return &RegisterPage{} }
