package web

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	gotl "github.com/panyam/templar"
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
type DrawingApi struct {
	// Root folder where the case study is hosted
	ContentRoot string

	Templates *gotl.TemplateGroup

	mux *http.ServeMux
}

func NewDrawingApi(contentRoot string) *DrawingApi {
	out := &DrawingApi{
		ContentRoot: contentRoot,
		mux:         http.NewServeMux(),
	}

	out.setupRoutes()
	return out
}

func (c *DrawingApi) Handler() http.Handler {
	return c.mux
}

func (c *DrawingApi) setupRoutes() {
	// Allow the following handlers
	// GET /drawings/<id>			-		Gets the content of a particular drawing by ID
	// POST /drawings/<id>		-		Creates a new drawing by ID
	// PUT /drawings/<id>			-		Updates a drawing ID
	c.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Otherwise serve the file
		drawingId := r.PathValue("drawingId")
		reqPath := r.URL.Path[1:]
		drawingPath, exists := c.PathForDrawingId(reqPath)
		log.Println("drawingId: ", drawingId)
		log.Println("reqPath: ", reqPath)
		log.Println("drawingPath: ", drawingPath)
		if drawingPath == "" || !exists {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "Invalid Drawing")
		}
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			http.ServeFile(w, r, drawingPath)
		} else {
			// We udpate it
			log.Println("Updating drawing...")
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("Error reading body: %v", err)
				http.Error(w, "can't read body", http.StatusBadRequest)
				return
			}

			err = os.WriteFile(drawingPath, body, 0666)
		}
	})
}

func (c *DrawingApi) PathForDrawingId(drawingId string) (fullPath string, exists bool) {
	fullPath, err := filepath.Abs(filepath.Join(c.ContentRoot, fmt.Sprintf("%s.drawing", drawingId)))
	log.Println("Full Drawing Path: ", drawingId, fullPath)
	if err != nil {
		log.Println("Error: ", err)
		return "", false
	}
	if _, err := os.Stat(fullPath); err == nil {
		exists = true
	} else if os.IsNotExist(err) {
		exists = false
	}
	return
}
