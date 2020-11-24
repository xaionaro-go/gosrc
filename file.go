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

// File represents one source code file.
type File struct {
	Path    string
	Package *Package
	Ast     *ast.File
}

// Files is a set of File-s.
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

// IsPassBuildTags returns true if file satisfies specified build tags.
//
// See also https://golang.org/cmd/go/#hdr-Build_constraints
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

// GoGenerateTags returns all tags listed in `go:generate` magic comments.
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

// String just implements fmt.Stringer
func (file File) String() string {
	return fmt.Sprintf("%s[%s]", file.Path, file.PackageName())
}

// Dir returns the path to the directory of the source code file.
func (file File) Dir() string {
	return path.Dir(file.Path)
}

// PackageName returns the name of the Go package.
func (file File) PackageName() string {
	return file.Ast.Name.Name
}

// ToType returns types.TypeAndValue for a specified type expression
// (if one was parsed while scanning).
func (file *File) ToType(expr ast.Expr) (types.TypeAndValue, error) {
	if file == nil {
		return types.TypeAndValue{}, fmt.Errorf("file is nil")
	}
	return file.Package.ToType(expr)
}

// Funcs returns all functions defined in the file.
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

// Structs returns all structures defined in the file.
func (file *File) Structs() Structs {
	return file.structsWithMagicComment(nil)
}

// StructsWithMagicComment returns structures (defined in the file),
// which has the specified "go:" magic comment.
func (file *File) StructsWithMagicComment(magicComment string) Structs {
	return file.structsWithMagicComment(&magicComment)
}

func (file *File) structsWithMagicComment(magicComment *string) Structs {
	var structs Structs
	file.findTypes(magicComment, func(typeSpec *ast.TypeSpec) {
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok || structType.Incomplete {
			return
		}
		structs = append(structs, &Struct{
			AstTypeSpec: AstTypeSpec{
				File:     file,
				TypeSpec: typeSpec,
			},
		})
	})
	return structs
}

// AstTypeSpecs returns all type definitions.
func (file *File) AstTypeSpecs() AstTypeSpecs {
	var astTypeSpecs AstTypeSpecs
	file.findTypes(nil, func(typeSpec *ast.TypeSpec) {
		astTypeSpecs = append(astTypeSpecs, &AstTypeSpec{
			File:     file,
			TypeSpec: typeSpec,
		})
	})
	return astTypeSpecs
}

func (file *File) findTypes(magicComment *string, appendFn func(typeSpec *ast.TypeSpec)) {
	var expectedComment string
	if magicComment != nil {
		expectedComment = `//go:` + *magicComment
	}

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
			if typeSpec.Doc == nil {
				typeSpec.Doc = genDecl.Doc
			}

			appendFn(typeSpec)
		}
	}
	return
}

// IsTest returns true if the source code file is of unit-tests.
func (file File) IsTest() bool {
	return strings.HasSuffix(file.Path, `_test.go`)
}

// ImportPaths returns all imports defined in the file.
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

// FindByPath finds a file by its path (if there's one), otherwise returns nil.
func (files Files) FindByPath(path string) *File {
	for _, file := range files {
		if file.Path == path {
			return file
		}
	}
	return nil
}

// FilterByBuildTags returns files which satisfies specified build tags.
//
// See also https://golang.org/cmd/go/#hdr-Build_constraints
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

// FilterByGoGenerateTag returns files which has a specified "go:generate" tag.
func (files Files) FilterByGoGenerateTag(goGenerateTag string) Files {
	var filteredFiles Files
	for _, file := range files {
		for _, tag := range file.GoGenerateTags() {
			if tag == goGenerateTag {
				filteredFiles = append(filteredFiles, file)
				break
			}
		}
	}
	return filteredFiles
}
