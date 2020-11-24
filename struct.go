package gosrc

import (
	"fmt"
	"go/ast"
)

// Struct represents one structure type of the source code file.
type Struct struct {
	AstTypeSpec
}

// Structs is a set of Struct-s
type Structs []*Struct

// String just implements fmt.Stringer
func (_struct Struct) String() string {
	return fmt.Sprintf("struct:%s", _struct.Name())
}

// Fields returns all the Fields of the structure.
func (_struct *Struct) Fields() (Fields, error) {
	var goFields Fields
	structType := _struct.StructType()
	if structType == nil {
		return nil, fmt.Errorf("no struct type in %#+v", _struct)
	}
	for idx, field := range structType.Fields.List {
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

func (_struct Struct) StructType() *ast.StructType {
	structType, ok := _struct.TypeSpec.Type.(*ast.StructType)
	if !ok {
		return nil
	}
	return structType
}
