# 13. Compiler Architecture

## 13.1 Overview

The Haira compiler transforms high-level, intent-based source code into optimized native binaries. It consists of multiple phases, each with a specific responsibility.

```
┌─────────────────────────────────────────────────────────────────┐
│                        Haira Source Files                       │
│                         (.haira files)                          │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     1. LEXICAL ANALYSIS                         │
│                         (Lexer)                                 │
│         Source text → Token stream                              │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     2. SYNTACTIC ANALYSIS                       │
│                         (Parser)                                │
│         Token stream → Abstract Syntax Tree (AST)               │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     3. NAME RESOLUTION                          │
│                    (Symbol Resolution)                          │
│         AST → Resolved AST with symbol references               │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     4. AUTO-GENERATION                          │
│                   (Intent Interpreter)                          │
│         Unresolved calls → Generated function stubs             │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     5. TYPE INFERENCE                           │
│                     (Type Checker)                              │
│         Resolved AST → Fully typed AST                          │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     6. HIR GENERATION                           │
│              (High-level Intermediate Repr)                     │
│         Typed AST → HIR (desugared, normalized)                 │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     7. MIR GENERATION                           │
│               (Mid-level Intermediate Repr)                     │
│         HIR → MIR (control flow graph, SSA form)                │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     8. OPTIMIZATION                             │
│                (Analysis & Transformation)                      │
│         MIR → Optimized MIR                                     │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     9. LLVM IR GENERATION                       │
│                    (Code Generation)                            │
│         Optimized MIR → LLVM IR                                 │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     10. LLVM BACKEND                            │
│                  (Native Code Generation)                       │
│         LLVM IR → Native binary                                 │
└─────────────────────────────────────────────────────────────────┘
```

## 13.2 Phase 1: Lexical Analysis (Lexer)

### Purpose
Convert source text into a stream of tokens.

### Input
Raw source code (UTF-8 text)

### Output
Token stream with:
- Token type (keyword, identifier, operator, literal, etc.)
- Token value
- Source location (file, line, column)

### Implementation

```
Lexer {
    source: string
    position: int
    line: int
    column: int
    tokens: [Token]
}

Token {
    type: TokenType
    value: string
    location: SourceLocation
}

TokenType =
    | Identifier
    | Keyword
    | IntLiteral
    | FloatLiteral
    | StringLiteral
    | Operator
    | Delimiter
    | Newline
    | EOF

SourceLocation {
    file: string
    line: int
    column: int
}
```

### Key Features
- UTF-8 support
- Automatic newline handling (no semicolons)
- String interpolation tokenization
- Error recovery with location tracking

## 13.3 Phase 2: Syntactic Analysis (Parser)

### Purpose
Build an Abstract Syntax Tree (AST) from the token stream.

### Input
Token stream

### Output
Untyped AST

### Implementation Strategy
Recursive descent parser with:
- Pratt parsing for expressions (operator precedence)
- Error recovery and synchronization
- Span tracking for error messages

### AST Node Types

```
// Top-level
Program { declarations: [Declaration] }

Declaration =
    | TypeDef { name, fields: [Field], visibility }
    | FunctionDef { name, params: [Param], body: Block, visibility }
    | MethodDef { type_name, name, params: [Param], body: Block }
    | TypeAlias { name, target: Type }

// Statements
Statement =
    | Assignment { pattern: Pattern, value: Expr }
    | If { condition: Expr, then: Block, else: Option<Block> }
    | For { pattern: Pattern, iterator: Expr, body: Block }
    | While { condition: Expr, body: Block }
    | Match { subject: Expr, arms: [MatchArm] }
    | Return { values: [Expr] }
    | Try { body: Block, catch_var: string, catch_body: Block }
    | ExprStmt { expr: Expr }
    | Block { statements: [Statement] }

// Expressions
Expr =
    | Literal { value: LiteralValue }
    | Identifier { name: string }
    | Binary { left: Expr, op: BinaryOp, right: Expr }
    | Unary { op: UnaryOp, operand: Expr }
    | Call { callee: Expr, args: [Arg] }
    | Member { object: Expr, field: string }
    | Index { object: Expr, index: Expr }
    | Pipe { left: Expr, right: Expr }
    | Lambda { params: [Param], body: Expr | Block }
    | Match { subject: Expr, arms: [MatchArm] }
    | If { condition: Expr, then: Block, else: Block }
    | List { elements: [Expr] }
    | Map { entries: [(Expr, Expr)] }
    | Instance { type_name: string, fields: [InstanceField] }
    | Async { body: Block }
    | Spawn { body: Block }
    | Select { arms: [SelectArm] }
    | Propagate { expr: Expr }  // expr?
```

## 13.4 Phase 3: Name Resolution

### Purpose
- Resolve all identifiers to their definitions
- Build symbol tables
- Detect undefined references
- Identify auto-generation candidates

### Input
Untyped AST from all project files

### Output
- Resolved AST with symbol references
- List of unresolved calls (auto-generation candidates)
- Symbol table

### Implementation

```
SymbolTable {
    scopes: [Scope]
    current: int
}

Scope {
    parent: Option<int>
    symbols: {string: Symbol}
}

Symbol {
    name: string
    kind: SymbolKind
    type: Option<Type>
    location: SourceLocation
    visibility: Visibility
}

SymbolKind =
    | Variable
    | Function
    | Type
    | Field
    | Parameter
```

### Resolution Order
1. Current scope (local variables, parameters)
2. Enclosing scopes
3. Current file (top-level definitions)
4. Other project files
5. Dependencies
6. Standard library
7. Mark as unresolved (auto-generation candidate)

## 13.5 Phase 4: AI-Powered Intent Resolution

### Purpose
Use AI (Claude) to interpret developer intent and generate code implementations. This is the **core innovation** of Haira—AI is not optional, it's fundamental.

### Input
- List of unresolved calls
- Type definitions (schema)
- Project configuration
- Full context (surrounding code, imports, etc.)

### Output
- Generated function definitions (Canonical IR → AST)
- New type definitions (if needed)
- Confidence scores

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    AI INTENT ENGINE                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                    CACHE LAYER                              │ │
│  │   Check: hash(function_name + context) → cached CIR?        │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           │ miss                                 │
│                           ▼                                      │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                  CONTEXT BUILDER                            │ │
│  │   • Types in scope (User, Post, etc.)                       │ │
│  │   • Fields and their types                                  │ │
│  │   • Call site information                                   │ │
│  │   • Project schema (DB, HTTP, etc.)                         │ │
│  │   • Surrounding code for context                            │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           │                                      │
│                           ▼                                      │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                   CLAUDE API                                │ │
│  │   Prompt: "Interpret '{name}' given context..."             │ │
│  │   Response: Canonical IR (CIR) + confidence                 │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           │                                      │
│                           ▼                                      │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                  CIR VALIDATOR                              │ │
│  │   • Verify CIR is well-formed                               │ │
│  │   • Type check against context                              │ │
│  │   • Security validation                                     │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           │                                      │
│                           ▼                                      │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                  CIR → AST CONVERTER                        │ │
│  │   Canonical IR → Haira AST nodes                            │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           │                                      │
│                           ▼                                      │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                  CACHE UPDATE                               │ │
│  │   Store: hash → CIR + metadata                              │ │
│  │   Update: haira.lock                                        │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### How AI Interprets Intent

Given this code:
```haira
User { name, email, activity_log: [Activity] }
Activity { type, timestamp, details }

summary = summarize_user_activity(user)
```

The AI receives:
```json
{
  "function": "summarize_user_activity",
  "arguments": [{"name": "user", "type": "User"}],
  "context": {
    "types": {
      "User": {"name": "string", "email": "string", "activity_log": "[Activity]"},
      "Activity": {"type": "string", "timestamp": "datetime", "details": "string"}
    }
  }
}
```

And returns Canonical IR:
```json
{
  "function": "summarize_user_activity",
  "params": [{"name": "user", "type": "User"}],
  "returns": "ActivitySummary",
  "new_types": [{
    "name": "ActivitySummary",
    "fields": [
      {"name": "total", "type": "int"},
      {"name": "most_common_type", "type": "string"},
      {"name": "last_active", "type": "datetime"}
    ]
  }],
  "body": [
    {"op": "get_field", "source": "user", "field": "activity_log", "result": "activities"},
    {"op": "count", "source": "activities", "result": "total"},
    {"op": "group_by", "source": "activities", "key": "type", "result": "grouped"},
    {"op": "max_by_count", "source": "grouped", "result": "most_common"},
    {"op": "max_by", "source": "activities", "key": "timestamp", "result": "last"},
    {"op": "construct", "type": "ActivitySummary", "fields": {...}, "result": "return"}
  ],
  "confidence": 0.95
}
```

### Canonical IR (CIR) Specification

CIR is a JSON-based intermediate representation that AI outputs. It's:
- **Deterministic** - Can be reliably converted to AST
- **Type-safe** - All operations have defined type semantics
- **Sandboxed** - Only allowed operations, no arbitrary code

See [Chapter 15: AI Integration](15-ai-integration.md) for full CIR specification.

### Confidence Handling

```
HIGH (≥0.9):    Compile normally
MEDIUM (0.7-0.9): Compile with info message
LOW (0.5-0.7):    Compile with warning
FAILED (<0.5):    Compilation error
```

### Caching for Reproducibility

```
.haira-cache/
└── ai/
    ├── index.json              # Hash → file mapping
    ├── abc123.cir              # Cached CIR
    ├── def456.cir
    └── ...

haira.lock:
[ai_generated]
summarize_user_activity = "sha256:abc123..."
calculate_engagement = "sha256:def456..."
```

### Pattern Shortcuts (Optimization)

For common patterns, the compiler can skip AI and use built-in generators:

```
// These patterns are so common, no AI needed
get_users()                 → Built-in: fetch all
get_user_by_id(id)         → Built-in: fetch by key
save_user(user)            → Built-in: upsert
filter_active              → Built-in: filter by bool field
sort_by_name               → Built-in: sort by field
```

AI is called for anything not matching these patterns.

## 13.6 Phase 5: Type Inference

### Purpose
Infer and check types for all expressions.

### Input
Resolved AST with generated functions

### Output
Fully typed AST

### Algorithm: Hindley-Milner with Extensions

```
TypeChecker {
    constraints: [Constraint]
    substitution: {TypeVar: Type}
}

Type =
    | Primitive { kind: PrimitiveKind }
    | List { element: Type }
    | Map { key: Type, value: Type }
    | Function { params: [Type], returns: Type }
    | UserDefined { name: string, fields: [(string, Type)] }
    | Option { inner: Type }
    | Union { variants: [Type] }
    | TypeVar { id: int }
    | Generic { name: string, constraints: [Trait] }

Constraint =
    | Equal { left: Type, right: Type }
    | HasField { type: Type, field: string, field_type: Type }
    | Implements { type: Type, trait: Trait }
```

### Inference Steps
1. **Generate constraints** - Walk AST, create type variables, add constraints
2. **Unify** - Solve constraints using unification
3. **Substitute** - Replace type variables with concrete types
4. **Check** - Verify all types are concrete, report errors

## 13.7 Phase 6: HIR Generation

### Purpose
Desugar and normalize the AST into High-level IR.

### Transformations
- Desugar pipes into function calls
- Desugar string interpolation
- Desugar range expressions
- Normalize match expressions
- Expand method calls
- Simplify control flow

### HIR Structure

```
HIR {
    functions: [HIRFunction]
    types: [HIRType]
}

HIRFunction {
    name: string
    params: [(string, Type)]
    return_type: Type
    body: HIRBlock
}

HIRBlock {
    statements: [HIRStatement]
    result: Option<HIRExpr>
}

HIRStatement =
    | Let { name: string, type: Type, value: HIRExpr }
    | Assign { target: HIRPlace, value: HIRExpr }
    | If { cond: HIRExpr, then: HIRBlock, else: Option<HIRBlock> }
    | Loop { body: HIRBlock }
    | Break
    | Continue
    | Return { value: Option<HIRExpr> }
    | Call { func: string, args: [HIRExpr] }

HIRExpr =
    | Literal { value, type: Type }
    | Var { name: string, type: Type }
    | Call { func: string, args: [HIRExpr], type: Type }
    | Binary { op, left: HIRExpr, right: HIRExpr, type: Type }
    | Unary { op, operand: HIRExpr, type: Type }
    | Field { object: HIRExpr, field: string, type: Type }
    | Index { object: HIRExpr, index: HIRExpr, type: Type }
    | Lambda { params, body: HIRBlock, type: Type }
    | Construct { type_name: string, fields: [(string, HIRExpr)] }
```

## 13.8 Phase 7: MIR Generation

### Purpose
Convert HIR to control flow graph in SSA form.

### MIR Structure

```
MIR {
    functions: [MIRFunction]
}

MIRFunction {
    name: string
    params: [(string, Type)]
    return_type: Type
    blocks: [BasicBlock]
    entry: BlockId
}

BasicBlock {
    id: BlockId
    statements: [MIRStatement]
    terminator: Terminator
}

MIRStatement =
    | Assign { place: Place, rvalue: RValue }
    | StorageLive { local: LocalId }
    | StorageDead { local: LocalId }

Terminator =
    | Goto { target: BlockId }
    | If { cond: Operand, then: BlockId, else: BlockId }
    | Switch { value: Operand, targets: [(Value, BlockId)], default: BlockId }
    | Call { func: Operand, args: [Operand], dest: Place, next: BlockId }
    | Return { value: Option<Operand> }
    | Unreachable

Place = LocalId | Place.Field | Place.Index
Operand = Copy(Place) | Move(Place) | Constant(Value)
RValue = Use(Operand) | BinaryOp(op, Operand, Operand) | UnaryOp(op, Operand) | ...
```

## 13.9 Phase 8: Optimization

### Optimization Passes

```
// Analysis passes (gather information)
- Liveness analysis
- Escape analysis
- Alias analysis
- Dominator tree

// Transformation passes
- Constant folding
- Constant propagation
- Dead code elimination
- Common subexpression elimination
- Inlining
- Loop unrolling
- Tail call optimization
- Closure optimization
```

### Pass Manager

```
PassManager {
    passes: [Pass]

    run(mir: MIR) -> MIR {
        for pass in passes {
            mir = pass.run(mir)
        }
        mir
    }
}
```

## 13.10 Phase 9: LLVM IR Generation

### Purpose
Convert optimized MIR to LLVM IR.

### Key Mappings

| Haira Concept | LLVM Representation |
|---------------|---------------------|
| int | i64 |
| float | double |
| bool | i1 |
| string | struct { i8*, i64 } |
| List<T> | struct { T*, i64, i64 } |
| Map<K,V> | opaque pointer (runtime) |
| Option<T> | struct { i1, T } |
| User types | struct |
| Functions | functions |
| Closures | struct + function pointer |

### Runtime Library

The compiler links against a runtime library providing:
- Garbage collector
- String operations
- Collection operations
- Channel operations
- HTTP client/server
- File I/O
- JSON parsing

## 13.11 Phase 10: LLVM Backend

### Purpose
Use LLVM to generate native code.

### Steps
1. **Optimization** - Run LLVM optimization passes
2. **Code generation** - Generate machine code
3. **Linking** - Link with runtime and system libraries
4. **Output** - Produce executable binary

### Targets
- Native (current platform)
- Cross-compilation (other platforms)
- WebAssembly (future)

## 13.12 Compiler Driver

### CLI Interface

```bash
haira build              # Build project
haira run                # Build and run
haira check              # Type check only
haira inspect <func>     # Show generated code
haira fmt                # Format source
haira test               # Run tests
```

### Driver Implementation

```
Driver {
    config: Config

    compile(files: [string]) -> Result<Binary> {
        // Phase 1-2: Parse all files
        asts = files | parallel(f => {
            tokens = Lexer.tokenize(read_file(f))
            Parser.parse(tokens)
        })

        // Phase 3: Name resolution
        resolved = NameResolver.resolve(asts)

        // Phase 4: Auto-generation
        with_generated = AutoGenerator.generate(resolved)

        // Phase 5: Type inference
        typed = TypeChecker.check(with_generated)

        // Phase 6: HIR
        hir = HIRGenerator.generate(typed)

        // Phase 7: MIR
        mir = MIRGenerator.generate(hir)

        // Phase 8: Optimize
        optimized = Optimizer.optimize(mir)

        // Phase 9: LLVM IR
        llvm_ir = LLVMGenerator.generate(optimized)

        // Phase 10: Native code
        LLVMBackend.compile(llvm_ir)
    }
}
```

## 13.13 Error Handling

### Error Types

```
CompilerError {
    kind: ErrorKind
    message: string
    location: SourceLocation
    hints: [string]
    related: [RelatedInfo]
}

ErrorKind =
    | LexError
    | ParseError
    | NameError
    | TypeError
    | AutoGenError
    | CodeGenError
```

### Error Reporting

```
error[E0001]: undefined variable 'usre'
  --> src/main.haira:15:5
   |
15 |     print(usre.name)
   |           ^^^^ not found in this scope
   |
   = hint: did you mean 'user'?
```

## 13.14 Incremental Compilation

### Strategy
- Cache results at phase boundaries
- Use content hashing to detect changes
- Recompile only affected modules

### Cache Structure

```
.haira-cache/
├── lexer/
│   ├── main.haira.tokens
│   └── user.haira.tokens
├── parser/
│   ├── main.haira.ast
│   └── user.haira.ast
├── types/
│   └── project.types
├── mir/
│   ├── main.mir
│   └── user.mir
└── objects/
    ├── main.o
    └── user.o
```

## 13.15 Testing Strategy

### Unit Tests
- Lexer: token stream correctness
- Parser: AST structure
- Type checker: inference and errors
- Code gen: output correctness

### Integration Tests
- End-to-end compilation
- Example programs
- Error message quality

### Fuzzing
- Random program generation
- Grammar-based fuzzing
- Crash detection
