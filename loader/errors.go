package loader

import (
	"fmt"
	"os"
)

type ErrorCollector struct {
	// Errors for this file
	Errors []error

	// Max errors before we panic
	// 0 => no limit
	MaxErrors int
}

func (f *ErrorCollector) HasErrors() bool {
	return len(f.Errors) > 0
}

func (f *ErrorCollector) PrintErrors() {
	for _, err := range f.Errors {
		fmt.Fprintln(os.Stderr, err)
	}
}

func (i *ErrorCollector) AddErrors(errs ...error) {
	for _, err := range errs {
		// TODO - Remove after testing - right now we fail on first error to detect and fix with call stack
		i.Errors = append(i.Errors, err)
		if i.MaxErrors > 0 && len(i.Errors) >= i.MaxErrors {
			panic(err)
		}
	}
}

func (i *ErrorCollector) Errorf(pos Location, format string, args ...any) bool {
	i.AddErrors(InfErrorf(pos, format, args...))
	return false
}
