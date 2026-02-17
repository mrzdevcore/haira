package formatter

import (
	"github.com/haira-lang/haira/internal/ast"
)

// formatSourceFile is the entry point for formatting an entire file.
// It groups imports together and separates other items with blank lines.
func (f *Formatter) formatSourceFile(sf *ast.SourceFile) {
	lastWasImport := false
	for i, item := range sf.Items {
		isImport := false
		if _, ok := item.Node.(ast.ImportDecl); ok {
			isImport = true
		}

		f.emitCommentsBefore(item.Span.Start)

		// Insert blank line between non-import items, or between
		// an import group and a non-import item.
		if i > 0 && (!isImport || !lastWasImport) {
			f.blank()
		}

		f.formatItem(item)
		lastWasImport = isImport
	}
	f.emitRemainingComments()
}

// formatItem formats a single top-level item.
func (f *Formatter) formatItem(item ast.Item) {
	switch it := item.Node.(type) {
	case ast.ImportDecl:
		f.formatImport(it)
	case ast.ExportDecl:
		f.formatExport(it)
	case ast.FunctionDef:
		f.formatFunction(it)
	case ast.MethodDef:
		f.formatMethod(it)
	case ast.TypeDef:
		f.formatStruct(it)
	case ast.EnumDef:
		f.formatEnum(it)
	case ast.TypeAlias:
		f.formatTypeAlias(it)
	case ast.ProviderDecl:
		f.formatProvider(it)
	case ast.AgentDecl:
		f.formatAgent(it)
	case ast.ToolDecl:
		f.formatTool(it)
	case ast.WorkflowDecl:
		f.formatWorkflow(it)
	case ast.TestDecl:
		f.formatTest(it)
	case ast.ItemStatement:
		f.formatStatement(it.Stmt)
	}
}

// --- Imports ---

func (f *Formatter) formatImport(imp ast.ImportDecl) {
	if imp.IsGlob {
		// import * from "path"
		f.write("import * from ")
		f.write(`"`)
		f.write(imp.Path)
		f.writeln(`"`)
		return
	}
	if len(imp.Names) > 0 {
		// import { Name1, Name2 } from "path"
		f.write("import { ")
		for i, name := range imp.Names {
			if i > 0 {
				f.write(", ")
			}
			f.write(name.Node)
		}
		f.write(" } from ")
		f.write(`"`)
		f.write(imp.Path)
		f.writeln(`"`)
		return
	}
	if imp.Alias != nil {
		// import alias from "path"
		f.write("import ")
		f.write(imp.Alias.Node)
		f.write(" from ")
		f.write(`"`)
		f.write(imp.Path)
		f.writeln(`"`)
		return
	}
	// import "path"
	f.write(`import "`)
	f.write(imp.Path)
	f.writeln(`"`)
}

func (f *Formatter) formatExport(exp ast.ExportDecl) {
	f.write("export { ")
	for i, name := range exp.Names {
		if i > 0 {
			f.write(", ")
		}
		f.write(name.Node)
	}
	f.writeln(" }")
}

// --- Functions & Methods ---

func (f *Formatter) formatFunction(fn ast.FunctionDef) {
	if fn.IsPublic {
		f.write("pub ")
	}
	f.write("fn ")
	f.write(fn.Name.Node)
	f.write("(")
	f.formatParams(fn.Params)
	f.write(")")
	f.formatTypeAnnotation(fn.ReturnTy)
	f.write(" ")
	f.formatBlock(fn.Body)
	f.newline()
}

func (f *Formatter) formatMethod(m ast.MethodDef) {
	f.write(m.TypeName.Node)
	f.write(".")
	f.write(m.Name.Node)
	f.write("(")
	f.formatParams(m.Params)
	f.write(")")
	f.formatTypeAnnotation(m.ReturnTy)
	f.write(" ")
	f.formatBlock(m.Body)
	f.newline()
}

func (f *Formatter) formatParams(params []ast.Param) {
	for i, p := range params {
		if i > 0 {
			f.write(", ")
		}
		if p.IsRest {
			f.write("...")
		}
		f.write(p.Name.Node)
		f.formatColonType(p.Ty)
		if p.Default != nil {
			f.write(" = ")
			f.formatExpr(*p.Default)
		}
	}
}

// --- Type definitions ---

func (f *Formatter) formatStruct(td ast.TypeDef) {
	if td.IsPublic {
		f.write("pub ")
	}
	f.write("struct ")
	f.write(td.Name.Node)
	if len(td.Fields) == 0 {
		f.writeln(" {}")
		return
	}
	f.writeln(" {")
	f.incIndent()
	for i, field := range td.Fields {
		f.writeIndent()
		f.write(field.Name.Node)
		f.formatColonType(field.Ty)
		if field.Default != nil {
			f.write(" = ")
			f.formatExpr(*field.Default)
		}
		if i < len(td.Fields)-1 {
			f.write(",")
		}
		f.newline()
	}
	f.decIndent()
	f.writeln("}")
}

func (f *Formatter) formatEnum(ed ast.EnumDef) {
	if ed.IsPublic {
		f.write("pub ")
	}
	f.write("enum ")
	f.write(ed.Name.Node)
	if len(ed.Variants) == 0 {
		f.writeln(" {}")
		return
	}
	f.writeln(" {")
	f.incIndent()
	for i, v := range ed.Variants {
		f.writeIndent()
		f.write(v.Name.Node)
		if len(v.Fields) > 0 {
			f.write("(")
			f.formatParams(v.Fields)
			f.write(")")
		}
		if i < len(ed.Variants)-1 {
			f.write(",")
		}
		f.newline()
	}
	f.decIndent()
	f.writeln("}")
}

func (f *Formatter) formatTypeAlias(ta ast.TypeAlias) {
	f.write("type ")
	f.write(ta.Name.Node)
	f.write(" = ")
	f.formatType(ta.Ty.Node)
	f.newline()
}

// --- Agentic declarations ---

func (f *Formatter) formatProvider(pd ast.ProviderDecl) {
	f.write("provider ")
	f.write(pd.Name.Node)
	if len(pd.Fields) == 0 {
		f.writeln(" {}")
		return
	}
	f.writeln(" {")
	f.incIndent()
	for _, field := range pd.Fields {
		f.writeIndent()
		f.write(field.Key.Node)
		f.write(": ")
		f.formatExpr(field.Value)
		f.newline()
	}
	f.decIndent()
	f.writeln("}")
}

func (f *Formatter) formatAgent(ad ast.AgentDecl) {
	f.write("agent ")
	f.write(ad.Name.Node)
	if len(ad.Fields) == 0 {
		f.writeln(" {}")
		return
	}
	f.writeln(" {")
	f.incIndent()
	for _, field := range ad.Fields {
		f.writeIndent()
		f.write(field.Key.Node)
		f.write(": ")
		f.formatExpr(field.Value)
		f.newline()
	}
	f.decIndent()
	f.writeln("}")
}

func (f *Formatter) formatTool(td ast.ToolDecl) {
	for _, dec := range td.Decorators {
		f.formatDecorator(dec)
		f.newline()
	}
	f.write("tool ")
	f.write(td.Name.Node)
	f.write("(")
	f.formatParams(td.Params)
	f.write(")")
	f.formatTypeAnnotation(td.ReturnTy)
	f.writeln(" {")
	f.incIndent()
	// Triple-quoted description
	if td.Description != "" {
		f.writeIndent()
		f.writeln(`"""`)
		lines := splitLines(td.Description)
		for _, line := range lines {
			f.writeIndent()
			f.writeln(trimIndent(line))
		}
		f.writeIndent()
		f.writeln(`"""`)
	}
	// Tool body
	if td.Body != nil {
		f.formatBlockStatements(td.Body.Statements)
	}
	f.decIndent()
	f.writeln("}")
}

func (f *Formatter) formatWorkflow(wd ast.WorkflowDecl) {
	// Non-trigger decorators
	for _, dec := range wd.Decorators {
		f.formatDecorator(dec)
		f.newline()
	}
	// Trigger decorator
	if wd.Trigger != nil {
		f.formatDecorator(*wd.Trigger)
		f.newline()
	}
	f.write("workflow ")
	f.write(wd.Name.Node)
	f.write("(")
	f.formatParams(wd.Params)
	f.write(")")
	f.formatTypeAnnotation(wd.ReturnTy)
	f.writeln(" {")
	f.incIndent()
	// Description
	if wd.Description != "" {
		f.writeIndent()
		f.writeln(`"""`)
		lines := splitLines(wd.Description)
		for _, line := range lines {
			f.writeIndent()
			f.writeln(trimIndent(line))
		}
		f.writeIndent()
		f.writeln(`"""`)
	}
	// Lifecycle hooks go first (canonical: declare error handlers at top)
	for i, hook := range wd.Hooks {
		f.formatLifecycleHook(hook)
		if i < len(wd.Hooks)-1 || len(wd.Body.Statements) > 0 {
			f.newline()
		}
	}
	f.formatBlockStatements(wd.Body.Statements)
	f.decIndent()
	f.writeln("}")
}

func (f *Formatter) formatTest(td ast.TestDecl) {
	f.write("test ")
	f.write(`"`)
	f.write(td.Name.Node)
	f.write(`"`)
	f.write(" ")
	f.formatBlock(td.Body)
	f.newline()
}

// trimIndent removes leading whitespace from a string.
func trimIndent(s string) string {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	return s[i:]
}

// splitLines splits a string into lines, trimming trailing empty lines.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	// Trim trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
