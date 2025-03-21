package web

import (
	"os"

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

var dpUtils = DrawingService{
	ContentRoot:     "./content",
	CaseStudiesRoot: "casestudies",
}

func init() {
	if os.Getenv("APP_ENV") != "production" {
		site.CommonFuncMap = dpUtils.TemplateFunctions()
		// site.NewViewFunc = NewView
		site.Watch()
	}
}
