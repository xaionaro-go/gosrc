package gosrc

import (
	"fmt"
	"go/ast"
	"go/types"
)

// Struct represents one structure type of the source code file.
type Struct struct {
	File       *File
	TypeSpec   *ast.TypeSpec
	StructType *ast.StructType
}

type Structs []*Struct

// String just implements fmt.Stringer
func (_struct Struct) String() string {
	return fmt.Sprintf("struct:%s", _struct.Name())
}

// Fields returns all the Fields of the structure.
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
	return _struct.File.ToType(expr)
}

// Name returns the type name of the structure.
func (_struct Struct) Name() string {
	return _struct.TypeSpec.Name.String()
}
