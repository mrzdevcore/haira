//! HIF Writer - Serializes HIF structures to human-readable format.

use super::types::*;
use std::fmt::Write;

/// Writer for HIF format.
pub struct HIFWriter {
    /// Indentation string (default: 2 spaces).
    indent: String,
}

impl Default for HIFWriter {
    fn default() -> Self {
        Self::new()
    }
}

impl HIFWriter {
    /// Create a new HIF writer.
    pub fn new() -> Self {
        Self {
            indent: "  ".to_string(),
        }
    }

    /// Create a writer with custom indentation.
    pub fn with_indent(indent: impl Into<String>) -> Self {
        Self {
            indent: indent.into(),
        }
    }

    /// Write a complete HIF file to string.
    pub fn write(&self, file: &HIFFile) -> String {
        let mut output = String::new();

        // Header
        writeln!(output, "# Haira Intent Format v{}", file.version).unwrap();
        writeln!(output).unwrap();

        // Write structs
        for s in file.structs.values() {
            self.write_struct(&mut output, s, 0);
            writeln!(output).unwrap();
        }

        // Write intents
        for intent in file.intents.values() {
            self.write_intent(&mut output, intent, 0);
            writeln!(output).unwrap();
        }

        output.trim_end().to_string() + "\n"
    }

    /// Write a struct definition.
    fn write_struct(&self, output: &mut String, s: &HIFStruct, depth: usize) {
        let base_indent = self.indent.repeat(depth);
        let field_indent = self.indent.repeat(depth + 1);

        writeln!(output, "{}struct {} @{}", base_indent, s.name, s.hash).unwrap();

        for field in &s.fields {
            writeln!(
                output,
                "{}{}: {}",
                field_indent,
                field.name,
                field.ty.to_hif_string()
            )
            .unwrap();
        }
    }

    /// Write an intent definition.
    fn write_intent(&self, output: &mut String, intent: &HIFIntent, depth: usize) {
        let base_indent = self.indent.repeat(depth);
        let content_indent = self.indent.repeat(depth + 1);

        writeln!(
            output,
            "{}intent {} @{}",
            base_indent, intent.name, intent.hash
        )
        .unwrap();

        // Parameters
        for param in &intent.params {
            writeln!(
                output,
                "{}param {}: {}",
                content_indent,
                param.name,
                param.ty.to_hif_string()
            )
            .unwrap();
        }

        // Return type
        writeln!(
            output,
            "{}returns {}",
            content_indent,
            intent.returns.to_hif_string()
        )
        .unwrap();

        // Body
        if !intent.body.is_empty() {
            writeln!(output, "{}body", content_indent).unwrap();
            for op in &intent.body {
                self.write_operation(output, op, depth + 2);
            }
        }
    }

    /// Write an operation.
    fn write_operation(&self, output: &mut String, op: &HIFOperation, depth: usize) {
        let indent = self.indent.repeat(depth);

        let result_suffix = match (&op.result, &op.result_type) {
            (Some(name), Some(ty)) => format!(" -> {}: {}", name, ty.to_hif_string()),
            (Some(name), None) => format!(" -> {}", name),
            _ => String::new(),
        };

        match &op.kind {
            // Literals
            HIFOpKind::Literal(value) => {
                writeln!(
                    output,
                    "{}literal {}{}",
                    indent,
                    value.to_hif_string(),
                    result_suffix
                )
                .unwrap();
            }

            // Variable reference
            HIFOpKind::Var(name) => {
                writeln!(output, "{}var {}{}", indent, name, result_suffix).unwrap();
            }

            // Return
            HIFOpKind::Return(value) => {
                writeln!(output, "{}return {}", indent, value).unwrap();
            }

            // Binary operations
            HIFOpKind::Add(left, right) => {
                writeln!(output, "{}add {} {}{}", indent, left, right, result_suffix).unwrap();
            }
            HIFOpKind::Sub(left, right) => {
                writeln!(output, "{}sub {} {}{}", indent, left, right, result_suffix).unwrap();
            }
            HIFOpKind::Mul(left, right) => {
                writeln!(output, "{}mul {} {}{}", indent, left, right, result_suffix).unwrap();
            }
            HIFOpKind::Div(left, right) => {
                writeln!(output, "{}div {} {}{}", indent, left, right, result_suffix).unwrap();
            }
            HIFOpKind::Mod(left, right) => {
                writeln!(output, "{}mod {} {}{}", indent, left, right, result_suffix).unwrap();
            }
            HIFOpKind::Eq(left, right) => {
                writeln!(output, "{}eq {} {}{}", indent, left, right, result_suffix).unwrap();
            }
            HIFOpKind::Ne(left, right) => {
                writeln!(output, "{}ne {} {}{}", indent, left, right, result_suffix).unwrap();
            }
            HIFOpKind::Lt(left, right) => {
                writeln!(output, "{}lt {} {}{}", indent, left, right, result_suffix).unwrap();
            }
            HIFOpKind::Gt(left, right) => {
                writeln!(output, "{}gt {} {}{}", indent, left, right, result_suffix).unwrap();
            }
            HIFOpKind::Le(left, right) => {
                writeln!(output, "{}le {} {}{}", indent, left, right, result_suffix).unwrap();
            }
            HIFOpKind::Ge(left, right) => {
                writeln!(output, "{}ge {} {}{}", indent, left, right, result_suffix).unwrap();
            }
            HIFOpKind::And(left, right) => {
                writeln!(output, "{}and {} {}{}", indent, left, right, result_suffix).unwrap();
            }
            HIFOpKind::Or(left, right) => {
                writeln!(output, "{}or {} {}{}", indent, left, right, result_suffix).unwrap();
            }

            // Unary operations
            HIFOpKind::Neg(operand) => {
                writeln!(output, "{}neg {}{}", indent, operand, result_suffix).unwrap();
            }
            HIFOpKind::Not(operand) => {
                writeln!(output, "{}not {}{}", indent, operand, result_suffix).unwrap();
            }

            // Field access
            HIFOpKind::GetField(source, field) => {
                writeln!(
                    output,
                    "{}get_field {}.{}{}",
                    indent, source, field, result_suffix
                )
                .unwrap();
            }
            HIFOpKind::SetField(target, field, value) => {
                writeln!(
                    output,
                    "{}set_field {}.{} = {}",
                    indent, target, field, value
                )
                .unwrap();
            }

            // Index access
            HIFOpKind::GetIndex(source, index) => {
                writeln!(
                    output,
                    "{}get_index {}[{}]{}",
                    indent, source, index, result_suffix
                )
                .unwrap();
            }

            // Collection operations
            HIFOpKind::Map {
                source,
                element_var,
                body,
            } => {
                writeln!(
                    output,
                    "{}map {} as {}{}",
                    indent, source, element_var, result_suffix
                )
                .unwrap();
                for inner_op in body {
                    self.write_operation(output, inner_op, depth + 1);
                }
                writeln!(output, "{}end", indent).unwrap();
            }
            HIFOpKind::Filter {
                source,
                element_var,
                body,
            } => {
                writeln!(
                    output,
                    "{}filter {} as {}{}",
                    indent, source, element_var, result_suffix
                )
                .unwrap();
                for inner_op in body {
                    self.write_operation(output, inner_op, depth + 1);
                }
                writeln!(output, "{}end", indent).unwrap();
            }
            HIFOpKind::Reduce {
                source,
                initial,
                accumulator_var,
                element_var,
                body,
            } => {
                writeln!(
                    output,
                    "{}reduce {} from {} as {}, {}{}",
                    indent, source, initial, accumulator_var, element_var, result_suffix
                )
                .unwrap();
                for inner_op in body {
                    self.write_operation(output, inner_op, depth + 1);
                }
                writeln!(output, "{}end", indent).unwrap();
            }
            HIFOpKind::Sum(source) => {
                writeln!(output, "{}sum {}{}", indent, source, result_suffix).unwrap();
            }
            HIFOpKind::Min(source) => {
                writeln!(output, "{}min {}{}", indent, source, result_suffix).unwrap();
            }
            HIFOpKind::Max(source) => {
                writeln!(output, "{}max {}{}", indent, source, result_suffix).unwrap();
            }
            HIFOpKind::Avg(source) => {
                writeln!(output, "{}avg {}{}", indent, source, result_suffix).unwrap();
            }
            HIFOpKind::Count(source) => {
                writeln!(output, "{}count {}{}", indent, source, result_suffix).unwrap();
            }
            HIFOpKind::Take(source, count) => {
                writeln!(
                    output,
                    "{}take {} {}{}",
                    indent, source, count, result_suffix
                )
                .unwrap();
            }
            HIFOpKind::Skip(source, count) => {
                writeln!(
                    output,
                    "{}skip {} {}{}",
                    indent, source, count, result_suffix
                )
                .unwrap();
            }
            HIFOpKind::Find {
                source,
                element_var,
                body,
            } => {
                writeln!(
                    output,
                    "{}find {} as {}{}",
                    indent, source, element_var, result_suffix
                )
                .unwrap();
                for inner_op in body {
                    self.write_operation(output, inner_op, depth + 1);
                }
                writeln!(output, "{}end", indent).unwrap();
            }
            HIFOpKind::Any {
                source,
                element_var,
                body,
            } => {
                writeln!(
                    output,
                    "{}any {} as {}{}",
                    indent, source, element_var, result_suffix
                )
                .unwrap();
                for inner_op in body {
                    self.write_operation(output, inner_op, depth + 1);
                }
                writeln!(output, "{}end", indent).unwrap();
            }
            HIFOpKind::All {
                source,
                element_var,
                body,
            } => {
                writeln!(
                    output,
                    "{}all {} as {}{}",
                    indent, source, element_var, result_suffix
                )
                .unwrap();
                for inner_op in body {
                    self.write_operation(output, inner_op, depth + 1);
                }
                writeln!(output, "{}end", indent).unwrap();
            }

            // Control flow
            HIFOpKind::If {
                condition,
                then_ops,
                else_ops,
            } => {
                writeln!(output, "{}if{}", indent, result_suffix).unwrap();
                for inner_op in condition {
                    self.write_operation(output, inner_op, depth + 1);
                }
                writeln!(output, "{}then", indent).unwrap();
                for inner_op in then_ops {
                    self.write_operation(output, inner_op, depth + 1);
                }
                if !else_ops.is_empty() {
                    writeln!(output, "{}else", indent).unwrap();
                    for inner_op in else_ops {
                        self.write_operation(output, inner_op, depth + 1);
                    }
                }
                writeln!(output, "{}end", indent).unwrap();
            }
            HIFOpKind::Loop {
                source,
                element_var,
                body,
            } => {
                writeln!(
                    output,
                    "{}loop {} as {}{}",
                    indent, source, element_var, result_suffix
                )
                .unwrap();
                for inner_op in body {
                    self.write_operation(output, inner_op, depth + 1);
                }
                writeln!(output, "{}end", indent).unwrap();
            }

            // Construction
            HIFOpKind::Construct { ty, fields } => {
                writeln!(output, "{}construct {}{}", indent, ty, result_suffix).unwrap();
                let field_indent = self.indent.repeat(depth + 1);
                for (name, value) in fields {
                    writeln!(output, "{}{}: {}", field_indent, name, value).unwrap();
                }
                writeln!(output, "{}end", indent).unwrap();
            }
            HIFOpKind::CreateList(elements) => {
                let elements_str = elements.join(", ");
                writeln!(output, "{}list [{}]{}", indent, elements_str, result_suffix).unwrap();
            }

            // Function call
            HIFOpKind::Call { function, args } => {
                let args_str = args.join(", ");
                writeln!(
                    output,
                    "{}call {}({}){}",
                    indent, function, args_str, result_suffix
                )
                .unwrap();
            }

            // String operations
            HIFOpKind::Concat(parts) => {
                let parts_str = parts.join(", ");
                writeln!(output, "{}concat [{}]{}", indent, parts_str, result_suffix).unwrap();
            }
            HIFOpKind::Format { template, values } => {
                writeln!(
                    output,
                    "{}format \"{}\"{}",
                    indent,
                    template.replace('\"', "\\\""),
                    result_suffix
                )
                .unwrap();
                let field_indent = self.indent.repeat(depth + 1);
                for (key, value) in values {
                    writeln!(output, "{}{}: {}", field_indent, key, value).unwrap();
                }
                writeln!(output, "{}end", indent).unwrap();
            }
        }
    }
}

/// Write a HIF file to string using default settings.
pub fn write_hif(file: &HIFFile) -> String {
    HIFWriter::new().write(file)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_write_struct() {
        let mut file = HIFFile::new();
        file.add_struct(HIFStruct {
            name: "User".to_string(),
            hash: "abc123".to_string(),
            fields: vec![
                HIFField {
                    name: "name".to_string(),
                    ty: HIFType::String,
                },
                HIFField {
                    name: "age".to_string(),
                    ty: HIFType::Int,
                },
                HIFField {
                    name: "email".to_string(),
                    ty: HIFType::Optional(Box::new(HIFType::String)),
                },
            ],
        });

        let output = write_hif(&file);
        assert!(output.contains("struct User @abc123"));
        assert!(output.contains("name: string"));
        assert!(output.contains("age: int"));
        assert!(output.contains("email: string?"));
    }

    #[test]
    fn test_write_intent() {
        let mut file = HIFFile::new();
        file.add_intent(HIFIntent {
            name: "get_user_name".to_string(),
            hash: "def456".to_string(),
            params: vec![HIFParam {
                name: "user".to_string(),
                ty: HIFType::Struct("User".to_string()),
            }],
            returns: HIFType::String,
            body: vec![
                HIFOperation {
                    kind: HIFOpKind::GetField("user".to_string(), "name".to_string()),
                    result: Some("_name".to_string()),
                    result_type: Some(HIFType::String),
                },
                HIFOperation {
                    kind: HIFOpKind::Return("_name".to_string()),
                    result: None,
                    result_type: None,
                },
            ],
        });

        let output = write_hif(&file);
        assert!(output.contains("intent get_user_name @def456"));
        assert!(output.contains("param user: User"));
        assert!(output.contains("returns string"));
        assert!(output.contains("get_field user.name -> _name: string"));
        assert!(output.contains("return _name"));
    }

    #[test]
    fn test_write_complex_types() {
        let mut file = HIFFile::new();
        file.add_struct(HIFStruct {
            name: "Config".to_string(),
            hash: "xyz789".to_string(),
            fields: vec![
                HIFField {
                    name: "items".to_string(),
                    ty: HIFType::Array(Box::new(HIFType::String)),
                },
                HIFField {
                    name: "settings".to_string(),
                    ty: HIFType::Map(Box::new(HIFType::String), Box::new(HIFType::Int)),
                },
            ],
        });

        let output = write_hif(&file);
        assert!(output.contains("items: [string]"));
        assert!(output.contains("settings: {string: int}"));
    }
}
