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

// Directory contains multiple Packages
type Directory struct {
	FileSet  *token.FileSet
	Packages Packages
}

func normalizePkgPath(
	path string,
	lookupPaths []string,
) (pkgPath, dirPath, lookupPath string, err error) {
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
		return path, path, "", nil
	}

	parts := strings.Split(path, string(filepath.Separator))
	wd, err := os.Getwd()
	if err != nil {
		return "", "", "", fmt.Errorf("unable to get workdir: %w", err)
	}

	if parts[0] == "." {
		parts[0] = wd
		dirPath = filepath.Join(parts...)
		for _, lookupPath = range lookupPaths {
			if !strings.HasPrefix(wd, lookupPath) {
				continue
			}
			pkgPath = strings.Trim(dirPath[len(lookupPath):], string(filepath.Separator))
			return
		}
	}

	for _, lookupPath := range lookupPaths {
		dirPath := filepath.Join(lookupPath, path)
		if _, err := os.Stat(dirPath); err == nil {
			return path, dirPath, lookupPath, nil
		}
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
		return pkgPath, filepath.Join(lookupPath, pkgPath), lookupPath, nil
	}

	return "", "", "", fmt.Errorf("unable to find directory '%s' in paths %v", path, lookupPaths)
}

// OpenDirectoryByPkgPath finds a real directory using Go's pkg path,
// scans it for source code files, parses them and returns an instance of
// Directory (which contains everything inside).
func OpenDirectoryByPkgPath(
	buildCtx *build.Context,
	pkgPath string,
	includeTestFiles bool,
	includeTestPkg bool,
	onlyFiles bool,
	externalImporter Importer,
) (*Directory, error) {
	if !includeTestFiles {
		includeTestPkg = false
	}

	lookupPaths := buildCtx.SrcDirs()
	if len(lookupPaths) == 0 {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("unable to determine the homedir of the user: %w", err)
		}
		lookupPaths = append(lookupPaths, filepath.Join(homeDir, "go", "src"))
	}

	pkgPath, dirPath, lookupPath, err := normalizePkgPath(pkgPath, lookupPaths)
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
		conf = types.Config{Importer: importer.ForCompiler(token.NewFileSet(), "source", nil)}
		// Unfortunately, I haven't found another way to set the context of this importer:
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
			if !includeTestFiles && strings.HasSuffix(file.Path, `_test.go`) {
				continue
			}
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
