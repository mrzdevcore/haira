package codegen

import (
	"fmt"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
)

// ResolveStdlibCall checks if a call is a known stdlib function and returns Go code.
func ResolveStdlibCall(call ast.CallExpr) (string, bool) {
	// Qualified calls: io.println, http.get, etc.
	if field, ok := call.Callee.Node.(ast.FieldExpr); ok {
		if ident, ok := field.Object.Node.(ast.IdentExpr); ok {
			method := field.Field.Node
			args := callArgsToGo(call.Args)
			if resolved, ok := resolveQualified(ident.Name, method, args, call); ok {
				return resolved, true
			}
		}
	}

	// Bare calls: env(), len(), keys(), join()
	if ident, ok := call.Callee.Node.(ast.IdentExpr); ok {
		args := callArgsToGo(call.Args)
		switch ident.Name {
		case "env":
			return fmt.Sprintf("haira.Env(%s)", args), true
		case "len":
			return fmt.Sprintf("haira.Len(%s)", args), true
		case "keys":
			return fmt.Sprintf("haira.Keys(%s)", args), true
		case "join":
			return fmt.Sprintf("haira.Join(%s)", args), true
		}
	}

	return "", false
}

// ResolveStdlibMethodCall checks if a method call is a known stdlib call.
func ResolveStdlibMethodCall(mc ast.MethodCallExpr) (string, bool) {
	if ident, ok := mc.Receiver.Node.(ast.IdentExpr); ok {
		method := mc.Method.Node
		args := callArgsToGo(mc.Args)
		stub := ast.CallExpr{Callee: mc.Receiver, Args: mc.Args}
		return resolveQualified(ident.Name, method, args, stub)
	}
	return "", false
}

func resolveQualified(module, method, args string, call ast.CallExpr) (string, bool) {
	switch module {
	case "io":
		switch method {
		case "println":
			return fmt.Sprintf("haira.Println(%s)", args), true
		case "print":
			return fmt.Sprintf("haira.Print(%s)", args), true
		case "printf":
			return fmt.Sprintf("haira.Printf(%s)", args), true
		case "readln":
			return "haira.Readln()", true
		case "read_file":
			return fmt.Sprintf("haira.ReadFile(%s)", args), true
		case "eprintln":
			return fmt.Sprintf("haira.Eprintln(%s)", args), true
		case "eprintf":
			return fmt.Sprintf("haira.Eprintf(%s)", args), true
		}
	case "http":
		switch method {
		case "get":
			return fmt.Sprintf("haira.HttpGet(%s)", args), true
		case "get_with_headers":
			return fmt.Sprintf("haira.HttpGetWithHeaders(%s)", args), true
		case "post":
			return fmt.Sprintf("haira.HttpPost(%s)", args), true
		case "post_with_headers":
			return fmt.Sprintf("haira.HttpPostWithHeaders(%s)", args), true
		case "put":
			return fmt.Sprintf("haira.HttpPut(%s)", args), true
		case "put_with_headers":
			return fmt.Sprintf("haira.HttpPutWithHeaders(%s)", args), true
		case "delete":
			return fmt.Sprintf("haira.HttpDelete(%s)", args), true
		case "delete_with_headers":
			return fmt.Sprintf("haira.HttpDeleteWithHeaders(%s)", args), true
		case "Server":
			return resolveServerCall(call), true
		}
	case "mcp":
		switch method {
		case "Server":
			return resolveMCPServerCall(call), true
		}
	case "vector":
		switch method {
		case "embed":
			return resolveVectorEmbed(call), true
		case "embed_batch":
			return resolveVectorEmbedBatch(call), true
		case "collection":
			return resolveVectorCollection(call), true
		case "insert":
			return fmt.Sprintf("haira.VectorInsert(%s)", args), true
		case "search":
			return fmt.Sprintf("haira.VectorSearch(%s)", args), true
		case "format":
			return fmt.Sprintf("haira.VectorFormat(%s)", args), true
		}
	case "json":
		switch method {
		case "marshal", "encode":
			return fmt.Sprintf("haira.JSONMarshal(%s)", args), true
		case "unmarshal", "decode":
			return fmt.Sprintf("haira.JSONUnmarshal(%s)", args), true
		case "marshal_pretty":
			return fmt.Sprintf("haira.JSONMarshalPretty(%s)", args), true
		case "parse":
			return fmt.Sprintf("haira.JSONParse(%s)", args), true
		}
	case "string":
		switch method {
		case "len":
			return fmt.Sprintf("haira.StringLen(%s)", args), true
		case "is_empty":
			return fmt.Sprintf("haira.StringIsEmpty(%s)", args), true
		case "trim":
			return fmt.Sprintf("haira.StringTrim(%s)", args), true
		case "trim_left":
			return fmt.Sprintf("haira.StringTrimLeft(%s)", args), true
		case "trim_right":
			return fmt.Sprintf("haira.StringTrimRight(%s)", args), true
		case "to_upper":
			return fmt.Sprintf("haira.StringToUpper(%s)", args), true
		case "to_lower":
			return fmt.Sprintf("haira.StringToLower(%s)", args), true
		case "contains":
			return fmt.Sprintf("haira.StringContains(%s)", args), true
		case "starts_with":
			return fmt.Sprintf("haira.StringStartsWith(%s)", args), true
		case "ends_with":
			return fmt.Sprintf("haira.StringEndsWith(%s)", args), true
		case "index_of":
			return fmt.Sprintf("haira.StringIndexOf(%s)", args), true
		case "last_index_of":
			return fmt.Sprintf("haira.StringLastIndexOf(%s)", args), true
		case "split":
			return fmt.Sprintf("haira.StringSplit(%s)", args), true
		case "join":
			return fmt.Sprintf("haira.StringJoin(%s)", args), true
		case "substring":
			return fmt.Sprintf("haira.StringSubstring(%s)", args), true
		case "char_at":
			return fmt.Sprintf("haira.StringCharAt(%s)", args), true
		case "replace":
			return fmt.Sprintf("haira.StringReplace(%s)", args), true
		case "replace_all":
			return fmt.Sprintf("haira.StringReplaceAll(%s)", args), true
		case "repeat":
			return fmt.Sprintf("haira.StringRepeat(%s)", args), true
		}
	case "regex":
		switch method {
		case "is_match":
			return fmt.Sprintf("haira.RegexIsMatch(%s)", args), true
		case "find":
			return fmt.Sprintf("haira.RegexFind(%s)", args), true
		case "find_all":
			return fmt.Sprintf("haira.RegexFindAll(%s)", args), true
		case "replace":
			return fmt.Sprintf("haira.RegexReplace(%s)", args), true
		case "replace_all":
			return fmt.Sprintf("haira.RegexReplaceAll(%s)", args), true
		case "captures":
			return fmt.Sprintf("haira.RegexCaptures(%s)", args), true
		case "split":
			return fmt.Sprintf("haira.RegexSplit(%s)", args), true
		}
	case "math":
		switch method {
		case "abs":
			return fmt.Sprintf("haira.MathAbs(%s)", args), true
		case "min":
			return fmt.Sprintf("haira.MathMin(%s)", args), true
		case "max":
			return fmt.Sprintf("haira.MathMax(%s)", args), true
		case "clamp":
			return fmt.Sprintf("haira.MathClamp(%s)", args), true
		case "floor":
			return fmt.Sprintf("haira.MathFloor(%s)", args), true
		case "ceil":
			return fmt.Sprintf("haira.MathCeil(%s)", args), true
		case "round":
			return fmt.Sprintf("haira.MathRound(%s)", args), true
		case "trunc":
			return fmt.Sprintf("haira.MathTrunc(%s)", args), true
		case "pow":
			return fmt.Sprintf("haira.MathPow(%s)", args), true
		case "sqrt":
			return fmt.Sprintf("haira.MathSqrt(%s)", args), true
		case "cbrt":
			return fmt.Sprintf("haira.MathCbrt(%s)", args), true
		case "exp":
			return fmt.Sprintf("haira.MathExp(%s)", args), true
		case "log":
			return fmt.Sprintf("haira.MathLog(%s)", args), true
		case "log10":
			return fmt.Sprintf("haira.MathLog10(%s)", args), true
		case "log2":
			return fmt.Sprintf("haira.MathLog2(%s)", args), true
		case "sin":
			return fmt.Sprintf("haira.MathSin(%s)", args), true
		case "cos":
			return fmt.Sprintf("haira.MathCos(%s)", args), true
		case "tan":
			return fmt.Sprintf("haira.MathTan(%s)", args), true
		case "asin":
			return fmt.Sprintf("haira.MathAsin(%s)", args), true
		case "acos":
			return fmt.Sprintf("haira.MathAcos(%s)", args), true
		case "atan":
			return fmt.Sprintf("haira.MathAtan(%s)", args), true
		case "atan2":
			return fmt.Sprintf("haira.MathAtan2(%s)", args), true
		case "random":
			return "haira.MathRandom()", true
		case "random_int":
			return fmt.Sprintf("haira.MathRandomInt(%s)", args), true
		case "pi":
			return "haira.MathPI", true
		case "e":
			return "haira.MathE", true
		}
	case "conv":
		switch method {
		case "int_to_string":
			return fmt.Sprintf("haira.ConvIntToString(%s)", args), true
		case "float_to_string":
			return fmt.Sprintf("haira.ConvFloatToString(%s)", args), true
		case "bool_to_string":
			return fmt.Sprintf("haira.ConvBoolToString(%s)", args), true
		case "string_to_int":
			return fmt.Sprintf("haira.ConvStringToInt(%s)", args), true
		case "string_to_float":
			return fmt.Sprintf("haira.ConvStringToFloat(%s)", args), true
		case "string_to_bool":
			return fmt.Sprintf("haira.ConvStringToBool(%s)", args), true
		case "int_to_float":
			return fmt.Sprintf("haira.ConvIntToFloat(%s)", args), true
		case "float_to_int":
			return fmt.Sprintf("haira.ConvFloatToInt(%s)", args), true
		case "int_to_hex":
			return fmt.Sprintf("haira.ConvIntToHex(%s)", args), true
		case "int_to_binary":
			return fmt.Sprintf("haira.ConvIntToBinary(%s)", args), true
		case "int_to_octal":
			return fmt.Sprintf("haira.ConvIntToOctal(%s)", args), true
		case "hex_to_int":
			return fmt.Sprintf("haira.ConvHexToInt(%s)", args), true
		}
	case "array":
		switch method {
		case "len":
			return fmt.Sprintf("haira.ArrayLen(%s)", args), true
		case "is_empty":
			return fmt.Sprintf("haira.ArrayIsEmpty(%s)", args), true
		case "first":
			return fmt.Sprintf("haira.ArrayFirst(%s)", args), true
		case "last":
			return fmt.Sprintf("haira.ArrayLast(%s)", args), true
		case "get":
			return fmt.Sprintf("haira.ArrayGet(%s)", args), true
		case "push":
			return fmt.Sprintf("haira.ArrayPush(%s)", args), true
		case "pop":
			return fmt.Sprintf("haira.ArrayPop(%s)", args), true
		case "insert":
			return fmt.Sprintf("haira.ArrayInsert(%s)", args), true
		case "remove":
			return fmt.Sprintf("haira.ArrayRemove(%s)", args), true
		case "slice":
			return fmt.Sprintf("haira.ArraySlice(%s)", args), true
		case "take":
			return fmt.Sprintf("haira.ArrayTake(%s)", args), true
		case "drop":
			return fmt.Sprintf("haira.ArrayDrop(%s)", args), true
		case "contains":
			return fmt.Sprintf("haira.ArrayContains(%s)", args), true
		case "index_of":
			return fmt.Sprintf("haira.ArrayIndexOf(%s)", args), true
		case "reverse":
			return fmt.Sprintf("haira.ArrayReverse(%s)", args), true
		case "concat":
			return fmt.Sprintf("haira.ArrayConcat(%s)", args), true
		case "flatten":
			return fmt.Sprintf("haira.ArrayFlatten(%s)", args), true
		case "unique":
			return fmt.Sprintf("haira.ArrayUnique(%s)", args), true
		case "sort":
			return fmt.Sprintf("haira.ArraySort(%s)", args), true
		case "join":
			return fmt.Sprintf("haira.ArrayJoin(%s)", args), true
		case "map":
			return fmt.Sprintf("haira.ArrayMap(%s)", args), true
		case "filter":
			return fmt.Sprintf("haira.ArrayFilter(%s)", args), true
		case "reduce":
			return fmt.Sprintf("haira.ArrayReduce(%s)", args), true
		case "find":
			return fmt.Sprintf("haira.ArrayFind(%s)", args), true
		case "find_index":
			return fmt.Sprintf("haira.ArrayFindIndex(%s)", args), true
		case "sort_by":
			return fmt.Sprintf("haira.ArraySortBy(%s)", args), true
		case "every":
			return fmt.Sprintf("haira.ArrayEvery(%s)", args), true
		case "some":
			return fmt.Sprintf("haira.ArraySome(%s)", args), true
		case "for_each":
			return fmt.Sprintf("haira.ArrayForEach(%s)", args), true
		case "flat_map":
			return fmt.Sprintf("haira.ArrayFlatMap(%s)", args), true
		}
	case "map":
		switch method {
		case "len":
			return fmt.Sprintf("haira.MapLen(%s)", args), true
		case "is_empty":
			return fmt.Sprintf("haira.MapIsEmpty(%s)", args), true
		case "get":
			return fmt.Sprintf("haira.MapGet(%s)", args), true
		case "has":
			return fmt.Sprintf("haira.MapHas(%s)", args), true
		case "set":
			return fmt.Sprintf("haira.MapSet(%s)", args), true
		case "remove":
			return fmt.Sprintf("haira.MapRemove(%s)", args), true
		case "keys":
			return fmt.Sprintf("haira.MapKeys(%s)", args), true
		case "values":
			return fmt.Sprintf("haira.MapValues(%s)", args), true
		case "entries":
			return fmt.Sprintf("haira.MapEntries(%s)", args), true
		case "merge":
			return fmt.Sprintf("haira.MapMerge(%s)", args), true
		case "contains_value":
			return fmt.Sprintf("haira.MapContainsValue(%s)", args), true
		}
	case "postgres":
		if method == "connect" {
			return fmt.Sprintf("haira.PostgresConnect(%s)", args), true
		}
	case "slack":
		if method == "send" {
			return fmt.Sprintf("haira.SlackSend(%s)", args), true
		}
	case "excel":
		if method == "open" {
			return fmt.Sprintf("haira.ExcelOpen(%s)", args), true
		}
	case "log":
		// log.info/warn/error inside steps → haira.StepLog with injected context
		if activeWorkflowName != "" && activeStepName != "" {
			switch method {
			case "info":
				return fmt.Sprintf("haira.StepLog(%q, %q, \"info\", %s)", activeWorkflowName, activeStepName, args), true
			case "warn":
				return fmt.Sprintf("haira.StepLog(%q, %q, \"warn\", %s)", activeWorkflowName, activeStepName, args), true
			case "error":
				return fmt.Sprintf("haira.StepLog(%q, %q, \"error\", %s)", activeWorkflowName, activeStepName, args), true
			}
		} else {
			// Outside steps: print with level prefix to stdout/stderr
			switch method {
			case "info":
				return fmt.Sprintf("haira.LogPrint(\"info\", %s)", args), true
			case "warn":
				return fmt.Sprintf("haira.LogPrint(\"warn\", %s)", args), true
			case "error":
				return fmt.Sprintf("haira.LogPrint(\"error\", %s)", args), true
			}
		}
	case "time":
		switch method {
		case "sleep":
			return fmt.Sprintf("haira.TimeSleep(%s)", args), true
		case "now":
			return "haira.TimeNow()", true
		case "format":
			return fmt.Sprintf("haira.TimeFormat(%s)", args), true
		case "parse":
			return fmt.Sprintf("haira.TimeParse(%s)", args), true
		case "since":
			return fmt.Sprintf("haira.TimeSince(%s)", args), true
		case "after":
			return fmt.Sprintf("haira.TimeAfter(%s)", args), true
		case "tick":
			return fmt.Sprintf("haira.TimeTick(%s)", args), true
		case "slug":
			return "haira.TimeSlug()", true
		}
	case "fs":
		switch method {
		case "read_file":
			return fmt.Sprintf("haira.FsReadFile(%s)", args), true
		case "write_file":
			return fmt.Sprintf("haira.FsWriteFile(%s)", args), true
		case "append_file":
			return fmt.Sprintf("haira.FsAppendFile(%s)", args), true
		case "exists":
			return fmt.Sprintf("haira.FsExists(%s)", args), true
		case "remove":
			return fmt.Sprintf("haira.FsRemove(%s)", args), true
		case "remove_all":
			return fmt.Sprintf("haira.FsRemoveAll(%s)", args), true
		case "rename":
			return fmt.Sprintf("haira.FsRename(%s)", args), true
		case "copy":
			return fmt.Sprintf("haira.FsCopy(%s)", args), true
		case "mkdir":
			return fmt.Sprintf("haira.FsMkdir(%s)", args), true
		case "mkdir_all":
			return fmt.Sprintf("haira.FsMkdirAll(%s)", args), true
		case "read_dir":
			return fmt.Sprintf("haira.FsReadDir(%s)", args), true
		case "stat":
			return fmt.Sprintf("haira.FsStat(%s)", args), true
		}
	case "observe":
		switch method {
		case "usage":
			return "haira.ObserveUsage()", true
		case "agent_usage":
			return fmt.Sprintf("haira.ObserveAgentUsage(%s)", args), true
		case "session_usage":
			return fmt.Sprintf("haira.ObserveSessionUsage(%s)", args), true
		case "model_usage":
			return fmt.Sprintf("haira.ObserveModelUsage(%s)", args), true
		case "events":
			return "haira.ObserveEvents()", true
		case "agent_events":
			return fmt.Sprintf("haira.ObserveAgentEvents(%s)", args), true
		case "reset":
			return "haira.ObserveReset()", true
		case "start":
			return fmt.Sprintf("haira.ObserveStart(%s)", args), true
		}
	}
	return "", false
}

func resolveServerCall(call ast.CallExpr) string {
	if len(call.Args) == 1 {
		if list, ok := call.Args[0].Value.Node.(ast.ListExpr); ok {
			refs := make([]string, len(list.Elems))
			for i, item := range list.Elems {
				if ident, ok := item.Node.(ast.IdentExpr); ok {
					refs[i] = "workflowDef" + SnakeToPascal(ident.Name)
				} else {
					refs[i] = ExprToGo(item)
				}
			}
			return fmt.Sprintf("haira.NewServer([]*haira.WorkflowDef{%s})", strings.Join(refs, ", "))
		}
	}
	args := callArgsToGo(call.Args)
	return fmt.Sprintf("haira.NewServer(%s)", args)
}

func resolveMCPServerCall(call ast.CallExpr) string {
	if len(call.Args) == 1 {
		if list, ok := call.Args[0].Value.Node.(ast.ListExpr); ok {
			refs := make([]string, len(list.Elems))
			for i, item := range list.Elems {
				if ident, ok := item.Node.(ast.IdentExpr); ok {
					refs[i] = "workflowDef" + SnakeToPascal(ident.Name)
				} else {
					refs[i] = ExprToGo(item)
				}
			}
			return fmt.Sprintf("haira.NewMCPServer([]*haira.WorkflowDef{%s})", strings.Join(refs, ", "))
		}
	}
	args := callArgsToGo(call.Args)
	return fmt.Sprintf("haira.NewMCPServer(%s)", args)
}

func callArgsToGo(args []ast.Argument) string {
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = ExprToGo(a.Value)
	}
	return strings.Join(parts, ", ")
}

// resolveVectorEmbed converts vector.embed(provider, text) → haira.VectorEmbed(providerVar, text)
func resolveVectorEmbed(call ast.CallExpr) string {
	if len(call.Args) < 2 {
		return "haira.VectorEmbed(nil, \"\")"
	}
	providerArg := resolveProviderArg(call.Args[0].Value)
	textArg := ExprToGo(call.Args[1].Value)
	return fmt.Sprintf("haira.VectorEmbed(%s, %s)", providerArg, textArg)
}

// resolveVectorEmbedBatch converts vector.embed_batch(provider, texts) → haira.VectorEmbedBatch(providerVar, texts)
func resolveVectorEmbedBatch(call ast.CallExpr) string {
	if len(call.Args) < 2 {
		return "haira.VectorEmbedBatch(nil, nil)"
	}
	providerArg := resolveProviderArg(call.Args[0].Value)
	textsArg := ExprToGo(call.Args[1].Value)
	return fmt.Sprintf("haira.VectorEmbedBatch(%s, %s)", providerArg, textsArg)
}

// resolveVectorCollection converts vector.collection(db, "name", dimensions: N)
// → haira.VectorNewCollection(db, "name", N)
func resolveVectorCollection(call ast.CallExpr) string {
	if len(call.Args) < 2 {
		return "haira.VectorNewCollection(nil, \"\", 0)"
	}
	dbArg := ExprToGo(call.Args[0].Value)
	nameArg := ExprToGo(call.Args[1].Value)
	// Look for named arg "dimensions"
	dimArg := "0"
	for _, arg := range call.Args[2:] {
		if arg.Name != nil && arg.Name.Node == "dimensions" {
			dimArg = ExprToGo(arg.Value)
			break
		}
	}
	return fmt.Sprintf("haira.VectorNewCollection(%s, %s, %s)", dbArg, nameArg, dimArg)
}

// resolveProviderArg resolves a provider identifier to its Go variable name.
// e.g., "openai_embed" → "providerOpenaiEmbed"
func resolveProviderArg(expr ast.Expr) string {
	if ident, ok := expr.Node.(ast.IdentExpr); ok {
		return goVarName("provider", ident.Name)
	}
	return ExprToGo(expr)
}

// IsStdlibImport returns whether a Haira import path maps to a stdlib module.
func IsStdlibImport(path string) bool {
	switch path {
	case "io", "http", "mcp", "env", "json", "postgres", "slack", "excel", "time",
		"string", "regex", "math", "conv", "array", "map", "log", "ui", "vector",
		"observe", "fs":
		return true
	}
	return false
}
