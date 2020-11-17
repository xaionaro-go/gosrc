package gosrc

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/fatih/structtag"
)

type Field struct {
	ast.Field
	Struct *Struct

	Names     []*ast.Ident
	Index     uint
	TypeValue types.TypeAndValue
}
type Fields []*Field

func (field *Field) IsPointer() bool {
	_, ok := TypeDeepest(field.TypeValue.Type).(*types.Pointer)
	return ok
}

func (field *Field) TypeElem() types.Type {
	return TypeElem(field.TypeValue.Type)
}

func (field *Field) TagGet(key string) (string, bool) {
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
