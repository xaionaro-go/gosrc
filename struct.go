package gosrc

import (
	"fmt"
	"go/ast"
	"go/types"
)

type Struct struct {
	File       *File
	TypeSpec   *ast.TypeSpec
	StructType *ast.StructType
}

type Structs []*Struct

func (_struct Struct) String() string {
	return fmt.Sprintf("struct:%s", _struct.TypeSpec.Name)
}

func (_struct *Struct) Fields() (Fields, error) {
	var goFields Fields
	for idx, field := range _struct.StructType.Fields.List {
		typ, err := _struct.toType(field.Type)
		if err != nil {
			return nil, fmt.Errorf("unable to lookup type '%s': %w", field.Type, err)
		}
		goFields = append(goFields, &Field{
			Field:     *field,
			Struct:    _struct,
			Names:     field.Names,
			Index:     uint(idx),
			TypeValue: typ,
		})
	}
	return goFields, nil
}

func (_struct Struct) toType(expr ast.Expr) (types.TypeAndValue, error) {
	return _struct.File.toType(expr)
}
