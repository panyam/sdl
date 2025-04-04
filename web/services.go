package web

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type DrawingService struct {
	ContentRoot     string
	OutputRoot      string
	CaseStudiesRoot string
}

func (d *DrawingService) SaveDrawing(caseStudyId, drawingId, format string, body []byte) (err error) {
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

func (d *DrawingService) FolderForDrawingId(caseStudyId, drawingId string, ensure bool) (folderPath string, exists bool, err error) {
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

func (d *DrawingService) PathForDrawingId(caseStudyId, drawingId string, ensure bool, extension string) (fullPath string, exists bool, err error) {
	folderPath, exists, err := d.FolderForDrawingId(caseStudyId, drawingId, ensure)

	fullPath, err = filepath.Abs(filepath.Join(folderPath, fmt.Sprintf("contents.%s", extension)))
	// log.Println("Full Drawing Path: ", drawingId, extension, fullPath, err)
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

func (d *DrawingService) EditorUrl(caseStudyId, drawingId string) (out string) {
	return fmt.Sprintf("/drawings/%s/%s/editor", caseStudyId, drawingId)
}

func (d *DrawingService) PreviewExists(caseStudyId, drawingId, extension string) bool {
	_, exists, err := dpUtils.PathForDrawingId(caseStudyId, drawingId, false, extension)
	//log.Println("Path: ", path, exists)
	return exists && err == nil
}

func (d *DrawingService) PreviewUrl(caseStudyId, drawingId, extension string) (out string, err error) {
	// return filepath.Abs(filepath.Join(site.OutputDir, fmt.Sprintf("%s.svg", drawingId)))
	out, _, err = dpUtils.PathForDrawingId(caseStudyId, drawingId, false, extension)
	out = strings.Replace(out, site.ContentRoot, "", 1)
	return
}

func (d *DrawingService) DrawingData(caseStudyId, drawingId string) (out string, err error) {
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

func (d *DrawingService) TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"DrawingData":          d.DrawingData,
		"DrawingPreviewUrl":    d.PreviewUrl,
		"DrawingPreviewExists": d.PreviewExists,
		"DrawingEditorUrl":     d.EditorUrl,
	}
}
