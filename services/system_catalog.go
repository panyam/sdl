package services

import (
	"os"
	"path/filepath"
	"time"
)

// SystemInfo represents a system in the catalog
type SystemInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Difficulty  string   `json:"difficulty"`
	Tags        []string `json:"tags"`
	Icon        string   `json:"icon,omitempty"`
	LastUpdated string   `json:"lastUpdated"`
}

// SystemProject represents a full system project
type SystemProject struct {
	ID             string                   `json:"id"`
	Name           string                   `json:"name"`
	Description    string                   `json:"description"`
	Category       string                   `json:"category"`
	Difficulty     string                   `json:"difficulty"`
	Tags           []string                 `json:"tags"`
	Icon           string                   `json:"icon,omitempty"`
	Versions       map[string]SystemVersion `json:"versions"`
	DefaultVersion string                   `json:"defaultVersion"`
	LastUpdated    string                   `json:"lastUpdated"`
	// Internal file paths (not exposed in JSON)
	sdlFile    string
	recipeFile string
}

// SystemVersion represents a version of a system
type SystemVersion struct {
	SDL    string `json:"sdl"`
	Recipe string `json:"recipe"`
	Readme string `json:"readme,omitempty"`
}

// SystemCatalogService manages the system examples catalog
type SystemCatalogService struct {
	systems  map[string]*SystemProject
	basePath string // Base path for examples, e.g., "examples"
}

// NewSystemCatalogService creates a new system catalog service
func NewSystemCatalogService() *SystemCatalogService {
	service := &SystemCatalogService{
		systems:  make(map[string]*SystemProject),
		basePath: "examples", // Default to examples directory
	}
	service.initializeCatalog()
	return service
}

// initializeCatalog loads the example systems
func (s *SystemCatalogService) initializeCatalog() {
	// Define curated systems with their file paths
	s.addSystem(&SystemProject{
		ID:             "bitly",
		Name:           "Bitly URL Shortener",
		Description:    "A scalable URL shortening service with analytics and caching",
		Category:       "Web Services",
		Difficulty:     "beginner",
		Tags:           []string{"web", "database", "caching", "rest-api"},
		Icon:           "🔗",
		DefaultVersion: "v1",
		sdlFile:        "bitly/mvp.sdl",
		recipeFile:     "bitly/mvp.recipe",
	})

	s.addSystem(&SystemProject{
		ID:             "uber-basic",
		Name:           "Uber Ride Sharing (Basic)",
		Description:    "Simplified ride-sharing platform with driver matching",
		Category:       "Transportation",
		Difficulty:     "intermediate",
		Tags:           []string{"microservices", "real-time", "geo-spatial", "matching"},
		Icon:           "🚗",
		DefaultVersion: "v1",
		sdlFile:        "uber/mvp.sdl",
		recipeFile:     "uber/mvp.recipe",
	})

	s.addSystem(&SystemProject{
		ID:             "uber-intermediate",
		Name:           "Uber Ride Sharing (Intermediate)",
		Description:    "Enhanced ride-sharing with real-time tracking and pricing",
		Category:       "Transportation",
		Difficulty:     "intermediate",
		Tags:           []string{"microservices", "real-time", "geo-spatial", "matching", "pricing"},
		Icon:           "🚕",
		DefaultVersion: "v1",
		sdlFile:        "uber/intermediate.sdl",
		recipeFile:     "uber/intermediate.recipe",
	})

	s.addSystem(&SystemProject{
		ID:             "uber-advanced",
		Name:           "Uber Ride Sharing (Advanced)",
		Description:    "Full ride-sharing platform with surge pricing and machine learning",
		Category:       "Transportation",
		Difficulty:     "advanced",
		Tags:           []string{"microservices", "real-time", "geo-spatial", "pricing", "ml", "analytics"},
		Icon:           "🚖",
		DefaultVersion: "v1",
		sdlFile:        "uber/modern.sdl",
		recipeFile:     "uber/modern.recipe",
	})

}

// addSystem adds a system to the catalog and loads its content from files
func (s *SystemCatalogService) addSystem(project *SystemProject) {
	// Load SDL and recipe content from files
	sdlPath := filepath.Join(s.basePath, project.sdlFile)
	recipePath := filepath.Join(s.basePath, project.recipeFile)

	sdlContent := ""
	recipeContent := ""
	var lastMod time.Time

	// Read SDL file
	if sdlBytes, err := os.ReadFile(sdlPath); err == nil {
		sdlContent = string(sdlBytes)
		if info, err := os.Stat(sdlPath); err == nil {
			lastMod = info.ModTime()
		}
	}

	// Read Recipe file
	if recipeBytes, err := os.ReadFile(recipePath); err == nil {
		recipeContent = string(recipeBytes)
		if info, err := os.Stat(recipePath); err == nil {
			if info.ModTime().After(lastMod) {
				lastMod = info.ModTime()
			}
		}
	}

	// If files don't exist, use empty content and current time
	if lastMod.IsZero() {
		lastMod = time.Now()
	}

	// Create version with file content
	project.Versions = map[string]SystemVersion{
		"v1": {
			SDL:    sdlContent,
			Recipe: recipeContent,
		},
	}
	project.LastUpdated = lastMod.Format(time.RFC3339)

	s.systems[project.ID] = project
}

// ListSystems returns all systems as SystemInfo
func (s *SystemCatalogService) ListSystems() []SystemInfo {
	var systems []SystemInfo
	for _, project := range s.systems {
		systems = append(systems, SystemInfo{
			ID:          project.ID,
			Name:        project.Name,
			Description: project.Description,
			Category:    project.Category,
			Difficulty:  project.Difficulty,
			Tags:        project.Tags,
			Icon:        project.Icon,
			LastUpdated: project.LastUpdated,
		})
	}
	return systems
}

// GetSystem returns a specific system project
func (s *SystemCatalogService) GetSystem(id string) *SystemProject {
	return s.systems[id]
}
