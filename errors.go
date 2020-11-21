package gosrc

import (
	"fmt"
)

// ErrPackageNotFound is returned when was unable to perform an operation
// due to inability to find the Go package.
type ErrPackageNotFound struct {
	GoPath      string
	LookupPaths []string
}

// Error implements error
func (err ErrPackageNotFound) Error() string {
	return fmt.Sprintf("unable to find package with path '%s' in %s",
		err.GoPath, err.LookupPaths)
}
