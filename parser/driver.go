package parser

// ParseFile is a convenience function to parse directly from a file path.
/* // Uncomment if needed
import "os"
func ParseFile(filePath string) (*File, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, fmt.Errorf("error opening file %s: %w", filePath, err)
    }
    defer file.Close()
    return Parse(file)
}
*/
