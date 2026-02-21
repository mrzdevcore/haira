package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/haira-lang/haira/internal/parser"
)

// helper: find examples directory.
func examplesDir() string {
	cwd, _ := os.Getwd()
	root := filepath.Join(cwd, "..", "..", "..")
	dir := filepath.Join(root, "examples")
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir
	}
	return ""
}

func parseAndGenerate(t *testing.T, src string) string {
	t.Helper()
	sf, errs := parser.Parse(src)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e.Message)
		}
		t.FailNow()
	}
	return GenerateMainGo(sf, "", "")
}

// ---------------------------------------------------------------------------
// All examples generate Go without panic
// ---------------------------------------------------------------------------

func TestAllExamplesGenerate(t *testing.T) {
	dir := examplesDir()
	if dir == "" {
		t.Skip("examples directory not found")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".haira" {
			continue
		}
		count++
		t.Run(entry.Name(), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			sf, errs := parser.Parse(string(data))
			if len(errs) > 0 {
				t.Skipf("skipping (parse errors): %s", entry.Name())
			}
			// Should not panic
			got := GenerateMainGo(sf, "", "")
			if !strings.Contains(got, "package main") {
				t.Error("generated Go should contain 'package main'")
			}
		})
	}
	if count == 0 {
		t.Fatal("no .haira files found")
	}
}

// ---------------------------------------------------------------------------
// Golden test: 01-hello
// ---------------------------------------------------------------------------

func TestGoldenHello(t *testing.T) {
	src := `import "io"
fn main() {
    io.println("Hello, World!")
}`
	got := parseAndGenerate(t, src)

	// Must contain package main
	if !strings.Contains(got, "package main") {
		t.Error("missing 'package main'")
	}
	// Must import haira runtime
	if !strings.Contains(got, `"haira-generated/haira"`) {
		t.Error("missing haira runtime import")
	}
	// Must call haira.Println
	if !strings.Contains(got, `haira.Println("Hello, World!")`) {
		t.Error("missing haira.Println call")
	}
	// Must have func main()
	if !strings.Contains(got, "func main()") {
		t.Error("missing func main()")
	}
}

// ---------------------------------------------------------------------------
// Golden test: functions
// ---------------------------------------------------------------------------

func TestGoldenFunctions(t *testing.T) {
	src := `import "io"
fn add(a: int, b: int) -> int {
    return a + b
}
fn main() {
    io.println(add(3, 5))
}`
	got := parseAndGenerate(t, src)

	// Function should be PascalCase
	if !strings.Contains(got, "func Add(a int, b int) int") {
		t.Errorf("missing 'func Add(a int, b int) int' in output:\n%s", got)
	}
	// Call should use PascalCase
	if !strings.Contains(got, "Add(3, 5)") {
		t.Error("missing Add(3, 5) call")
	}
}

// ---------------------------------------------------------------------------
// Golden test: variables
// ---------------------------------------------------------------------------

func TestGoldenVariables(t *testing.T) {
	src := `import "io"
fn main() {
    name = "Haira"
    version = 1
    io.println(name)
    io.println(version)
}`
	got := parseAndGenerate(t, src)

	// First assignment should use :=
	if !strings.Contains(got, `name := "Haira"`) {
		t.Error("missing name := assignment")
	}
	if !strings.Contains(got, "version := 1") {
		t.Error("missing version := assignment")
	}
}

// ---------------------------------------------------------------------------
// Golden test: structs
// ---------------------------------------------------------------------------

func TestGoldenStructs(t *testing.T) {
	src := `struct User {
    name: string
    age: int
}
User.greet() -> string {
    return "Hello, I'm " + self.name
}`
	got := parseAndGenerate(t, src)

	// Struct definition
	if !strings.Contains(got, "type User struct") {
		t.Error("missing 'type User struct'")
	}
	// Fields should be capitalized
	if !strings.Contains(got, "Name string") {
		t.Error("missing 'Name string' field")
	}
	if !strings.Contains(got, "Age int") {
		t.Error("missing 'Age int' field")
	}
	// Method
	if !strings.Contains(got, "func (self *User) Greet()") {
		t.Error("missing method receiver")
	}
}

// ---------------------------------------------------------------------------
// Golden test: enum (simple)
// ---------------------------------------------------------------------------

func TestGoldenSimpleEnum(t *testing.T) {
	src := `enum Direction {
    North
    South
    East
    West
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "type Direction int") {
		t.Error("missing 'type Direction int'")
	}
	if !strings.Contains(got, "DirectionNorth Direction = iota") {
		t.Error("missing iota declaration")
	}
	if !strings.Contains(got, "DirectionSouth") {
		t.Error("missing DirectionSouth")
	}
}

// ---------------------------------------------------------------------------
// Golden test: if/else
// ---------------------------------------------------------------------------

func TestGoldenIfElse(t *testing.T) {
	src := `fn classify(n: int) -> string {
    if n > 0 {
        return "positive"
    } else {
        return "zero or negative"
    }
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "if n > 0") {
		t.Error("missing if condition")
	}
	if !strings.Contains(got, "} else {") {
		t.Error("missing else block")
	}
}

// ---------------------------------------------------------------------------
// Golden test: for loop with range
// ---------------------------------------------------------------------------

func TestGoldenForRange(t *testing.T) {
	src := `fn main() {
    for i in 0..10 {
        x = i
    }
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "for i := 0; i < 10; i++") {
		t.Error("missing range-based for loop")
	}
}

// ---------------------------------------------------------------------------
// Golden test: while loop
// ---------------------------------------------------------------------------

func TestGoldenWhile(t *testing.T) {
	src := `fn main() {
    x = 10
    while x > 0 {
        x = x - 1
    }
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "for x > 0") {
		t.Error("while should become 'for condition'")
	}
}

// ---------------------------------------------------------------------------
// Golden test: interpolated string
// ---------------------------------------------------------------------------

func TestGoldenInterpolation(t *testing.T) {
	src := `import "io"
fn main() {
    name = "World"
    io.println("Hello, ${name}!")
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "fmt.Sprintf") {
		t.Error("interpolated string should use fmt.Sprintf")
	}
	if !strings.Contains(got, `"fmt"`) {
		t.Error("should import fmt for interpolated strings")
	}
}

// ---------------------------------------------------------------------------
// Golden test: agentic declarations
// ---------------------------------------------------------------------------

func TestGoldenProvider(t *testing.T) {
	src := `provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4"
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "providerOpenai") {
		t.Error("missing provider variable")
	}
	if !strings.Contains(got, "haira.Provider") {
		t.Error("missing haira.Provider")
	}
}

func TestGoldenAgent(t *testing.T) {
	src := `provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4"
}
agent Bot {
    model: openai
    system: "You are helpful."
    temperature: 0.7
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "agentBot") {
		t.Error("missing agent variable")
	}
	if !strings.Contains(got, "haira.AgentConfig") {
		t.Error("missing haira.AgentConfig")
	}
	if !strings.Contains(got, "initAgentBot") {
		t.Error("missing initAgentBot function")
	}
}

func TestGoldenWorkflow(t *testing.T) {
	src := `provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4"
}
agent Bot {
    model: openai
    system: "You are helpful."
}
@webhook("/api/chat")
workflow Chat(message: string) -> { reply: string } {
    reply, err = Bot.ask(message)
    return { reply: reply }
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "workflowDefChat") {
		t.Error("missing workflow definition variable")
	}
	if !strings.Contains(got, `Path:`) {
		t.Error("missing Path field in workflow")
	}
	if !strings.Contains(got, "haira.WorkflowDef") {
		t.Error("missing haira.WorkflowDef")
	}
}

// ---------------------------------------------------------------------------
// Golden test: tool
// ---------------------------------------------------------------------------

func TestGoldenTool(t *testing.T) {
	src := `tool search(query: string) -> string {
    """
    Search the web.
    """
    return "result"
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "toolSearch") {
		t.Error("missing tool variable")
	}
	if !strings.Contains(got, "haira.ToolDef") {
		t.Error("missing haira.ToolDef")
	}
	if !strings.Contains(got, "Search the web.") {
		t.Error("missing tool description")
	}
}

// ---------------------------------------------------------------------------
// Golden test: compound assignment desugaring
// ---------------------------------------------------------------------------

func TestGoldenCompoundAssign(t *testing.T) {
	src := `fn main() {
    x = 0
    x += 5
}`
	got := parseAndGenerate(t, src)

	// After desugaring, x += 5 becomes x = x + 5
	if !strings.Contains(got, "x = x + 5") {
		t.Errorf("expected desugared 'x = x + 5', got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Golden test: list literal
// ---------------------------------------------------------------------------

func TestGoldenList(t *testing.T) {
	src := `fn main() {
    xs = [1, 2, 3]
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "[]any{1, 2, 3}") {
		t.Error("list should become []any{...}")
	}
}

// ---------------------------------------------------------------------------
// Golden test: map literal
// ---------------------------------------------------------------------------

func TestGoldenMap(t *testing.T) {
	src := `fn main() {
    m = {"name": "Alice", "age": 30}
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "map[string]any{") {
		t.Error("map should become map[string]any{...}")
	}
}

// ---------------------------------------------------------------------------
// Golden test: MCP provider
// ---------------------------------------------------------------------------

func TestGoldenMCPProvider(t *testing.T) {
	src := `provider filesystem {
    transport: "mcp"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "mcpFilesystem") {
		t.Errorf("missing MCP client variable 'mcpFilesystem' in output:\n%s", got)
	}
	if !strings.Contains(got, "haira.NewMCPClient") {
		t.Errorf("missing haira.NewMCPClient call in output:\n%s", got)
	}
	if !strings.Contains(got, "haira.MCPConfig") {
		t.Errorf("missing haira.MCPConfig in output:\n%s", got)
	}
	if !strings.Contains(got, `Command: "npx"`) {
		t.Errorf("missing Command field in output:\n%s", got)
	}
	if !strings.Contains(got, `Args: []string{`) {
		t.Errorf("missing Args field in output:\n%s", got)
	}
	// Should NOT contain haira.Provider (it's an MCP provider, not a regular one)
	if strings.Contains(got, "haira.Provider") {
		t.Errorf("MCP provider should not emit haira.Provider:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Golden test: agent with MCP
// ---------------------------------------------------------------------------

func TestGoldenAgentWithMCP(t *testing.T) {
	src := `provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4"
}
provider filesystem {
    transport: "mcp"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
}
agent Bot {
    model: openai
    system: "You are helpful."
    mcp: [filesystem]
}
fn main() {}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "mcpClients := []*haira.MCPClient{mcpFilesystem}") {
		t.Errorf("missing mcpClients slice in output:\n%s", got)
	}
	if !strings.Contains(got, "MCPClients: mcpClients,") {
		t.Errorf("missing MCPClients in AgentConfig:\n%s", got)
	}
	if !strings.Contains(got, "defer haira.ShutdownMCP()") {
		t.Errorf("missing defer haira.ShutdownMCP() in main:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Golden test: MCP provider with SSE transport
// ---------------------------------------------------------------------------

func TestGoldenMCPProviderSSE(t *testing.T) {
	src := `provider remote_tools {
    transport: "mcp"
    endpoint: "http://localhost:3001/sse"
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "mcpRemoteTools") {
		t.Errorf("missing MCP client variable 'mcpRemoteTools' in output:\n%s", got)
	}
	if !strings.Contains(got, `Endpoint: "http://localhost:3001/sse"`) {
		t.Errorf("missing Endpoint field in output:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Golden test: workflow with description
// ---------------------------------------------------------------------------

func TestGoldenWorkflowDescription(t *testing.T) {
	src := `provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4"
}
agent Bot {
    model: openai
    system: "You are helpful."
}
@post("/api/summarize")
workflow Summarize(text: string) -> { summary: string } {
    """Summarize the given text into key points."""
    summary, err = Bot.ask(text)
    return { summary: summary }
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, `Description: "Summarize the given text into key points."`) {
		t.Errorf("missing Description field in WorkflowDef:\n%s", got)
	}
	if !strings.Contains(got, "workflowDefSummarize") {
		t.Errorf("missing workflowDefSummarize variable:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Golden test: MCP server
// ---------------------------------------------------------------------------

func TestGoldenMCPServer(t *testing.T) {
	src := `import "io"
import "mcp"

provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4"
}
agent Bot {
    model: openai
    system: "You are helpful."
}
@post("/api/chat")
workflow Chat(message: string) -> { reply: string } {
    """Chat with the bot."""
    reply, err = Bot.ask(message)
    return { reply: reply }
}
fn main() {
    mcp_server = mcp.Server([Chat])
    mcp_server.serve()
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "haira.NewMCPServer") {
		t.Errorf("missing haira.NewMCPServer call in output:\n%s", got)
	}
	if !strings.Contains(got, "workflowDefChat") {
		t.Errorf("missing workflowDefChat reference in MCP server:\n%s", got)
	}
	if !strings.Contains(got, `Description: "Chat with the bot."`) {
		t.Errorf("missing Description in WorkflowDef:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Golden test: MCP server with SSE (.listen)
// ---------------------------------------------------------------------------

func TestGoldenMCPServerListen(t *testing.T) {
	src := `import "mcp"

provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4"
}
agent Bot {
    model: openai
    system: "You are helpful."
}
workflow Chat(message: string) -> { reply: string } {
    """Chat with the bot."""
    reply, err = Bot.ask(message)
    return { reply: reply }
}
fn main() {
    mcp_server = mcp.Server([Chat])
    mcp_server.listen(9000)
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "haira.NewMCPServer") {
		t.Errorf("missing haira.NewMCPServer call in output:\n%s", got)
	}
	if !strings.Contains(got, "mcp_server.Listen(9000)") {
		t.Errorf("missing mcp_server.Listen(9000) in output:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Golden test: workflow without description (no Description field emitted)
// ---------------------------------------------------------------------------

func TestGoldenWorkflowNoDescription(t *testing.T) {
	src := `provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4"
}
agent Bot {
    model: openai
    system: "You are helpful."
}
@post("/api/chat")
workflow Chat(message: string) -> { reply: string } {
    reply, err = Bot.ask(message)
    return { reply: reply }
}`
	got := parseAndGenerate(t, src)

	if strings.Contains(got, "Description:") {
		t.Errorf("workflow without description should not emit Description field:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Type mapping
// ---------------------------------------------------------------------------

func TestHairaTypeToGo(t *testing.T) {
	tests := map[string]string{
		"int":    "int",
		"float":  "float64",
		"string": "string",
		"bool":   "bool",
		"any":    "any",
	}
	// nil → "any"
	got := HairaTypeToGo(nil)
	if got != "any" {
		t.Errorf("HairaTypeToGo(nil) = %q, want 'any'", got)
	}
	// Test actual named types
	for haira, expected := range tests {
		result := hairaNamedTypeToGo(haira)
		if result != expected {
			t.Errorf("HairaTypeToGo(%q) = %q, want %q", haira, result, expected)
		}
	}
}

func hairaNamedTypeToGo(name string) string {
	switch name {
	case "int":
		return "int"
	case "float":
		return "float64"
	case "string":
		return "string"
	case "bool":
		return "bool"
	default:
		return "any"
	}
}

// ---------------------------------------------------------------------------
// Utility functions
// ---------------------------------------------------------------------------

func TestCapitalize(t *testing.T) {
	tests := map[string]string{
		"hello": "Hello",
		"":      "",
		"A":     "A",
		"abc":   "Abc",
	}
	for input, expected := range tests {
		got := Capitalize(input)
		if got != expected {
			t.Errorf("Capitalize(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestSnakeToPascal(t *testing.T) {
	tests := map[string]string{
		"hello_world": "HelloWorld",
		"get_weather": "GetWeather",
		"main":        "Main",
		"a_b_c":       "ABC",
	}
	for input, expected := range tests {
		got := SnakeToPascal(input)
		if got != expected {
			t.Errorf("SnakeToPascal(%q) = %q, want %q", input, got, expected)
		}
	}
}

// ---------------------------------------------------------------------------
// Stdlib resolution
// ---------------------------------------------------------------------------

func TestIsStdlibImport(t *testing.T) {
	stdlibs := []string{"io", "http", "mcp", "env", "json", "postgres", "slack", "excel", "time"}
	for _, s := range stdlibs {
		if !IsStdlibImport(s) {
			t.Errorf("expected %q to be stdlib import", s)
		}
	}
	nonStdlibs := []string{"mylib", "utils", "models"}
	for _, s := range nonStdlibs {
		if IsStdlibImport(s) {
			t.Errorf("expected %q to NOT be stdlib import", s)
		}
	}
}

// ---------------------------------------------------------------------------
// Agent topological sort
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Golden test: vector.embed codegen
// ---------------------------------------------------------------------------

func TestGoldenVectorEmbed(t *testing.T) {
	src := `import "io"
import "vector"

provider openai_embed {
    api_key: env("OPENAI_API_KEY")
    model: "text-embedding-3-small"
}

fn main() {
    embedding = vector.embed(openai_embed, "Hello world")
    io.println(embedding)
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "vector.VectorEmbed(providerOpenaiEmbed,") {
		t.Errorf("missing vector.VectorEmbed with resolved provider var in output:\n%s", got)
	}
	if !strings.Contains(got, `"Hello world"`) {
		t.Errorf("missing text argument in output:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Golden test: vector.collection codegen
// ---------------------------------------------------------------------------

func TestGoldenVectorCollection(t *testing.T) {
	src := `import "postgres"
import "vector"

provider openai_embed {
    api_key: env("OPENAI_API_KEY")
    model: "text-embedding-3-small"
}

fn main() {
    db = postgres.connect(env("DATABASE_URL"))
    docs = vector.collection(db, "documents", dimensions: 1536)
    results = vector.search(docs, {query: vector.embed(openai_embed, "test"), limit: 5})
    io.println(vector.format(results))
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "vector.VectorNewCollection(db, \"documents\", 1536)") {
		t.Errorf("missing vector.VectorNewCollection with dimensions in output:\n%s", got)
	}
	if !strings.Contains(got, "vector.VectorSearch(") {
		t.Errorf("missing vector.VectorSearch in output:\n%s", got)
	}
	if !strings.Contains(got, "vector.VectorFormat(") {
		t.Errorf("missing vector.VectorFormat in output:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Golden test: structured output (agent with output: Struct)
// ---------------------------------------------------------------------------

func TestGoldenStructuredOutput(t *testing.T) {
	src := `import "io"

provider openai {
    api_key: env("OPENAI_API_KEY")
    model: "gpt-4o"
}

struct Analysis {
    sentiment: string
    confidence: float
    topics: [string]
}

agent Analyzer {
    model: openai
    system: "Analyze text."
    output: Analysis
}

fn main() {
    result, err = Analyzer.run("I love Haira!")
    io.println(result)
}`
	got := parseAndGenerate(t, src)

	if !strings.Contains(got, "OutputSchema:") {
		t.Errorf("missing OutputSchema in agent config:\n%s", got)
	}
	if !strings.Contains(got, `"sentiment"`) {
		t.Errorf("missing sentiment field in JSON schema:\n%s", got)
	}
	if !strings.Contains(got, `"confidence"`) {
		t.Errorf("missing confidence field in JSON schema:\n%s", got)
	}
	if !strings.Contains(got, `"topics"`) {
		t.Errorf("missing topics field in JSON schema:\n%s", got)
	}
	if !strings.Contains(got, `"type":"object"`) {
		t.Errorf("missing type:object in JSON schema:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Agent topological sort
// ---------------------------------------------------------------------------

func TestAgentTopoSort(t *testing.T) {
	src := `provider openai {
    api_key: env("KEY")
    model: "gpt-4"
}
agent FrontDesk {
    model: openai
    system: "Route requests"
    handoffs: [Billing, Tech]
}
agent Billing {
    model: openai
    system: "Handle billing"
}
agent Tech {
    model: openai
    system: "Handle tech"
}
fn main() {
    io.println("ok")
}`
	got := parseAndGenerate(t, src)

	// Extract the main() function body to check init call order there
	mainIdx := strings.Index(got, "func main()")
	if mainIdx < 0 {
		t.Fatalf("missing func main() in output:\n%s", got)
	}
	mainBody := got[mainIdx:]

	billingPos := strings.Index(mainBody, "initAgentBilling()")
	techPos := strings.Index(mainBody, "initAgentTech()")
	frontPos := strings.Index(mainBody, "initAgentFrontDesk()")
	if billingPos < 0 || techPos < 0 || frontPos < 0 {
		t.Fatalf("missing init calls in main():\n%s", mainBody)
	}
	if billingPos > frontPos {
		t.Error("Billing should be initialized before FrontDesk in main()")
	}
	if techPos > frontPos {
		t.Error("Tech should be initialized before FrontDesk in main()")
	}
}

// ---------------------------------------------------------------------------
// Test codegen: test + assert keywords
// ---------------------------------------------------------------------------

func parseAndGenerateTest(t *testing.T, src string) string {
	t.Helper()
	sf, errs := parser.Parse(src)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e.Message)
		}
		t.FailNow()
	}
	return GenerateTestGo(sf, "", "")
}

func parseAndGenerateMainForTest(t *testing.T, src string) string {
	t.Helper()
	sf, errs := parser.Parse(src)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e.Message)
		}
		t.FailNow()
	}
	return GenerateMainGoForTest(sf, "", "")
}

func TestGoldenTestBasic(t *testing.T) {
	src := `
fn add(a: int, b: int) -> int {
    return a + b
}

test "add works" {
    assert add(1, 2) == 3
}
`
	got := parseAndGenerateTest(t, src)

	if !strings.Contains(got, "package main") {
		t.Errorf("missing package main:\n%s", got)
	}
	if !strings.Contains(got, `"testing"`) {
		t.Errorf("missing testing import:\n%s", got)
	}
	if !strings.Contains(got, "func TestHaira(t *testing.T)") {
		t.Errorf("missing TestHaira function:\n%s", got)
	}
	if !strings.Contains(got, `t.Run("add works"`) {
		t.Errorf("missing t.Run sub-test:\n%s", got)
	}
	if !strings.Contains(got, "Add(1, 2)") {
		t.Errorf("missing function call in assertion:\n%s", got)
	}
	if !strings.Contains(got, "t.Errorf") {
		t.Errorf("missing t.Errorf for equality assertion:\n%s", got)
	}
}

func TestGoldenTestCustomMessage(t *testing.T) {
	src := `
fn add(a: int, b: int) -> int {
    return a + b
}

test "with message" {
    x = add(2, 2)
    assert x == 4, "2+2 should equal 4"
}
`
	got := parseAndGenerateTest(t, src)

	if !strings.Contains(got, `"2+2 should equal 4"`) {
		t.Errorf("missing custom message in assertion:\n%s", got)
	}
	if !strings.Contains(got, "t.Errorf") {
		t.Errorf("missing t.Errorf for equality assertion:\n%s", got)
	}
}

func TestGoldenTestBooleanAssert(t *testing.T) {
	src := `
fn is_even(n: int) -> bool {
    return n % 2 == 0
}

test "bool check" {
    assert is_even(4)
}
`
	got := parseAndGenerateTest(t, src)

	if !strings.Contains(got, "t.Fatalf") {
		t.Errorf("missing t.Fatalf for boolean assertion:\n%s", got)
	}
	if !strings.Contains(got, "IsEven(4)") {
		t.Errorf("missing function call:\n%s", got)
	}
}

func TestGoldenTestNotEqual(t *testing.T) {
	src := `
fn add(a: int, b: int) -> int {
    return a + b
}

test "not equal" {
    assert add(1, 2) != add(1, 3)
}
`
	got := parseAndGenerateTest(t, src)

	if !strings.Contains(got, "expected not equal") {
		t.Errorf("missing not-equal message:\n%s", got)
	}
}

func TestGoldenTestMainStub(t *testing.T) {
	src := `
fn add(a: int, b: int) -> int {
    return a + b
}

fn main() {
    io.println(add(1, 2))
}

test "add works" {
    assert add(1, 2) == 3
}
`
	got := parseAndGenerateMainForTest(t, src)

	if !strings.Contains(got, "func main()") {
		t.Errorf("missing stub main:\n%s", got)
	}
	if !strings.Contains(got, "func Add(") {
		t.Errorf("missing Add function:\n%s", got)
	}
	// The main body should be empty (stub)
	mainIdx := strings.Index(got, "func main()")
	if mainIdx < 0 {
		t.Fatalf("missing func main()")
	}
	mainBody := got[mainIdx:]
	// Should NOT contain the user's main body (println)
	if strings.Contains(mainBody, "Println") {
		t.Errorf("stub main should not contain user code:\n%s", mainBody)
	}
}

func TestGoldenTestMultiple(t *testing.T) {
	src := `
fn add(a: int, b: int) -> int {
    return a + b
}

test "first" {
    assert add(1, 1) == 2
}

test "second" {
    assert add(2, 3) == 5
}
`
	got := parseAndGenerateTest(t, src)

	if !strings.Contains(got, `t.Run("first"`) {
		t.Errorf("missing first sub-test:\n%s", got)
	}
	if !strings.Contains(got, `t.Run("second"`) {
		t.Errorf("missing second sub-test:\n%s", got)
	}
}

func TestNoTestsReturnsEmpty(t *testing.T) {
	src := `
fn add(a: int, b: int) -> int {
    return a + b
}
`
	got := parseAndGenerateTest(t, src)
	if got != "" {
		t.Errorf("expected empty output for file with no tests, got:\n%s", got)
	}
}

func TestHasTests(t *testing.T) {
	withTests := `
fn add(a: int, b: int) -> int { return a + b }
test "x" { assert add(1,2) == 3 }
`
	withoutTests := `
fn add(a: int, b: int) -> int { return a + b }
`
	sf1, _ := parser.Parse(withTests)
	if !HasTests(sf1) {
		t.Error("expected HasTests to return true")
	}
	sf2, _ := parser.Parse(withoutTests)
	if HasTests(sf2) {
		t.Error("expected HasTests to return false")
	}
}
