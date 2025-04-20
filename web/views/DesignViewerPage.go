package views

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
)

type DesignViewerPage struct {
	Header   Header
	IsOwner  bool
	DesignId string
	Design   *protos.Design
}

func (g *DesignViewerPage) Copy() View { return &DesignViewerPage{} }

func (v *DesignViewerPage) SetupDefaults() {
}

func (v *DesignViewerPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	v.Header.Load(r, w, vc)
	v.SetupDefaults()
	v.DesignId = r.PathValue("designId")
	slog.Info("Loading design with ID: ", "nid", v.DesignId)
	client, _ := vc.ClientMgr.GetDesignSvcClient()
	resp, err := client.GetDesign(context.Background(), &protos.GetDesignRequest{
		Id: v.DesignId,
	})
	if err != nil {
		log.Println("Error getting design: ", err)
		// c.HTML(404, "Design not found", nil)
		return err, false
	}

	currOwnerId := vc.AuthMiddleware.GetLoggedInUserId(r)
	v.IsOwner = currOwnerId == resp.Design.OwnerId

	v.Design = resp.Design

	/*
		if v.IsOwner {
			v.Header.RightMenuItems = append(v.Header.RightMenuItems, HeaderMenuItem{Title: "Delete", Link: fmt.Sprintf("/designs/%s/delete", v.DesignId)})
		}
	*/
	return
}
