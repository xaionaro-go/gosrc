package gosrc

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var magicGoGenerateComment = regexp.MustCompile(`go:generate ([0-9A-Za-z_\.]+)`)
var magicBuildTagsComment = regexp.MustCompile(`\+build (.*)`)

type File struct {
	Path    string
	Package *Package
	Ast     *ast.File
}

type Files []*File

func newFile(fileSet *token.FileSet, path string) (*File, error) {
	parsedFile, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("cannot parse go file '%s': %w", path, err)
	}

	return &File{
		Path: path,
		Ast:  parsedFile,
	}, nil
}

func (file File) IsPassBuildTags(haveBuildTags []string) bool {
	have := map[string]bool{}
	for _, buildTag := range haveBuildTags {
		have[buildTag] = true
	}

	for _, commentGroup := range file.Ast.Comments {
		for _, comment := range commentGroup.List {
			buildTagsMatches := magicBuildTagsComment.FindStringSubmatch(comment.Text)
			if len(buildTagsMatches) < 2 {
				continue
			}
			combinations := strings.Split(strings.TrimSpace(buildTagsMatches[1]), ",")
			for _, combination := range combinations {
				requiredBuildTags := strings.Split(strings.TrimSpace(combination), " ")
				satisfied := true
				for _, requiredBuildTag := range requiredBuildTags {
					if strings.HasSuffix(requiredBuildTag, "!") {
						shouldAbsentBuildTag := strings.TrimSpace(requiredBuildTag[:1])
						if have[shouldAbsentBuildTag] {
							satisfied = false
							break
						}
					} else {
						if !have[requiredBuildTag] {
							satisfied = false
							break
						}
					}
				}
				if !satisfied {
					return false
				}
			}
		}
	}

	return true
}

func (file File) GoGenerateTags() []string {
	var goGenerateTags []string
	for _, commentGroup := range file.Ast.Comments {
		for _, comment := range commentGroup.List {
			goGenerateMatches := magicGoGenerateComment.FindStringSubmatch(comment.Text)
			if len(goGenerateMatches) < 2 {
				continue
			}
			goGenerateTags = append(goGenerateTags, goGenerateMatches[1])
		}
	}
	return goGenerateTags
}

func (file File) String() string {
	return fmt.Sprintf("%s[%s]", file.Path, file.PackageName())
}

func (file File) Dir() string {
	return path.Dir(file.Path)
}

func (file File) PackageName() string {
	return file.Ast.Name.Name
}

func (file *File) toType(expr ast.Expr) (types.TypeAndValue, error) {
	if file == nil {
		return types.TypeAndValue{}, fmt.Errorf("file is nil")
	}
	return file.Package.toType(expr)
}

func (file *File) Funcs() Funcs {
	var funcs Funcs
	for _, decl := range file.Ast.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if funcDecl.Recv == nil {
			continue
		}
		funcs = append(funcs, newFunc(funcDecl))
	}
	return funcs
}

func (file *File) Structs() Structs {
	return file.findStructs(nil)
}

func (file *File) FindStructsByMagicComment(magicComment string) Structs {
	return file.findStructs(&magicComment)
}

func (file *File) findStructs(magicComment *string) Structs {
	var expectedComment string
	if magicComment != nil {
		expectedComment = `//go:` + *magicComment
	}

	var structs Structs
	for _, decl := range file.Ast.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		if expectedComment != "" {
			if genDecl.Doc == nil {
				continue
			}
			shouldAdd := false
			for _, list := range genDecl.Doc.List {
				if list.Text == expectedComment {
					shouldAdd = true
					break
				}
			}
			if !shouldAdd {
				continue
			}
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			if structType.Incomplete {
				continue
			}

			structs = append(structs, &Struct{
				File:       file,
				TypeSpec:   typeSpec,
				StructType: structType,
			})
		}
	}
	return structs
}

func (file File) GoPath() string {
	return file.Path
}

func (file File) IsTest() bool {
	return strings.HasSuffix(file.Path, `_test.go`)
}

func (file File) ImportPaths() ([]string, error) {
	var result []string
	for _, importSpec := range file.Ast.Imports {
		if importSpec.Path == nil {
			return nil, fmt.Errorf("importSpec.Path == nil")
		}
		_path, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("unable to Unquote _path <%s>: %w",
				importSpec.Path.Value, err)
		}
		result = append(result, _path)
	}
	return result, nil
}

func (files Files) FindByPath(path string) *File {
	for _, file := range files {
		if file.Path == path {
			return file
		}
	}
	return nil
}

func (files Files) FilterByBuildTags(buildTags []string) Files {
	var result Files
	for _, file := range files {
		if !file.IsPassBuildTags(buildTags) {
			continue
		}
		result = append(result, file)
	}

	return result
}
