package gosrc

import (
	"go/ast"
)

type Func struct {
	*ast.FuncDecl
}
type Funcs []*Func

func newFunc(funcDecl *ast.FuncDecl) *Func {
	return &Func{
		FuncDecl: funcDecl,
	}
}

func (funcs Funcs) FindMethodsOf(typ *ast.Ident) Funcs {
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
			if ident.Name == typ.Name {
				result = append(result, fn)
				break
			}
		}
	}
	return result
}
