package views

import (
	"net/http"
)

type LoginPage struct {
	BasePage
	Header          Header
	CallbackURL     string
	CsrfToken       string
	EnableUserLogin bool
	Title           string
}

type RegisterPage struct {
	BasePage
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
	p.BodyClass = "h-full min-h-screen bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 transition-colors duration-200 flex flex-col items-center justify-center px-4 sm:px-6 lg:px-8"
	p.CustomHeader = true
	p.CallbackURL = r.URL.Query().Get("callbackURL")
	return
}

func (p *RegisterPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	err, finished = p.Header.Load(r, w, vc)
	p.BodyClass = "h-full min-h-screen bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 transition-colors duration-200 flex flex-col items-center justify-center px-4 sm:px-6 lg:px-8"
	p.CustomHeader = true
	p.CallbackURL = r.URL.Query().Get("callbackURL")
	return
}

func (g *LoginPage) Copy() View    { return &LoginPage{} }
func (g *RegisterPage) Copy() View { return &RegisterPage{} }
