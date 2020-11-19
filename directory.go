package gosrc

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/xaionaro-go/unsafetools"
)

type Directory struct {
	FileSet  *token.FileSet
	Packages Packages
}

func normalizePkgPath(
	path string,
	lookupPaths []string,
) (pkgPath string, dirPath string, err error) {
	defer func() {
		if err != nil {
			return
		}

		var st os.FileInfo
		st, err = os.Stat(dirPath)
		if err != nil {
			err = fmt.Errorf("unable to stat() on path '%s': %w", dirPath, err)
			return
		}

		if !st.IsDir() {
			pkgPath = filepath.Base(pkgPath)
			dirPath = filepath.Base(dirPath)
		}
	}()
	if filepath.IsAbs(path) {
		return path, path, nil
	}
	parts := strings.Split(path, string(filepath.Separator))
	wd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("unable to get workdir: %w", err)
	}
	for _, lookupPath := range lookupPaths {
		if !strings.HasPrefix(wd, lookupPath) {
			continue
		}
		parts[0] = wd[len(lookupPath):]
		if parts[0] == "" {
			parts = parts[1:]
		}
		pkgPath := strings.Join(parts, string(filepath.Separator))
		pkgPath = strings.Trim(pkgPath, string(filepath.Separator))
		return pkgPath, filepath.Join(lookupPath, pkgPath), nil
	}

	for _, lookupPath := range lookupPaths {
		dirPath := filepath.Join(lookupPath, path)
		if _, err := os.Stat(dirPath); err == nil {
			return path, dirPath, nil
		}
	}

	return "", "", fmt.Errorf("unable to find directory '%s' in paths %v", path, lookupPaths)
}

func OpenDirectoryByPkgPath(
	buildCtx *build.Context,
	pkgPath string,
	includeTestPkg bool,
	onlyFiles bool,
	externalImporter Importer,
) (*Directory, error) {
	var lookupPath string
	pkgPath, dirPath, err := normalizePkgPath(pkgPath, buildCtx.SrcDirs())
	if err != nil {
		return nil, fmt.Errorf("unable to normalize pkg path '%s': %w", pkgPath, err)
	}

	directory := &Directory{FileSet: token.NewFileSet()}

	if dirPath == `` {
		if externalImporter != nil { // TODO: remove this hack
			pkg, err := externalImporter.Import(pkgPath)
			if err != nil {
				return nil, fmt.Errorf("unable to import '%s': %w",
					pkgPath, err)
			}
			directory.Packages = append(directory.Packages, pkg)
			return directory, nil
		}

		return nil, ErrPackageNotFound{
			GoPath:      pkgPath,
			LookupPaths: buildCtx.SrcDirs(),
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
		conf = types.Config{Importer: importer.ForCompiler(token.NewFileSet(), "source", nil)}
		*(unsafetools.FieldByName(conf.Importer, "ctxt").(**build.Context)) = buildCtx
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
		for _, file := range pkgFiles.FilterByBuildTags(buildCtx.BuildTags) {
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
