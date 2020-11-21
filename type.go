package gosrc

import (
	"go/types"
)

// TypeDeepest just calls Underlying of a type until it will reach some end,
// and returns it.
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

// TypeElem returns Elem type of the specified type. It will panic, if the
// type is not a *types.Pointer.
func TypeElem(typ types.Type) types.Type {
	return TypeDeepest(typ).(*types.Pointer).Elem()
}
