package codegen

import (
	"fmt"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
	"github.com/haira-lang/haira/internal/checker"
)

// activeTypeInfo holds type info during code generation. Set by GenerateMainGo.
var activeTypeInfo *checker.TypeInfo

// activeSourceFile holds the source file being generated (for struct lookups).
var activeSourceFile *ast.SourceFile

// activeWorkflowName holds the current workflow name during codegen (for step telemetry).
var activeWorkflowName string

// activeStepName holds the current step name during codegen (for log.* calls).
var activeStepName string

// knownToolNames tracks declared tool names so tool-to-tool calls resolve correctly.
var knownToolNames map[string]bool

// activeTarget holds the compilation target during codegen ("native" or "workers").
var activeTarget = "native"

// CheckerTypeToGo converts a checker.Type to a Go type string.
func CheckerTypeToGo(ty checker.Type) string {
	if ty == nil {
		return "any"
	}
	switch t := ty.(type) {
	case checker.IntType:
		return "int"
	case checker.FloatType:
		return "float64"
	case checker.StringType:
		return "string"
	case checker.BoolType:
		return "bool"
	case checker.AnyType:
		return "any"
	case checker.VoidType:
		return ""
	case checker.ListType:
		return "[]" + CheckerTypeToGo(t.Elem)
	case checker.MapType:
		return fmt.Sprintf("map[%s]%s", CheckerTypeToGo(t.Key), CheckerTypeToGo(t.Value))
	case checker.StructType:
		return t.Name
	case checker.EnumType:
		return t.Name
	case checker.FuncType:
		params := make([]string, len(t.Params))
		for i, p := range t.Params {
			params[i] = CheckerTypeToGo(p)
		}
		ret := CheckerTypeToGo(t.Return)
		if ret == "" {
			return fmt.Sprintf("func(%s)", strings.Join(params, ", "))
		}
		return fmt.Sprintf("func(%s) %s", strings.Join(params, ", "), ret)
	default:
		return "any"
	}
}

// qualifiedTypeToGo maps a dotted stdlib type name (e.g. "ui.StatusCard")
// to its Go runtime type.
// Core modules: module.Type → haira.<Module><Type> (e.g. ui.StatusCard → haira.UiStatusCard)
// Stdlib packages: module.Type → <module>.<Type> (e.g. postgres.DB → postgres.DB)
func qualifiedTypeToGo(name string) (string, bool) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) != 2 {
		return "", false
	}
	if !IsStdlibImport(parts[0]) {
		return "", false
	}
	// Stdlib packages that have their own Go package — types keep their original names
	if _, ok := stdlibGoPackage(parts[0]); ok {
		return parts[0] + "." + parts[1], true
	}
	// Core modules: prefix with haira.
	return "haira." + Capitalize(parts[0]) + parts[1], true
}

// lookupExprGoType returns the Go type string for an expression span, or "".
func lookupExprGoType(span ast.Span) string {
	if activeTypeInfo == nil {
		return ""
	}
	ty, ok := activeTypeInfo.ExprTypes[span]
	if !ok {
		return ""
	}
	return CheckerTypeToGo(ty)
}

// isTypedNonAny returns true if the span has a known non-any type.
func isTypedNonAny(span ast.Span) bool {
	if activeTypeInfo == nil {
		return false
	}
	ty, ok := activeTypeInfo.ExprTypes[span]
	if !ok {
		return false
	}
	_, isAny := ty.(checker.AnyType)
	return !isAny
}

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
		case "file":
			return "string" // file is a temp path at runtime
		default:
			// Qualified stdlib types: ui.StatusCard → haira.UIStatusCard
			if goType, ok := qualifiedTypeToGo(t.Name); ok {
				return goType
			}
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
