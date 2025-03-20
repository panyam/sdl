package web

import (
	"fmt"
	"html/template"
	"os"
	"strings"

	s3 "github.com/panyam/s3gen"
)

var site = s3.Site{
	OutputDir:   "./output",
	ContentRoot: "./content",
	TemplateFolders: []string{
		"./web/templates",
	},
	StaticFolders: []string{
		"/static/", "web/static",
	},
	DefaultPageTemplate: s3.PageTemplate{
		Name:   "BasePage.html",
		Params: map[any]any{"BodyTemplateName": "BaseBody"},
	},
}

var dpUtils = DrawingPathUtils{
	ContentRoot:     "./content",
	CaseStudiesRoot: "casestudies",
}

func init() {
	if os.Getenv("APP_ENV") != "production" {
		site.CommonFuncMap = TemplateFunctions()
		// site.NewViewFunc = NewView
		site.Watch()
	}
}

// //////////// Functions for our site
func TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"DrawingData":          DrawingData,
		"DrawingPreviewUrl":    DrawingPreviewUrl,
		"DrawingPreviewExists": DrawingPreviewExists,
		"DrawingEditorUrl":     DrawingEditorUrl,
	}
}

func DrawingEditorUrl(caseStudyId, drawingId string) (out string) {
	return fmt.Sprintf("/drawings/%s/%s/editor", caseStudyId, drawingId)
}

func DrawingPreviewExists(caseStudyId, drawingId, extension string) bool {
	_, exists, err := dpUtils.PathForDrawingId(caseStudyId, drawingId, false, extension)
	return exists && err == nil
}

func DrawingPreviewUrl(caseStudyId, drawingId, extension string) (out string, err error) {
	// return filepath.Abs(filepath.Join(site.OutputDir, fmt.Sprintf("%s.svg", drawingId)))
	out, _, err = dpUtils.PathForDrawingId(caseStudyId, drawingId, false, extension)
	out = strings.Replace(out, site.ContentRoot, "", 1)
	return
}

func DrawingData(caseStudyId, drawingId string) (out string, err error) {
	filePath, _, err := dpUtils.PathForDrawingId(caseStudyId, drawingId, false, "json")
	if err != nil {
		contents, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Sprintf("%v", err), err
		}
		out = string(contents)
	}
	return
}
