/// <reference types="tree-sitter-cli/dsl" />
// Tree-sitter grammar for the Haira programming language.
// Designed for syntax highlighting — broad coverage, not full semantic precision.

module.exports = grammar({
  name: "haira",

  extras: ($) => [/\s/, $.line_comment, $.block_comment],

  word: ($) => $.identifier,

  conflicts: ($) => [
    [$.primary_expression, $.map_or_struct],
    [$.method_definition, $.primary_expression],
  ],

  rules: {
    source_file: ($) => repeat($._item),

    _item: ($) =>
      choice(
        $.import_statement,
        $.export_statement,
        $.function_definition,
        $.type_definition,
        $.type_alias,
        $.enum_definition,
        $.method_definition,
        $.provider_declaration,
        $.tool_declaration,
        $.agent_declaration,
        $.workflow_declaration,
        $._statement,
      ),

    // --- Import ---
    // import "io"
    // import fmt from "io"
    // import { User, Post } from "models"
    // import * from "math"
    import_statement: ($) =>
      choice(
        seq("import", $.string),
        seq("import", field("alias", $.identifier), "from", $.string),
        seq("import", "{", commaSep1($.identifier), "}", "from", $.string),
        seq("import", "*", "from", $.string),
      ),

    // --- Export ---
    export_statement: ($) => seq("export", "{", commaSep1($.identifier), "}"),

    // --- Functions ---
    function_definition: ($) =>
      seq(
        optional($.visibility),
        "fn",
        field("name", $.identifier),
        $.parameter_list,
        optional(seq("->", field("return_type", $._type))),
        $.block,
      ),

    visibility: (_$) => "pub",

    parameter_list: ($) => seq("(", optional(commaSep1($.parameter)), ")"),

    parameter: ($) =>
      seq(
        field("name", $.identifier),
        optional(seq(":", field("type", $._type))),
        optional(seq("=", field("default", $._expression))),
      ),

    // --- Types ---
    type_definition: ($) =>
      seq(
        optional($.visibility),
        "struct",
        field("name", $.identifier),
        "{",
        repeat($.field_definition),
        "}",
      ),

    type_alias: ($) =>
      seq(
        optional($.visibility),
        "type",
        field("name", $.identifier),
        "=",
        field("type", $._type),
      ),

    field_definition: ($) =>
      seq(field("name", $.identifier), ":", field("type", $._type)),

    // --- Methods ---
    method_definition: ($) =>
      seq(
        field("type", $.identifier),
        ".",
        field("name", $.identifier),
        $.parameter_list,
        optional(seq("->", field("return_type", $._type))),
        $.block,
      ),

    // --- Enums ---
    enum_definition: ($) =>
      seq(
        optional($.visibility),
        "enum",
        field("name", $.identifier),
        "{",
        repeat($.enum_variant),
        "}",
      ),

    enum_variant: ($) =>
      seq(
        field("name", $.identifier),
        optional(seq("(", commaSep1($._type), ")")),
      ),

    // --- Provider ---
    provider_declaration: ($) =>
      seq(
        "provider",
        field("name", $.identifier),
        "{",
        repeat($.key_value),
        "}",
      ),

    // --- Tool ---
    tool_declaration: ($) =>
      seq(
        "tool",
        field("name", $.identifier),
        $.parameter_list,
        optional(seq("->", field("return_type", $._type))),
        $.tool_body,
      ),

    tool_body: ($) =>
      seq("{", $.triple_quote_string, repeat($._statement), "}"),

    // --- Agent ---
    agent_declaration: ($) =>
      seq("agent", field("name", $.identifier), "{", repeat($.key_value), "}"),

    // --- Workflow ---
    workflow_declaration: ($) =>
      seq(
        repeat($.decorator),
        "workflow",
        field("name", $.identifier),
        $.parameter_list,
        optional(seq("->", field("return_type", $._type))),
        $.block,
      ),

    decorator: ($) =>
      seq(
        "@",
        $.identifier,
        optional(seq("(", optional(commaSep1($._expression)), ")")),
      ),

    // --- Key-value pair for provider/agent blocks ---
    key_value: ($) =>
      seq(field("key", $.identifier), ":", field("value", $._expression)),

    // --- Statements ---
    _statement: ($) =>
      choice(
        $.return_statement,
        $.if_statement,
        $.for_statement,
        $.while_statement,
        $.match_statement,
        $.try_statement,
        $.spawn_block,
        $.select_block,
        $.defer_statement,
        $.errdefer_statement,
        $.step_statement,
        $.lifecycle_hook,
        $.break_statement,
        $.continue_statement,
        $.assignment_or_call,
      ),

    // Unified statement: expression, assignment, or multi-return assignment
    assignment_or_call: ($) =>
      seq(
        $._expression,
        optional(seq(choice("=", "+=", "-=", "*=", "/=", "%="), $._expression)),
      ),

    return_statement: ($) =>
      prec.right(1, seq("return", optional($._expression))),

    if_statement: ($) =>
      prec.right(
        seq(
          "if",
          field("condition", $._expression),
          $.block,
          optional(seq("else", choice($.if_statement, $.block))),
        ),
      ),

    for_statement: ($) =>
      seq(
        "for",
        field("variable", $.identifier),
        "in",
        field("iterable", $._expression),
        $.block,
      ),

    while_statement: ($) =>
      seq("while", field("condition", $._expression), $.block),

    match_statement: ($) =>
      seq(
        "match",
        field("value", $._expression),
        "{",
        repeat($.match_arm),
        "}",
      ),

    match_arm: ($) =>
      seq(
        field("pattern", $._expression),
        "=>",
        choice($.block, $._expression),
      ),

    try_statement: ($) =>
      seq("try", $.block, "catch", field("error", $.identifier), $.block),

    spawn_block: ($) => seq("spawn", $.block),

    select_block: ($) => seq("select", "{", repeat($.select_arm), "}"),

    select_arm: ($) => seq($._expression, "=>", choice($.block, $._expression)),

    defer_statement: ($) =>
      prec.right(1, seq("defer", choice($.block, $._expression))),

    errdefer_statement: ($) =>
      prec.right(1, seq("errdefer", choice($.block, $._expression))),

    step_statement: ($) =>
      seq(
        repeat($.decorator),
        "step",
        field("name", $.string),
        "{",
        repeat(choice($.lifecycle_hook, $._statement)),
        "}",
      ),

    lifecycle_hook: ($) =>
      choice(
        seq("onerror", optional(field("error", $.identifier)), $.block),
        seq("onsuccess", optional(field("result", $.identifier)), $.block),
        seq("oncancel", $.block),
      ),

    break_statement: (_$) => "break",

    continue_statement: (_$) => "continue",

    block: ($) => prec(1, seq("{", repeat($._statement), "}")),

    // --- Expressions ---
    _expression: ($) =>
      choice(
        $.binary_expression,
        $.unary_expression,
        $.pipe_expression,
        $.range_expression,
        $.call_expression,
        $.method_call,
        $.member_access,
        $.index_expression,
        $.primary_expression,
      ),

    primary_expression: ($) =>
      choice(
        $.identifier,
        $.integer,
        $.float,
        $.string,
        $.interpolated_string,
        $.triple_quote_string,
        $.boolean,
        $.none_literal,
        $.list_expression,
        $.map_or_struct,
        $.parenthesized_expression,
        $.closure_expression,
      ),

    binary_expression: ($) =>
      choice(
        ...[
          ["+", 5],
          ["-", 5],
          ["*", 6],
          ["/", 6],
          ["%", 6],
          ["==", 3],
          ["!=", 3],
          ["<", 3],
          [">", 3],
          ["<=", 3],
          [">=", 3],
          ["and", 2],
          ["or", 1],
          ["orelse", 0],
        ].map(([op, p]) =>
          prec.left(
            p,
            seq(
              field("left", $._expression),
              field("operator", op),
              field("right", $._expression),
            ),
          ),
        ),
      ),

    unary_expression: ($) =>
      prec(
        7,
        seq(
          field("operator", choice("-", "not")),
          field("operand", $._expression),
        ),
      ),

    pipe_expression: ($) =>
      prec.right(
        5,
        seq(field("left", $._expression), "|>", field("right", $._expression)),
      ),

    range_expression: ($) =>
      prec.left(
        4,
        seq(
          field("start", $._expression),
          choice("..", "..="),
          field("end", $._expression),
        ),
      ),

    call_expression: ($) =>
      prec.left(8, seq(field("function", $._expression), $.argument_list)),

    method_call: ($) =>
      prec.left(
        10,
        seq(
          field("object", $._expression),
          ".",
          field("method", $.identifier),
          $.argument_list,
        ),
      ),

    member_access: ($) =>
      prec.left(
        9,
        seq(field("object", $._expression), ".", field("member", $.identifier)),
      ),

    index_expression: ($) =>
      prec.left(
        8,
        seq(
          field("object", $._expression),
          "[",
          field("index", $._expression),
          "]",
        ),
      ),

    parenthesized_expression: ($) => seq("(", $._expression, ")"),

    list_expression: ($) => seq("[", optional(commaSep1($._expression)), "]"),

    map_or_struct: ($) =>
      seq(
        optional(field("name", $.identifier)),
        "{",
        optional(
          commaSep1(
            choice(
              seq(
                field("key", $._expression),
                ":",
                field("value", $._expression),
              ),
              seq(
                field("field", $.identifier),
                "=",
                field("value", $._expression),
              ),
            ),
          ),
        ),
        "}",
      ),

    argument_list: ($) => seq("(", optional(commaSep1($.argument)), ")"),

    argument: ($) =>
      choice(
        seq(field("name", $.identifier), ":", field("value", $._expression)),
        $._expression,
      ),

    closure_expression: ($) =>
      seq("fn", $.parameter_list, optional(seq("->", $._type)), $.block),

    // --- Types (simple — no struct_type/function_type to avoid conflicts) ---
    _type: ($) =>
      choice($.type_identifier, $.list_type, $.map_type, $.option_type),

    type_identifier: ($) => $.identifier,

    list_type: ($) => seq("[", $._type, "]"),

    map_type: ($) => seq("{", $._type, ":", $._type, "}"),

    option_type: ($) => prec.left(seq($._type, "?")),

    // --- Literals ---
    identifier: (_$) => /[a-zA-Z_][a-zA-Z0-9_]*/,

    integer: (_$) => /\d+/,

    float: (_$) => /\d+\.\d+/,

    string: (_$) => token(seq('"', repeat(choice(/[^"\\]/, /\\./)), '"')),

    interpolated_string: (_$) =>
      token(seq('"', repeat(choice(/[^"\\$]/, /\\./, /\$\{[^}]*\}/)), '"')),

    triple_quote_string: (_$) =>
      token(seq('"""', repeat(choice(/[^"]/, /"[^"]/, /""[^"]/)), '"""')),

    boolean: (_$) => choice("true", "false"),

    none_literal: (_$) => "none",

    // --- Comments ---
    line_comment: (_$) => token(seq("//", /[^\n]*/)),

    block_comment: (_$) =>
      token(seq("/*", repeat(choice(/[^*]/, /\*[^/]/)), "*/")),
  },
});

function commaSep1(rule) {
  return seq(rule, repeat(seq(",", rule)), optional(","));
}
