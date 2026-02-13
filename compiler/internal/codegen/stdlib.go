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
	case "json":
		switch method {
		case "marshal":
			return fmt.Sprintf("haira.JSONMarshal(%s)", args), true
		case "unmarshal":
			return fmt.Sprintf("haira.JSONUnmarshal(%s)", args), true
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
	case "time":
		switch method {
		case "sleep":
			return fmt.Sprintf("haira.TimeSleep(%s)", args), true
		case "now":
			return "haira.TimeNow()", true
		case "slug":
			return "haira.TimeSlug()", true
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

func callArgsToGo(args []ast.Argument) string {
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = ExprToGo(a.Value)
	}
	return strings.Join(parts, ", ")
}

// IsStdlibImport returns whether a Haira import path maps to a stdlib module.
func IsStdlibImport(path string) bool {
	switch path {
	case "io", "http", "env", "json", "postgres", "slack", "excel", "time":
		return true
	}
	return false
}
