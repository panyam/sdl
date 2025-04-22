package views

import (
	"context"
	"log"
	"net/http"

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
)

type DesignListView struct {
	Designs   []*protos.Design
	Paginator Paginator
}

func (g *DesignListView) Copy() View { return &DesignListView{} }

func (p *DesignListView) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	userId := vc.AuthMiddleware.GetLoggedInUserId(r)

	// if we are an independent view then read its params from the query params
	// otherwise those will be passed in
	_, _ = p.Paginator.Load(r, w, vc)

	client, _ := vc.ClientMgr.GetDesignSvcClient()

	req := protos.ListDesignsRequest{
		Pagination: &protos.Pagination{
			PageOffset: int32(p.Paginator.CurrentPage * p.Paginator.PageSize),
			PageSize:   int32(p.Paginator.PageSize),
		},
		OwnerId: userId,
		// CollectionId: p.CollectionId,
	}
	resp, err := client.ListDesigns(context.Background(), &req)
	if err != nil {
		log.Println("error getting notations: ", err)
		return err, false
	}
	log.Println("Found Designs: ", resp.Designs)
	p.Designs = resp.Designs
	p.Paginator.HasPrevPage = p.Paginator.CurrentPage > 0
	p.Paginator.HasNextPage = resp.Pagination.HasMore
	p.Paginator.EvalPages(p.Paginator.CurrentPage*p.Paginator.PageSize + int(resp.Pagination.TotalResults))
	return nil, false
}
