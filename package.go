package gosrc

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/types"
	"path/filepath"
	"strings"
)

type Package struct {
	*types.Package

	Name       string
	LookupPath string
	DirPath    string
	Info       *types.Info
	Files      Files
}
type Packages []*Package

func (pkg *Package) String() string {
	return fmt.Sprintf("%+v", *pkg)
}

func (pkg *Package) toType(expr ast.Expr) (types.TypeAndValue, error) {
	if pkg == nil {
		return types.TypeAndValue{}, fmt.Errorf("got: Package == nil")
	}
	return pkg.Info.Types[expr], nil
}

type Importer interface {
	Import(goPath string) (*Package, error)
}

func (pkg *Package) Imports(buildCtx *build.Context, onlyFiles bool, externalImporter Importer) (Packages, error) {
	var result Packages

	for _, _import := range pkg.Package.Imports() {
		dirPath := _import.Path()
		dir, err := OpenDirectoryByPkgPath(buildCtx, dirPath, false, onlyFiles, externalImporter)
		if err != nil {
			return nil, fmt.Errorf("unable to scan directory '%s': %w", dirPath, err)
		}

		for _, pkg := range dir.Packages {
			if strings.HasSuffix(pkg.Name, `_test`) {
				continue
			}
			result = append(result, pkg)
		}
	}

	return result, nil
}

func (pkg Package) Path() string {
	if pkg.Package != nil {
		return pkg.Package.Path()
	}
	return strings.Trim(pkg.DirPath[len(pkg.LookupPath):], string(filepath.Separator))
}
