package gosrc

import (
	"go/types"
)

func TypeDeepest(typ types.Type) types.Type {
	oldType := typ
	for {
		typ = typ.Underlying()
		if typ == oldType {
			return typ
		}
		oldType = typ
	}
}

func TypeElem(typ types.Type) types.Type {
	return TypeDeepest(typ).(*types.Pointer).Elem()
}
