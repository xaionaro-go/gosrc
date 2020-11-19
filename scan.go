package gosrc

import (
	"fmt"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func scanForFiles(fileSet *token.FileSet, dirPath string, isRecursive bool) (Files, error) {
	var goFiles Files

	stat, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open '%s': %w", dirPath, err)
	}
	if !stat.IsDir() {
		return scanForFiles(fileSet, path.Dir(dirPath), isRecursive)
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open '%s' as dir: %w", dirPath, err)
	}

	for _, file := range files {
		path := filepath.Join(dirPath, file.Name())
		switch {
		case file.IsDir():
			if !isRecursive {
				continue
			}

			additionalFiles, err := scanForFiles(fileSet, path, isRecursive)
			if err != nil {
				return nil, fmt.Errorf("unable to scanForFiles dir '%s': %w", path, err)
			}
			goFiles = append(goFiles, additionalFiles...)
		default:
			fileName := file.Name()
			if !strings.HasSuffix(fileName, ".go") {
				continue
			}

			goFile, err := newFile(fileSet, path)
			if err != nil {
				return nil, fmt.Errorf("unable to open go file '%s': %w", path, err)
			}
			goFiles = append(goFiles, goFile)
		}
	}

	return goFiles, nil
}
