package formatter

import (
	"github.com/haira-lang/haira/internal/ast"
)

// formatType formats a type annotation.
func (f *Formatter) formatType(ty ast.Type) {
	switch t := ty.(type) {
	case ast.NamedType:
		f.write(t.Name)
	case ast.ListType:
		f.write("[")
		f.formatType(t.Elem.Node)
		f.write("]")
	case ast.MapType:
		f.write("{")
		f.formatType(t.Key.Node)
		f.write(": ")
		f.formatType(t.Value.Node)
		f.write("}")
	case ast.OptionType:
		f.formatType(t.Inner.Node)
		f.write("?")
	case ast.FunctionType:
		f.write("(")
		for i, p := range t.Params {
			if i > 0 {
				f.write(", ")
			}
			f.formatType(p.Node)
		}
		f.write(") -> ")
		f.formatType(t.Ret.Node)
	case ast.UnionType:
		for i, v := range t.Variants {
			if i > 0 {
				f.write(" | ")
			}
			f.formatType(v.Node)
		}
	case ast.GenericType:
		f.write(t.Name)
		f.write("[")
		for i, a := range t.Args {
			if i > 0 {
				f.write(", ")
			}
			f.formatType(a.Node)
		}
		f.write("]")
	}
}

// formatTypeAnnotation formats " -> Type" if a return type is present.
func (f *Formatter) formatTypeAnnotation(ty *ast.Spanned[ast.Type]) {
	if ty != nil {
		f.write(" -> ")
		f.formatType(ty.Node)
	}
}

// formatColonType formats ": Type" for param/field type annotations.
func (f *Formatter) formatColonType(ty *ast.Spanned[ast.Type]) {
	if ty != nil {
		f.write(": ")
		f.formatType(ty.Node)
	}
}
