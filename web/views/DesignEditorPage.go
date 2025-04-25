package views

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
)

type DesignEditorPage struct {
	BasePage
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
	v.DesignId = r.PathValue("designId")
	// queryParams := r.URL.Query()
	// templateName := queryParams.Get("template")
	loggedInUserId := vc.AuthMiddleware.GetLoggedInUserId(r)

	slog.Info("Loading editor for design with ID: ", "nid", v.DesignId)

	client, _ := vc.ClientMgr.GetDesignSvcClient()
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

		// create a new design
		ctx := vc.ClientMgr.ClientContext(nil, loggedInUserId)
		v.Design = &protos.Design{
			Name: randomDesignName(),
		}
		resp, err := client.CreateDesign(ctx, &protos.CreateDesignRequest{
			Design: v.Design,
		})
		if err != nil {
			log.Println("Error creating design: ", err)
			return err, false
		}
		http.Redirect(w, r, fmt.Sprintf("/designs/%s/edit", resp.Design.Id), http.StatusFound)
		// hxgeturl := fmt.Sprintf("/views/designs/MDEditorView?name=%s&description=%s", v.Design.Name, v.Design.Description)
	} else {
		ctx := vc.ClientMgr.ClientContext(nil, loggedInUserId)
		resp, err := client.GetDesign(ctx, &protos.GetDesignRequest{
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
