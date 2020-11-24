package gosrc

import (
	"fmt"
	"go/ast"
	"go/types"
)

// AstTypeSpec represents one ast.TypeSpec.
type AstTypeSpec struct {
	File     *File
	TypeSpec *ast.TypeSpec
}

func (astTypeSpec AstTypeSpec) toType(expr ast.Expr) (types.TypeAndValue, error) {
	return astTypeSpec.File.ToType(expr)
}

// Name returns the type name of the structure.
func (astTypeSpec AstTypeSpec) Name() string {
	return astTypeSpec.TypeSpec.Name.String()
}

// Methods returns all methods of the structure.
func (astTypeSpec AstTypeSpec) Methods() Funcs {
	return astTypeSpec.File.Package.Funcs().FindMethodsOf(astTypeSpec.TypeSpec.Name.Name)
}

// MethodByName returns the method of the type by its name (or nil of there
// is no such method)
func (astTypeSpec AstTypeSpec) MethodByName(methodName string) *Func {
	funcs := astTypeSpec.Methods().FindByName(methodName)
	switch len(funcs) {
	case 0:
		return nil
	case 1:
		return funcs[0]
	default:
		panic(fmt.Sprintf("found more than one method of '%s' with the same name '%s': %d", astTypeSpec.Name(), methodName, len(funcs)))
	}
}

// AstTypeSpecs represents multiple Type-s.
type AstTypeSpecs []*AstTypeSpec
