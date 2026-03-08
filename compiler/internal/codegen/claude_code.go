package codegen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
	"github.com/haira-lang/haira/internal/checker"
)

// ExportClaudeCode generates Claude Code agent config files and an MCP server binary for custom tools.
//
// Output structure:
//
//	.claude/
//	  agents/agent-name.md     ← agent system prompt + tools + config
//	  commands/workflow-name.md ← slash commands for workflows
//	  mcp-servers.json         ← points to MCP binary
//	bin/
//	  haira-tools              ← MCP server binary with custom tools
func ExportClaudeCode(file *ast.SourceFile, output, hairaFile, hairaSource string, typeInfo ...*checker.TypeInfo) error {
	if output == "" {
		output = "."
	}

	claudeDir := filepath.Join(output, ".claude")
	agentsDir := filepath.Join(claudeDir, "agents")
	commandsDir := filepath.Join(claudeDir, "commands")
	binDir := filepath.Join(output, "bin")

	// Create directories
	for _, dir := range []string{agentsDir, commandsDir, binDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create dir %s: %w", dir, err)
		}
	}

	var agents []ast.AgentDecl
	var workflows []ast.WorkflowDecl
	var tools []ast.ToolDecl
	var providers []ast.ProviderDecl

	for _, item := range file.Items {
		switch it := item.Node.(type) {
		case ast.AgentDecl:
			agents = append(agents, it)
		case ast.WorkflowDecl:
			workflows = append(workflows, it)
		case ast.ToolDecl:
			tools = append(tools, it)
		case ast.ProviderDecl:
			providers = append(providers, it)
		}
	}

	// Generate agent markdown files
	for _, agent := range agents {
		if err := writeAgentMarkdown(agentsDir, agent); err != nil {
			return err
		}
	}

	// Generate workflow slash command files
	for _, wf := range workflows {
		if err := writeWorkflowCommand(commandsDir, wf); err != nil {
			return err
		}
	}

	// Generate MCP server + mcp-servers.json if there are custom tools
	if len(tools) > 0 {
		// Build MCP binary from the Haira source
		mcpBinary := filepath.Join(binDir, "haira-tools")
		if err := buildMCPBinary(file, mcpBinary, hairaFile, hairaSource, typeInfo...); err != nil {
			return fmt.Errorf("build MCP binary: %w", err)
		}

		// Write mcp-servers.json
		absBinary, _ := filepath.Abs(mcpBinary)
		if err := writeMCPServersJSON(claudeDir, absBinary, tools); err != nil {
			return err
		}
	}

	// Print summary
	fmt.Fprintf(os.Stderr, "Exported Claude Code config:\n")
	fmt.Fprintf(os.Stderr, "  %d agent(s) → %s/\n", len(agents), agentsDir)
	fmt.Fprintf(os.Stderr, "  %d command(s) → %s/\n", len(workflows), commandsDir)
	if len(tools) > 0 {
		fmt.Fprintf(os.Stderr, "  %d tool(s) → %s/haira-tools (MCP)\n", len(tools), binDir)
	}
	return nil
}

// writeAgentMarkdown generates a Claude Code agent .md file.
func writeAgentMarkdown(dir string, agent ast.AgentDecl) error {
	name := agent.Name.Node
	filename := toKebabCase(name) + ".md"

	var system, model string
	var toolNames []string
	var handoffs []string

	for _, field := range agent.Fields {
		switch field.Key.Node {
		case "system":
			system = extractStringValue(field.Value)
		case "model":
			model = extractStringValue(field.Value)
		case "tools":
			toolNames = extractListStrings(field.Value)
		case "handoffs":
			handoffs = extractListStrings(field.Value)
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", name))

	if model != "" {
		sb.WriteString(fmt.Sprintf("Model: %s\n\n", model))
	}

	if system != "" {
		sb.WriteString(system)
		sb.WriteString("\n\n")
	}

	if len(toolNames) > 0 {
		sb.WriteString("## Tools\n\n")
		for _, t := range toolNames {
			sb.WriteString(fmt.Sprintf("- %s (via MCP: haira-tools)\n", t))
		}
		sb.WriteString("\n")
	}

	if len(handoffs) > 0 {
		sb.WriteString("## Handoffs\n\n")
		sb.WriteString("This agent can delegate to:\n")
		for _, h := range handoffs {
			sb.WriteString(fmt.Sprintf("- %s\n", h))
		}
		sb.WriteString("\n")
	}

	return os.WriteFile(filepath.Join(dir, filename), []byte(sb.String()), 0o644)
}

// writeWorkflowCommand generates a Claude Code slash command .md file.
func writeWorkflowCommand(dir string, wf ast.WorkflowDecl) error {
	name := wf.Name.Node
	filename := toKebabCase(name) + ".md"

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# /%s\n\n", toKebabCase(name)))

	if wf.Description != "" {
		sb.WriteString(wf.Description)
		sb.WriteString("\n\n")
	}

	if len(wf.Params) > 0 {
		sb.WriteString("## Parameters\n\n")
		for _, p := range wf.Params {
			ty := "any"
			if p.Ty != nil {
				ty = HairaTypeToGo(p.Ty.Node)
			}
			sb.WriteString(fmt.Sprintf("- `%s` (%s)\n", p.Name.Node, ty))
		}
		sb.WriteString("\n")
	}

	// Include trigger info
	if wf.Trigger != nil {
		sb.WriteString(fmt.Sprintf("Trigger: @%s\n\n", wf.Trigger.Name.Node))
	}

	sb.WriteString("## Steps\n\n")
	for _, stmt := range wf.Body.Statements {
		if step, ok := stmt.Node.(ast.StepStmt); ok {
			sb.WriteString(fmt.Sprintf("1. **%s**\n", step.Name.Node))
		}
	}

	return os.WriteFile(filepath.Join(dir, filename), []byte(sb.String()), 0o644)
}

// buildMCPBinary compiles the Haira source into an MCP server binary
// that exposes custom tools via stdio transport.
func buildMCPBinary(file *ast.SourceFile, output, hairaFile, hairaSource string, typeInfo ...*checker.TypeInfo) error {
	// Generate MCP server Go code
	mcpGo := generateMCPServerGo(file, hairaFile, hairaSource, typeInfo...)

	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("haira-mcp-%d", os.Getpid()))
	usedPkgs := collectUsedStdlibPackages(file)

	if err := writeProject(tmpDir, mcpGo, usedPkgs); err != nil {
		return err
	}
	if err := runGoModTidy(tmpDir); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}
	if err := runGoBuild(tmpDir, output); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	os.RemoveAll(tmpDir)
	return nil
}

// generateMCPServerGo generates Go source for an MCP server binary exposing tools.
func generateMCPServerGo(file *ast.SourceFile, sourceFile, sourceText string, typeInfo ...*checker.TypeInfo) string {
	em := NewEmitter()
	em.sourceFile = sourceFile
	em.sourceText = sourceText
	if len(typeInfo) > 0 && typeInfo[0] != nil {
		em.typeInfo = typeInfo[0]
		activeTypeInfo = typeInfo[0]
	} else {
		activeTypeInfo = nil
	}

	activeSourceFile = file

	knownToolNames = make(map[string]bool)
	for _, item := range file.Items {
		if t, ok := item.Node.(ast.ToolDecl); ok {
			knownToolNames[t.Name.Node] = true
		}
	}

	em.Line("package main")
	em.Blank()

	// Collect tool names for MCP registration
	var toolNames []string
	for _, item := range file.Items {
		if t, ok := item.Node.(ast.ToolDecl); ok {
			toolNames = append(toolNames, t.Name.Node)
		}
	}

	em.Line("import (")
	em.Line(`	"fmt"`)
	em.Line(`	"os"`)
	em.Line(`	haira "haira-generated/haira"`)

	// Check for stdlib imports
	for _, item := range file.Items {
		if imp, ok := item.Node.(ast.ImportDecl); ok {
			if goImport, ok := stdlibGoImport(imp.Path); ok {
				em.Line(fmt.Sprintf(`	%q`, "haira-generated/"+goImport))
			}
		}
	}

	em.Line(")")
	em.Blank()

	em.Line("var _ = fmt.Sprintf")
	em.Line("var _ = os.Exit")
	em.Blank()

	// Emit providers
	for _, item := range file.Items {
		if p, ok := item.Node.(ast.ProviderDecl); ok {
			EmitProvider(em, p)
		}
	}

	// Emit tools
	for _, item := range file.Items {
		if t, ok := item.Node.(ast.ToolDecl); ok {
			EmitTool(em, t)
		}
	}

	// Emit helper functions (non-main)
	for _, item := range file.Items {
		if f, ok := item.Node.(ast.FunctionDef); ok {
			if f.Name.Node != "main" {
				emitFunction(em, f)
			}
		}
	}

	// Generate main that starts MCP server
	em.OpenBlock("func main()")
	em.Line("tools := []haira.ToolDef{}")
	for _, name := range toolNames {
		funcName := fmt.Sprintf("hairaTool_%s", name)
		em.Line(fmt.Sprintf("tools = append(tools, %s)", funcName))
	}
	em.Blank()
	em.Line(`server := haira.NewMCPServer("haira-tools", tools)`)
	em.Line("server.ServeStdio()")
	em.CloseBlock()

	return em.String()
}

// writeMCPServersJSON writes the .claude/mcp-servers.json config.
func writeMCPServersJSON(claudeDir, binaryPath string, tools []ast.ToolDecl) error {
	config := map[string]any{
		"haira-tools": map[string]any{
			"command": binaryPath,
			"args":    []string{},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(claudeDir, "mcp-servers.json"), data, 0o644)
}

// --- helpers ---

func toKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteByte('-')
			}
			result.WriteByte(byte(r + 32))
		} else if r == '_' {
			result.WriteByte('-')
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func extractStringValue(expr ast.Expr) string {
	if lit, ok := expr.Node.(ast.LiteralExpr); ok {
		if s, ok := lit.Lit.(ast.StringLit); ok {
			return s.Value
		}
	}
	return ""
}

func extractListStrings(expr ast.Expr) []string {
	if list, ok := expr.Node.(ast.ListExpr); ok {
		var result []string
		for _, elem := range list.Elems {
			if ident, ok := elem.Node.(ast.IdentExpr); ok {
				result = append(result, ident.Name)
			}
		}
		return result
	}
	return nil
}

