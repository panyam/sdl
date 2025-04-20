package views

import (
	"net/http"
)

type Header struct {
	AppName        string
	IsLoggedIn     bool
	LoggedInUserId string
}

func (h *Header) SetupDefaults() {
	h.AppName = "LeetCoach"
}

func (v *Header) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	v.SetupDefaults()
	v.LoggedInUserId = vc.AuthMiddleware.GetLoggedInUserId(r)
	v.IsLoggedIn = v.LoggedInUserId != ""
	return
}
