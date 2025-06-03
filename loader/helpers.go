package loader

import "log"

func (l *Loader) LoadFilesAndValidate(sourceFiles ...string) (success bool) {
	success = true
	for _, f := range sourceFiles {
		fs, err := l.LoadFile(f, "", 0)
		if err != nil {
			log.Println("Error loading file: ", f, err)
			success = false
			continue
		}
		l.Validate(fs)
		if fs.HasErrors() {
			success = false
			log.Printf("\nError Validating File %s\n", fs.FullPath)
			fs.PrintErrors()
		} else {
			log.Printf("\nFile %s - Validated Successfully at: %v\n", fs.FullPath, fs.LastValidated)
		}
	}
	return
}
