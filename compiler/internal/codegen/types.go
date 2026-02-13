package codegen

import (
	"fmt"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
)

// HairaTypeToGo converts a Haira type to its Go representation.
func HairaTypeToGo(ty ast.Type) string {
	switch t := ty.(type) {
	case ast.NamedType:
		switch t.Name {
		case "int":
			return "int"
		case "float":
			return "float64"
		case "string":
			return "string"
		case "bool":
			return "bool"
		case "any":
			return "any"
		case "map":
			return "map[string]any"
		default:
			return t.Name
		}
	case ast.ListType:
		return "[]" + HairaTypeToGo(t.Elem.Node)
	case ast.MapType:
		return fmt.Sprintf("map[%s]%s", HairaTypeToGo(t.Key.Node), HairaTypeToGo(t.Value.Node))
	case ast.OptionType:
		return "*" + HairaTypeToGo(t.Inner.Node)
	case ast.FunctionType:
		params := make([]string, len(t.Params))
		for i, p := range t.Params {
			params[i] = HairaTypeToGo(p.Node)
		}
		return fmt.Sprintf("func(%s) %s", strings.Join(params, ", "), HairaTypeToGo(t.Ret.Node))
	case ast.UnionType:
		return "any"
	case ast.GenericType:
		args := make([]string, len(t.Args))
		for i, a := range t.Args {
			args[i] = HairaTypeToGo(a.Node)
		}
		return fmt.Sprintf("%s[%s]", t.Name, strings.Join(args, ", "))
	default:
		return "any"
	}
}
