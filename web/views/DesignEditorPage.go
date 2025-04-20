package views

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
)

type DesignEditorPage struct {
	Header        Header
	IsOwner       bool
	DesignId      string
	Design        *protos.Design
	Errors        map[string]string
	AllowCustomId bool
}

func (g *DesignEditorPage) Copy() View { return &DesignEditorPage{} }

func (v *DesignEditorPage) SetupDefaults() {
}

func (v *DesignEditorPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	v.Header.Load(r, w, vc)
	v.SetupDefaults()
	queryParams := r.URL.Query()
	v.DesignId = r.PathValue("designId")
	templateName := queryParams.Get("template")
	loggedInUserId := vc.AuthMiddleware.GetLoggedInUserId(r)

	slog.Info("Loading editor for design with ID: ", "nid", v.DesignId)

	if v.DesignId == "" {
		if loggedInUserId == "" {
			// For now enforce login even on new
			qs := r.URL.RawQuery
			if len(qs) > 0 {
				qs = "?" + qs
			}
			http.Redirect(w, r, fmt.Sprintf("/login?callbackURL=%s", fmt.Sprintf("/designs/new%s", qs)), http.StatusSeeOther)
			return nil, true
		}
		v.IsOwner = true
		v.Design = &protos.Design{}
		if v.Design.Name == "" {
			v.Design.Name = "Untitled Design"
		}
		log.Println("Using template: ", templateName)
		// hxgeturl := fmt.Sprintf("/views/designs/MDEditorView?name=%s&description=%s", v.Design.Name, v.Design.Description)
	} else {
		client, _ := vc.ClientMgr.GetDesignSvcClient()
		resp, err := client.GetDesign(context.Background(), &protos.GetDesignRequest{
			Id: v.DesignId,
		})
		if err != nil {
			log.Println("Error getting design: ", err)
			return err, false
		}

		v.IsOwner = loggedInUserId == resp.Design.OwnerId
		log.Println("LoggedUser: ", loggedInUserId, resp.Design.OwnerId)

		if !v.IsOwner {
			log.Println("DesignEditor is NOT the owner.  Redirecting to view page...")
			if loggedInUserId == "" {
				http.Redirect(w, r, fmt.Sprintf("/login?callbackURL=%s", fmt.Sprintf("/designs/%s/edit", v.DesignId)), http.StatusSeeOther)
			} else {
				http.Redirect(w, r, fmt.Sprintf("/designs/%s/view", v.DesignId), http.StatusSeeOther)
			}
			return nil, true
		}

		v.Design = resp.Design
	}
	return
}
