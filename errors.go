package gosrc

import (
	"fmt"
)

type ErrPackageNotFound struct {
	GoPath      string
	LookupPaths []string
}

func (err ErrPackageNotFound) Error() string {
	return fmt.Sprintf("unable to find package with path '%s' in %s",
		err.GoPath, err.LookupPaths)
}
