package views

import (
	"fmt"
	"net/http"
)

// This should be mirroring how we are setting up our app.yaml
func (n *LCViews) setupRoutes() {
	n.mux.Handle("/views/", http.StripPrefix("/views", n.setupViewsMux()))
	n.mux.Handle("/designs/", http.StripPrefix("/designs", n.setupDesignsMux()))

	n.mux.HandleFunc("/login", n.ViewRenderer(Copier(&LoginPage{}), ""))
	// n.mux.HandleFunc("/logout", n.onLogout)
	n.mux.HandleFunc("/privacy-policy", n.ViewRenderer(Copier(&PrivacyPolicy{}), ""))
	n.mux.HandleFunc("/terms-of-service", n.ViewRenderer(Copier(&TermsOfService{}), ""))
	n.mux.HandleFunc("/browse", n.ViewRenderer(Copier(&BrowsePage{}), ""))
	n.mux.HandleFunc("/", n.ViewRenderer(Copier(&HomePage{}), ""))
	n.mux.Handle("/{invalidbits}/", http.NotFoundHandler()) // <-- Default 404
}

func (n *LCViews) setupViewsMux() *http.ServeMux {
	mux := http.NewServeMux()

	// n.HandleView(Copier(&components.SelectTemplatePage{}), r, w)
	return mux
}

func (n *LCViews) setupDesignsMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/new", n.ViewRenderer(Copier(&DesignEditorPage{}), ""))
	mux.HandleFunc("/{designId}/view", n.ViewRenderer(Copier(&DesignViewerPage{}), ""))

	mux.HandleFunc("/{designId}/edit", n.ViewRenderer(Copier(&DesignEditorPage{}), ""))
	mux.HandleFunc("/{designId}/copy", func(w http.ResponseWriter, r *http.Request) {
		designId := r.PathValue("designId")
		http.Redirect(w, r, fmt.Sprintf("/designs/new?copyFrom=%s", designId), http.StatusFound)
	}) // .Methods("GET")
	mux.HandleFunc("/{designId}", n.ViewRenderer(Copier(&DesignPage{}), ""))
	return mux
}
