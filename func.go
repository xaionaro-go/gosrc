package gosrc

import (
	"go/ast"
)

// Func represents one function of a source code file.
type Func struct {
	*ast.FuncDecl
}

// Funcs is a set of Func-s.
type Funcs []*Func

func newFunc(funcDecl *ast.FuncDecl) *Func {
	return &Func{
		FuncDecl: funcDecl,
	}
}

// FindMethodsOf returns all methods of a specified type.
func (funcs Funcs) FindMethodsOf(typName string) Funcs {
	var result Funcs
	for _, fn := range funcs {
		for _, recv := range fn.FuncDecl.Recv.List {
			cmpTyp := recv.Type
			if starExpr, ok := cmpTyp.(*ast.StarExpr); ok {
				cmpTyp = starExpr.X
			}
			ident, ok := cmpTyp.(*ast.Ident)
			if !ok {
				continue
			}
			if ident.Name == typName {
				result = append(result, fn)
				break
			}
		}
	}
	return result
}

// FindByName returns functions which name is equals to the selected one.
func (funcs Funcs) FindByName(funcName string) []*Func {
	var result Funcs
	for _, fn := range funcs {
		if fn.FuncDecl.Name.Name == funcName {
			result = append(result, fn)
		}
	}
	return result
}
