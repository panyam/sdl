package views

import (
	"log"
	"net/http"
	"strconv"
)

type Paginator struct {
	CurrentPage int
	PageSize    int
	HasPrevPage bool
	HasNextPage bool
	TotalCount  int
	Pages       []int
}

func (v *Paginator) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	queryParams := r.URL.Query()
	v.CurrentPage = 0
	v.PageSize = 20
	v.PageSize, err = strconv.Atoi(queryParams.Get("paginatorPageSize"))
	if err != nil {
		v.PageSize = 20
	}
	v.CurrentPage, err = strconv.Atoi(queryParams.Get("paginatorCurrentPage"))
	if err != nil {
		v.CurrentPage = 0
	}
	v.EvalPages(10000)
	return
}

func (v *Paginator) EvalPages(total int) {
	if v.CurrentPage*v.PageSize > total {
		total = v.CurrentPage * v.PageSize
	}
	v.TotalCount = total
	maxPages := total / v.PageSize
	if total%v.PageSize != 0 {
		maxPages += 1
	}
	if v.HasNextPage {
		maxPages += 1
	}
	v.Pages = nil
	for i := -5; i < 5; i++ {
		currPage := i + v.CurrentPage
		if currPage >= 0 && currPage <= maxPages {
			v.Pages = append(v.Pages, currPage)
		}
	}
	if len(v.Pages) > 3 {
		v.Pages = v.Pages[len(v.Pages)-3:]
	}
	log.Println("PageSize, Total, MaxPages, Pages: ", v.PageSize, total, maxPages, v.Pages)
}
