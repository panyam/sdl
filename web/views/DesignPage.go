package views

import (
	"log"
	"net/http"

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
)

type DesignPage struct {
	BasePage
	Header        Header
	IsOwner       bool
	DesignId      string
	Design        *protos.Design
	Errors        map[string]string
	AllowCustomId bool
}

func (g *DesignPage) Copy() View { return &DesignPage{} }

func (v *DesignPage) SetupDefaults() {
}

func (v *DesignPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	v.Header.Load(r, w, vc)
	v.SetupDefaults()
	v.DesignId = r.PathValue("designId")
	// queryParams := r.URL.Query()
	// templateName := queryParams.Get("template")
	loggedInUserId := vc.AuthMiddleware.GetLoggedInUserId(r)
	ctx := vc.ClientMgr.ClientContext(nil, loggedInUserId)
	log.Println("Deleting Design, method: ", r.Method, loggedInUserId)
	if r.Method == "DELETE" {
		client, _ := vc.ClientMgr.GetDesignSvcClient()
		_, err := client.DeleteDesign(ctx, &protos.DeleteDesignRequest{
			Id: v.DesignId,
		})
		if err != nil {
			log.Println("Error getting design: ", err)
			http.Redirect(w, r, "/", http.StatusFound)
			return err, false
		}
		http.Redirect(w, r, "/", http.StatusFound)
		return nil, true
	}

	log.Println("=============")
	log.Println("Catch all - should not be coming here", r.Header)
	log.Println("=============")
	http.Redirect(w, r, "/", http.StatusFound)
	return nil, true
}
