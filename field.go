package gosrc

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"github.com/fatih/structtag"
)

// Field represents one field of a structure.
type Field struct {
	ast.Field
	Struct *Struct

	Names     []*ast.Ident
	Index     uint
	TypeValue types.TypeAndValue
}
type Fields []*Field

// IsPointer returns true if the field has a pointer value
func (field Field) IsPointer() bool {
	_, ok := TypeDeepest(field.TypeValue.Type).(*types.Pointer)
	return ok
}

// IsSlice returns true if the field has a slice value
func (field Field) IsSlice() bool {
	_, ok := TypeDeepest(field.TypeValue.Type).(*types.Slice)
	return ok
}

// TypeElem returns Elem type of the value type.
//
// It will panic if IsPointer is false.
func (field Field) TypeElem() types.Type {
	return TypeElem(field.TypeValue.Type)
}

// TagGet returns a value of the struct field tag with the specified key.
func (field Field) TagGet(key string) (string, bool) {
	if field.Tag == nil {
		return "", false
	}
	tags, err := structtag.Parse(strings.Trim(field.Tag.Value, "`"))
	if err != nil {
		return "", false
	}
	tag, err := tags.Get(key)
	if err != nil {
		return "", false
	}
	return tag.Value(), true
}

// TypeStdSize returns the size (in bytes) of the value for specified
// word size and max alignment. See more details in types.StdSizes.
func (field Field) TypeStdSize(wordSize, maxAlign int64) int64 {
	return (&types.StdSizes{
		WordSize: wordSize,
		MaxAlign: maxAlign,
	}).Sizeof(field.TypeValue.Type)
}

// Name returns the name of the field. For anonymous field it returns
// the name of the value type.
func (field Field) Name() string {
	if len(field.Names) > 0 {
		return field.Names[0].String()
	}

	// The field has no name, so we use Type name as the field name.
	return itemTypeName(field.TypeValue.Type).Name
}

// ItemTypeName returns the name of the value type of an item referenced
// by the field. Few examples:
// * If field value type is uint64, then the returned value will be 'uint64'.
// * If field value type is []uint64, then the returned value will be 'uint64'.
// * If field value type is *uint64, then the returned value will be 'uint64'.
//
// This could be useful to find dependencies by names.
func (field Field) ItemTypeName() TypeNameValue {
	return itemTypeName(field.TypeValue.Type)
}

// TypeNameValue is just a combination of a value type name and of a path
// where the type is defined.
type TypeNameValue struct {
	Name string
	Path string
}

func itemTypeName(typ types.Type) TypeNameValue {
	for {
		modified := true
		switch casted := typ.(type) {
		case interface{ Elem() types.Type }:
			typ = casted.Elem()
		case *types.Pointer:
			typ = casted.Underlying()
		default:
			modified = false
		}
		if !modified {
			break
		}
	}
	// typ.String() is presented in forms "my/path/MyType" and "my/path.MyType"
	parts := strings.Split(typ.String(), "/")
	// lastPart is presented in forms "MyType" and "path.MyType"
	lastPart := parts[len(parts)-1]
	// subParts either ["MyType"] or ["path", "MyType"]
	subParts := strings.Split(lastPart, ".")
	switch len(subParts) {
	case 1:
		return TypeNameValue{
			Name: subParts[0],
			Path: "",
		}
	case 2:
		parts[len(parts)-1] = subParts[0]
		return TypeNameValue{
			Name: subParts[1],
			Path: strings.Join(parts, "/"),
		}
	default:
		panic(fmt.Sprintf("do not know how to parse '%s'", typ.String()))
	}
}
