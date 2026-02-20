package ast

// SourceFile represents a complete Haira source file.
type SourceFile struct {
	Items []Item
	Span  Span
}

// Item is a top-level declaration or statement.
type Item = Spanned[ItemKind]

// ItemKind is a sum type for top-level items.
type ItemKind interface {
	itemKind()
}

// --- Item kinds ---

type TypeDef struct {
	IsPublic bool
	Name     Spanned[string]
	Fields   []Field
}

type Field struct {
	Name    Spanned[string]
	Ty      *Spanned[Type]
	Default *Expr
	Span    Span
}

type EnumDef struct {
	IsPublic bool
	Name     Spanned[string]
	Variants []EnumVariant
}

type EnumVariant struct {
	Name   Spanned[string]
	Fields []Param
	Span   Span
}

type TypeAlias struct {
	Name Spanned[string]
	Ty   Spanned[Type]
}

type ImportDecl struct {
	Path   string            // "io", "models/user", "tools/github"
	Alias  *Spanned[string]  // import fmt from "io" → Alias = "fmt"
	Names  []Spanned[string] // import { User, Post } from "models"
	IsGlob bool              // import * from "math"
}

type ExportDecl struct {
	Names []Spanned[string] // export { User, Post }
}

type ProviderDecl struct {
	Name   Spanned[string]
	Fields []ProviderField
}

type ProviderField struct {
	Key   Spanned[string]
	Value Expr
}

type ToolDecl struct {
	Decorators  []Decorator // future: @cache, @timeout, etc.
	Name        Spanned[string]
	Params      []Param
	ReturnTy    *Spanned[Type]
	Description string
	Body        *Block
}

type AgentDecl struct {
	Name   Spanned[string]
	Fields []AgentField
}

type AgentField struct {
	Key   Spanned[string]
	Value Expr
}

type WorkflowDecl struct {
	Name        Spanned[string]
	Trigger     *Decorator
	Decorators  []Decorator // @webui, etc.
	Params      []Param
	ReturnTy    *Spanned[Type]
	Description string // optional triple-quoted description (for MCP)
	Body        Block
	Hooks       []LifecycleHook
}

type Decorator struct {
	Name Spanned[string]
	Args []Expr
}

type FunctionDef struct {
	IsPublic bool
	Name     Spanned[string]
	Params   []Param
	ReturnTy *Spanned[Type]
	Body     Block
}

type MethodDef struct {
	TypeName Spanned[string]
	Name     Spanned[string]
	Params   []Param
	ReturnTy *Spanned[Type]
	Body     Block
}

type ItemStatement struct {
	Stmt Statement
}

type TestDecl struct {
	Name Spanned[string]
	Body Block
}

func (TypeDef) itemKind()       {}
func (EnumDef) itemKind()       {}
func (TypeAlias) itemKind()     {}
func (ImportDecl) itemKind()    {}
func (ExportDecl) itemKind()    {}
func (ProviderDecl) itemKind()  {}
func (ToolDecl) itemKind()      {}
func (AgentDecl) itemKind()     {}
func (WorkflowDecl) itemKind()  {}
func (FunctionDef) itemKind()   {}
func (MethodDef) itemKind()     {}
func (ItemStatement) itemKind() {}
func (TestDecl) itemKind()      {}

// --- Type expressions ---

type Type interface {
	typeKind()
}

type NamedType struct {
	Name string
}

type ListType struct {
	Elem Spanned[Type]
}

type MapType struct {
	Key   Spanned[Type]
	Value Spanned[Type]
}

type OptionType struct {
	Inner Spanned[Type]
}

type FunctionType struct {
	Params []Spanned[Type]
	Ret    Spanned[Type]
}

type UnionType struct {
	Variants []Spanned[Type]
}

type GenericType struct {
	Name string
	Args []Spanned[Type]
}

func (NamedType) typeKind()    {}
func (ListType) typeKind()     {}
func (MapType) typeKind()      {}
func (OptionType) typeKind()   {}
func (FunctionType) typeKind() {}
func (UnionType) typeKind()    {}
func (GenericType) typeKind()  {}

// --- Params ---

type Param struct {
	Name    Spanned[string]
	Ty      *Spanned[Type]
	Default *Expr
	IsRest  bool
	Span    Span
}

// --- Statements ---

type Statement = Spanned[StmtKind]

type StmtKind interface {
	stmtKind()
}

type AssignStmt struct {
	Targets []AssignTarget
	Value   Expr
}

type AssignTarget struct {
	Path AssignPath
	Ty   *Spanned[Type]
}

type AssignPath interface {
	assignPath()
}

type IdentPath struct {
	Name Spanned[string]
}

type FieldPath struct {
	Object AssignPath
	Field  Spanned[string]
}

type IndexPath struct {
	Object AssignPath
	Index  Expr
}

func (IdentPath) assignPath() {}
func (FieldPath) assignPath() {}
func (IndexPath) assignPath() {}

type IfStmt struct {
	Condition  Expr
	ThenBranch Block
	ElseBranch ElseBranch // nil, *ElseBlock, or *ElseIf
}

type ElseBranch interface {
	elseBranch()
}

type ElseBlock struct {
	Body Block
}

type ElseIf struct {
	If Spanned[IfStmt]
}

func (*ElseBlock) elseBranch() {}
func (*ElseIf) elseBranch()    {}

type ForStmt struct {
	Pattern  ForPattern
	Iterator Expr
	Body     Block
}

type ForPattern interface {
	forPattern()
}

type SinglePattern struct {
	Name Spanned[string]
}

type PairPattern struct {
	First  Spanned[string]
	Second Spanned[string]
}

func (SinglePattern) forPattern() {}
func (PairPattern) forPattern()   {}

type WhileStmt struct {
	Condition Expr
	Body      Block
}

type ReturnStmt struct {
	Values []Expr
}

type TryStmt struct {
	Body      Block
	ErrorName Spanned[string]
	CatchBody Block
}

type DeferStmt struct {
	Value Expr
}

type MatchStmt struct {
	Match MatchExpr
}

type BreakStmt struct{}
type ContinueStmt struct{}

type ExprStmt struct {
	Value Expr
}

type StepStmt struct {
	Name       Spanned[string]
	Body       []Statement
	Decorators []Decorator // @retry etc.
	Hooks      []LifecycleHook
}

// LifecycleHook represents onerror/onsuccess/oncancel blocks in workflows and steps.
type LifecycleHook struct {
	Kind    LifecycleHookKind
	ErrName string // only for Onerror: the error variable name
	ArgName string // only for Onsuccess at workflow level: the result variable name
	Body    Block
}

type LifecycleHookKind int

const (
	HookOnerror LifecycleHookKind = iota
	HookOnsuccess
	HookOncancel
)

type ErrDeferStmt struct {
	Value Expr
}

type AssertStmt struct {
	Condition Expr
	Message   *Expr // optional custom message
}

type LetStmt struct {
	Name    Spanned[string]
	TypeAnn *Spanned[Type] // optional type annotation
	Value   Expr
	IsConst bool // true for const declarations
}

func (AssignStmt) stmtKind()   {}
func (LetStmt) stmtKind()      {}
func (IfStmt) stmtKind()       {}
func (ForStmt) stmtKind()      {}
func (WhileStmt) stmtKind()    {}
func (ReturnStmt) stmtKind()   {}
func (TryStmt) stmtKind()      {}
func (DeferStmt) stmtKind()    {}
func (ErrDeferStmt) stmtKind() {}
func (MatchStmt) stmtKind()    {}
func (BreakStmt) stmtKind()    {}
func (ContinueStmt) stmtKind() {}
func (ExprStmt) stmtKind()     {}
func (StepStmt) stmtKind()     {}
func (AssertStmt) stmtKind()   {}

// --- Expressions ---

type Expr = Spanned[ExprKind]

type ExprKind interface {
	exprKind()
}

type LiteralExpr struct {
	Lit Literal
}

type Literal interface {
	literalKind()
}

type IntLit struct{ Value int64 }
type FloatLit struct{ Value float64 }
type StringLit struct{ Value string }
type BoolLit struct{ Value bool }
type InterpolatedStringLit struct{ Parts []StringPart }

func (IntLit) literalKind()                {}
func (FloatLit) literalKind()              {}
func (StringLit) literalKind()             {}
func (BoolLit) literalKind()               {}
func (InterpolatedStringLit) literalKind() {}

type StringPart interface {
	stringPart()
}

type LiteralPart struct{ Value string }
type ExprPart struct{ Value Expr }

func (LiteralPart) stringPart() {}
func (ExprPart) stringPart()    {}

type IdentExpr struct {
	Name string
}

type BinaryExpr struct {
	Left  Expr
	Op    Spanned[BinaryOp]
	Right Expr
}

type BinaryOp int

const (
	OpAdd BinaryOp = iota
	OpSub
	OpMul
	OpDiv
	OpMod
	OpEq
	OpNe
	OpLt
	OpGt
	OpLe
	OpGe
	OpAnd
	OpOr
	OpBitAnd // &
	OpBitOr  // |
	OpBitXor // ^
	OpShl    // <<
	OpShr    // >>
)

type UnaryExpr struct {
	Op      Spanned[UnaryOp]
	Operand Expr
}

type UnaryOp int

const (
	OpNeg UnaryOp = iota
	OpNot
	OpBitNot // ~
)

type CallExpr struct {
	Callee Expr
	Args   []Argument
}

type Argument struct {
	Name  *Spanned[string]
	Value Expr
	Span  Span
}

type MethodCallExpr struct {
	Receiver Expr
	Method   Spanned[string]
	Args     []Argument
}

type FieldExpr struct {
	Object Expr
	Field  Spanned[string]
}

type IndexExpr struct {
	Object Expr
	Index  Expr
}

type PipeExpr struct {
	Left  Expr
	Right Expr
}

type LambdaExpr struct {
	Params []Param
	Body   LambdaBody
}

type LambdaBody interface {
	lambdaBody()
}

type LambdaExprBody struct {
	Value Expr
}

type LambdaBlockBody struct {
	Value Block
}

func (LambdaExprBody) lambdaBody()  {}
func (LambdaBlockBody) lambdaBody() {}

type MatchExpr struct {
	Subject Expr
	Arms    []MatchArm
}

type MatchArm struct {
	Pattern Spanned[Pattern]
	Guard   *Expr
	Body    MatchArmBody
	Span    Span
}

type MatchArmBody interface {
	matchArmBody()
}

type MatchArmExpr struct{ Value Expr }
type MatchArmBlock struct{ Value Block }

func (MatchArmExpr) matchArmBody()  {}
func (MatchArmBlock) matchArmBody() {}

type Pattern interface {
	patternKind()
}

type WildcardPattern struct{}
type LiteralPattern struct{ Lit Literal }
type IdentPattern struct{ Name string }
type ConstructorPattern struct {
	Name   string
	Fields []Spanned[string]
}

type OrPattern struct {
	Patterns []Spanned[Pattern]
}
type RangePattern struct {
	Start     Literal // IntLit
	End       Literal // IntLit
	Inclusive bool    // ..= vs ..
}

func (WildcardPattern) patternKind()    {}
func (LiteralPattern) patternKind()     {}
func (IdentPattern) patternKind()       {}
func (ConstructorPattern) patternKind() {}
func (OrPattern) patternKind()          {}
func (RangePattern) patternKind()       {}

type IfExpr struct {
	If IfStmt
}

type BlockExpr struct {
	Body Block
}

type ListExpr struct {
	Elems []Expr
}

type MapExpr struct {
	Entries []MapEntry
}

type MapEntry struct {
	Key   Expr
	Value Expr
}

type InstanceExpr struct {
	TypeName Spanned[string]
	Fields   []InstanceField
}

type InstanceField struct {
	Name  *Spanned[string]
	Value Expr
	Span  Span
}

type RangeExpr struct {
	Start     Expr
	End       Expr
	Inclusive bool
}

type PropagateExpr struct {
	Inner Expr
}

type OrelseExpr struct {
	Left    Expr
	Default Expr
}

type SomeExpr struct {
	Inner Expr
}

type NoneExpr struct{}

type AsyncExpr struct {
	Body Block
}

type SpawnExpr struct {
	Body Block
}

type SelectExpr struct {
	Arms    []SelectArm
	Default *Block
}

type SelectArm struct {
	Binding Spanned[string]
	Channel Expr
	Body    MatchArmBody
	Span    Span
}

type ParenExpr struct {
	Inner Expr
}

func (LiteralExpr) exprKind()    {}
func (IdentExpr) exprKind()      {}
func (BinaryExpr) exprKind()     {}
func (UnaryExpr) exprKind()      {}
func (CallExpr) exprKind()       {}
func (MethodCallExpr) exprKind() {}
func (FieldExpr) exprKind()      {}
func (IndexExpr) exprKind()      {}
func (PipeExpr) exprKind()       {}
func (LambdaExpr) exprKind()     {}
func (MatchExpr) exprKind()      {}
func (IfExpr) exprKind()         {}
func (BlockExpr) exprKind()      {}
func (ListExpr) exprKind()       {}
func (MapExpr) exprKind()        {}
func (InstanceExpr) exprKind()   {}
func (RangeExpr) exprKind()      {}
func (PropagateExpr) exprKind()  {}
func (OrelseExpr) exprKind()     {}
func (SomeExpr) exprKind()       {}
func (NoneExpr) exprKind()       {}
func (AsyncExpr) exprKind()      {}
func (SpawnExpr) exprKind()      {}
func (SelectExpr) exprKind()     {}
func (ParenExpr) exprKind()      {}

// --- Blocks ---

type Block struct {
	Statements []Statement
	Span       Span
}
