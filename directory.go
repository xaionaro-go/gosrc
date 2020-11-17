package gosrc

import (
	"errors"
	"fmt"
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type Directory struct {
	FileSet  *token.FileSet
	Packages Packages
}

func OpenDirectoryByGoPath(
	lookupPaths []string,
	goPath string,
	includeTestPkg bool,
	onlyFiles bool,
	externalImporter Importer,
) (*Directory, error) {
	var dirPath string
	var pkgPath string
	var lookupPath string

	for _, lookupPath = range lookupPaths {
		var err error
		candidate := filepath.Join(lookupPath, goPath)
		dirPath, err = closestDir(candidate)
		var pathError *os.PathError
		if errors.As(err, &pathError) && pathError.Err == syscall.ENOENT {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("unable to find directory of '%s': %w", candidate, err)
		}
		if lookupPath == "." {
			lookupPath = ""
		}
		pkgPath = strings.Trim(dirPath[len(lookupPath):], string(filepath.Separator))
		break
	}

	directory := &Directory{FileSet: token.NewFileSet()}

	if dirPath == `` {
		if externalImporter != nil { // TODO: remove this hack
			pkg, err := externalImporter.Import(goPath)
			if err != nil {
				return nil, fmt.Errorf("unable to import '%s': %w",
					goPath, err)
			}
			directory.Packages = append(directory.Packages, pkg)
			return directory, nil
		}

		return nil, ErrPackageNotFound{
			GoPath:      goPath,
			LookupPaths: lookupPaths,
		}
	}

	files, err := scanForFiles(directory.FileSet, dirPath, false)
	if err != nil {
		return nil, fmt.Errorf("unable to open package at '%s': %w", dirPath, err)
	}
	pkgFilesMap := map[string]Files{}
	for _, file := range files {
		pkgFilesMap[file.PackageName()] = append(pkgFilesMap[file.PackageName()], file)
	}

	var conf types.Config
	var pkgRaw *types.Package
	if !onlyFiles {
		conf = types.Config{Importer: importer.For("source", nil)}
		pkgRaw, err = conf.Importer.Import(pkgPath)
		if err != nil {
			return nil, fmt.Errorf("unable to import package '%s' (in: '%s'): %w", pkgPath, dirPath, err)
		}
	}

	for pkgName, pkgFiles := range pkgFilesMap {
		if strings.HasSuffix(pkgName, `_test`) && !includeTestPkg {
			continue
		}
		pkg := &Package{
			Name:       pkgName,
			DirPath:    dirPath,
			LookupPath: lookupPath,
			Package:    pkgRaw,
			Files:      pkgFiles,
		}

		var fileAsts []*ast.File
		for _, file := range pkgFiles {
			file.Package = pkg
			fileAsts = append(fileAsts, file.Ast)
		}

		if !onlyFiles {
			info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}
			if _, err := conf.Check(dirPath, directory.FileSet, fileAsts, info); err != nil {
				return nil, fmt.Errorf("unable to get package info: %w", err)
			}
			pkg.Info = info
		}

		directory.Packages = append(directory.Packages, pkg)
	}

	return directory, nil
}
