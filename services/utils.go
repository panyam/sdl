package services

import (
	"fmt"
	"strings"
	"time"
)

// Helper function to generate a random section title (similar to randomDesignName)
// You might want to move this to a utils file later.
func randomSectionTitle(sectionType string) string {
	// TODO: Implement more sophisticated random title generation based on type
	switch sectionType {
	case "text":
		return fmt.Sprintf("New Text Section %d", time.Now().UnixNano()%1000)
	case "drawing":
		return fmt.Sprintf("New Drawing %d", time.Now().UnixNano()%1000)
	default:
		return fmt.Sprintf("New %s Section %d", strings.Title(sectionType), time.Now().UnixNano()%1000)
	}
}
