package console

import (
	"github.com/panyam/templar"
)

// SetupTemplates initializes the Templar template group
func SetupTemplates(templatesDir string) (*templar.TemplateGroup, error) {
	// Create a new template group
	group := templar.NewTemplateGroup()
	
	// Set up the file system loader with multiple paths
	group.Loader = templar.NewFileSystemLoader(
		templatesDir,
		templatesDir+"/shared",
		templatesDir+"/components",
	)
	
	// Preload common templates to ensure they're available
	commonTemplates := []string{
		"base.html",
		"systems/listing.html",
		"systems/details.html",
	}
	
	for _, tmpl := range commonTemplates {
		// Use defer to catch panics from MustLoad
		func() {
			defer func() {
				if r := recover(); r != nil {
					Debug("Template not found (will create): %s", tmpl)
				}
			}()
			group.MustLoad(tmpl, "")
		}()
	}
	
	return group, nil
}