package views

import "net/http"

type GenericPage struct {
	Header Header
}

func (g *GenericPage) Copy() View { return &GenericPage{} }

func (v *GenericPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	return
}
