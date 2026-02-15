; Haira tree-sitter highlights

; --- Comments ---
(line_comment) @comment
(block_comment) @comment

; --- Strings ---
(string) @string
(interpolated_string) @string
(triple_quote_string) @string

; --- Numbers ---
(integer) @number
(float) @number

; --- Booleans / None ---
(boolean) @boolean
(none_literal) @constant.builtin

; --- Keywords ---
[
  "if"
  "else"
  "for"
  "while"
  "return"
  "match"
  "in"
  "try"
  "catch"
  "spawn"
  "select"
  "defer"
  "errdefer"
  "step"
  "onerror"
  "onsuccess"
  "oncancel"
  "import"
  "export"
  "from"
] @keyword

(break_statement) @keyword
(continue_statement) @keyword

[
  "and"
  "or"
  "not"
  "orelse"
] @keyword.operator

; --- Agentic keywords ---
[
  "provider"
  "tool"
  "agent"
  "workflow"
] @keyword

; --- Definition keywords ---
[
  "fn"
  "struct"
  "type"
  "enum"
] @keyword

(visibility) @keyword

; --- Operators ---
(binary_expression operator: _ @operator)
(unary_expression operator: _ @operator)

[
  "="
  "+="
  "-="
  "*="
  "/="
  "%="
  "->"
  "=>"
  ".."
  "..="
  "|"
  "|>"
  "&"
  "^"
  "~"
  "<<"
  ">>"
  "&="
  "|="
  "^="
  "<<="
  ">>="
  "?"
] @operator

; --- Punctuation ---
["(" ")" "[" "]" "{" "}"] @punctuation.bracket
["," "." ":"] @punctuation.delimiter

; --- Decorator ---
(decorator "@" @attribute)
(decorator (identifier) @attribute)

; --- Definitions ---
(function_definition name: (identifier) @function)
(method_definition type: (identifier) @type)
(method_definition name: (identifier) @function)
(tool_declaration name: (identifier) @function)
(workflow_declaration name: (identifier) @function)
(type_alias name: (identifier) @type)
(agent_declaration name: (identifier) @type)
(provider_declaration name: (identifier) @type)
(type_definition name: (identifier) @type)
(enum_definition name: (identifier) @type)
(enum_variant name: (identifier) @constant)

; --- Parameters ---
(parameter name: (identifier) @variable.parameter)

; --- Types ---
(type_identifier (identifier) @type)
(field_definition name: (identifier) @property)

; --- Calls ---
(call_expression function: (primary_expression (identifier) @function.call))
(method_call method: (identifier) @function.method)

; --- Members ---
(member_access member: (identifier) @property)
(key_value key: (identifier) @property)

; --- Step ---
(step_statement name: (string) @string)

; --- Identifiers (fallback) ---
(identifier) @variable
