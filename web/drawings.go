package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	gotl "github.com/panyam/templar"
)

type DrawingPathUtils struct {
	ContentRoot     string
	CaseStudiesRoot string
}

func (d *DrawingPathUtils) SaveDrawing(caseStudyId, drawingId, format string, body []byte) (err error) {
	drawingPath, _, err := d.PathForDrawingId(caseStudyId, drawingId, true, format)
	if err == nil {
		log.Printf("Saving format (%s) -> %s", format, drawingPath)
		err = os.WriteFile(drawingPath, body, 0666)
	}
	if err != nil {
		// TODO - quit here or do someting else?
		log.Println("Could not write to file: ", drawingPath, err)
	}
	return
}

func (d *DrawingPathUtils) FolderForDrawingId(caseStudyId, drawingId string, ensure bool) (folderPath string, exists bool, err error) {
	folderPath, err = filepath.Abs(filepath.Join(d.ContentRoot, d.CaseStudiesRoot, caseStudyId, "drawings", drawingId))
	if err != nil {
		return
	}

	// check if fodler also exists
	if ensure {
		_, err := os.Stat(folderPath)
		if err != nil && os.IsNotExist(err) {
			err = os.MkdirAll(folderPath, os.ModePerm)
		}
		exists = err == nil
	}
	return
}

func (d *DrawingPathUtils) PathForDrawingId(caseStudyId, drawingId string, ensure bool, extension string) (fullPath string, exists bool, err error) {
	folderPath, exists, err := d.FolderForDrawingId(caseStudyId, drawingId, ensure)

	fullPath, err = filepath.Abs(filepath.Join(folderPath, fmt.Sprintf("contents.%s", extension)))
	log.Println("Full Drawing Path: ", drawingId, extension, fullPath, err)
	if err != nil {
		log.Println("Error accessing path: ", fullPath, err)
		return fullPath, false, err
	}

	if _, err := os.Stat(fullPath); err == nil {
		exists = true
	} else if os.IsNotExist(err) && ensure {
		// create an empty file
		err = os.WriteFile(fullPath, []byte(""), os.ModePerm)
	}

	return
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
	DrawingPathUtils

	Templates *gotl.TemplateGroup

	mux *http.ServeMux
}

func NewDrawingApi(contentRoot string) *DrawingApi {
	out := &DrawingApi{
		DrawingPathUtils: DrawingPathUtils{
			ContentRoot:     contentRoot,
			CaseStudiesRoot: "casestudies",
		},
		mux: http.NewServeMux(),
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
	c.mux.HandleFunc("/{caseStudyId}/", func(w http.ResponseWriter, r *http.Request) {
		// Otherwise serve the file
		caseStudyId := r.PathValue("caseStudyId")
		reqPath := r.URL.Path[1+len(caseStudyId):]
		log.Println("caseStudyId: ", caseStudyId)
		log.Println("drawingId: ", reqPath)
		if r.Method == "GET" {
			c.ServeDrawing(caseStudyId, reqPath, w, r)
		} else {
			// We udpate it
			c.UpdateDrawing(caseStudyId, reqPath, w, r)
		}
	})
}

func (c *DrawingApi) ServeDrawing(caseStudyId, drawingId string, w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}
	drawingPath, exists, err := c.PathForDrawingId(caseStudyId, drawingId, false, format)
	log.Println("drawingPath: ", drawingPath, err)
	if drawingPath == "" || !exists || err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, `{"error": "Invalid Drawing"}`)
		return
	}

	if format == "json" {
		w.Header().Set("Content-Type", "application/json")
	}
	http.ServeFile(w, r, drawingPath)
}

func (c *DrawingApi) UpdateDrawing(caseStudyId, drawingId string, w http.ResponseWriter, r *http.Request) {
	log.Println("Updating drawing...")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, fmt.Sprintf("can't read body: %v", err), http.StatusBadRequest)
		return
	}

	var payload map[string]any
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Printf("Error parsing body: %v", err)
		http.Error(w, fmt.Sprintf("can't parse body: %v", err), http.StatusBadRequest)
		return
	}

	if payload["formats"] != nil {
		formats, ok := payload["formats"].(map[string]any)
		if ok {
			for fmt, body := range formats {
				c.SaveDrawing(caseStudyId, drawingId, fmt, []byte(body.(string)))
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"success": true}`)
			return
		}
		log.Println("Not Ok: ")
	}

	log.Println("Did not find formats or error: ", payload)
	http.Error(w, fmt.Sprintf("can't parse body: %v", err), http.StatusBadRequest)
}
