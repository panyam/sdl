package web

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type Drawing struct {
	Id        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Title     string
	Format    string
	Editor    string
	Contents  any
}

// A handler for serving system design case studies along with ability to
// show them in certain easily consumeable ways as well as persistence for excalidraw

// A single case study handler for a case study hosted at a particular folder.
// For example we could serve case studies with:
// /casestudies/doordash/order_management_system/
//
// This file must have atleast one file - index.md (or index.mdx)
// Our index.mdx is special in that it contains our system design with extra annotations, eg:
//
//  1. for excalidraw images in view and edit mode + endpoint to save excalidraw images locally
//
//  2. "Notes" - Where we can add custom Notes to sections that will show up on the "right"
//     side bar so as to not loose focus from main content
//
//  3. Ability collapse/expand certain sections
//
//  4. A left bar showing table of contents.
//
//     Over time we will also add simulation/SDL components to view graphs and dynamic behavior of our systems.
type CaseStudy struct {
	// Root folder where the case study is hosted
	RootFolder string

	mux *http.ServeMux
}

func NewCaseStudy(rootFolder string) *CaseStudy {
	out := &CaseStudy{
		RootFolder: rootFolder,
		mux:        http.NewServeMux(),
	}

	out.setupRoutes()
	return out
}

func (c *CaseStudy) Handler() http.Handler {
	return c.mux
}

func (c *CaseStudy) setupRoutes() {
	// Allow the following handlers
	// GET /drawings/<id>			-		Gets the content of a particular drawing by ID
	// POST /drawings/<id>		-		Creates a new drawing by ID
	// PUT /drawings/<id>			-		Updates a drawing ID
	// /static								- 	Static files/folders in the case study
	// /											- 	Load/Serve/Process the index.md file with our special processing
	staticFolder, err := filepath.Abs(filepath.Join(c.RootFolder, "/static"))
	log.Println("SF: ", staticFolder, err)
	c.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticFolder))))
	// For handling drawings
	c.mux.Handle("/drawings/", http.StripPrefix("/drawings", c.setupDrawingsMux()))
	c.mux.HandleFunc("/", c.ServeCaseStudy)
	c.mux.Handle("/{invalidbits}", http.NotFoundHandler())
}

func (c *CaseStudy) ServeCaseStudy(w http.ResponseWriter, r *http.Request) {
	extensions := []string{"md", "mdx", "html", "htm"}
	var indexPath string
	var err error
	for _, ext := range extensions {
		indexPath, err = filepath.Abs(filepath.Join(c.RootFolder, fmt.Sprintf("index.%s", ext)))
		if err == nil {
			break
		}
		// serve this
		log.Println("ip, err: ", indexPath, err)
	}

	if err != nil {
		log.Println("Error with ext: ", err)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "Invalid Case Study")
		return
	}

	// Serve it
	if strings.HasSuffix(indexPath, ".html") || strings.HasSuffix(indexPath, ".htm") {
		// serve it as is
		http.ServeFile(w, r, indexPath)
	} else {
		c.ServeMDX(w, r, indexPath)
		// render as md
	}
}

func (c *CaseStudy) ServeMDX(w http.ResponseWriter, r *http.Request, path string) {
	fmt.Fprintf(w, "Hello world 333")
}

func (c *CaseStudy) PathForDrawingId(drawingId string) string {
	fullPath, err := filepath.Abs(filepath.Join(c.RootFolder, "drawings", fmt.Sprintf("%s.drawing", drawingId)))
	log.Println("Full Drawing Path: ", drawingId, fullPath)
	if err != nil {
		log.Println("Error: ", err)
		return ""
	}
	return fullPath
}

func (c *CaseStudy) setupDrawingsMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/{drawingId}", func(w http.ResponseWriter, r *http.Request) {
		// Otherwise serve the file
		drawingId := r.PathValue("drawingId")
		drawingPath := c.PathForDrawingId(drawingId)
		if drawingPath == "" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "Invalid Drawing")
		}
		if r.Method == "GET" {
			http.ServeFile(w, r, drawingPath)
		} else {
			// We udpate it
		}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			if true {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintln(w, "Listing not allowed")
				return
			}
		} else {
			fmt.Fprintf(w, "Creating drawings")
		}
	})
	return mux
}
