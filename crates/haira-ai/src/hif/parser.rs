//! HIF Parser - Parses human-readable HIF format into structures.

use super::types::*;

/// Error type for HIF parsing.
#[derive(Debug, Clone)]
pub struct HIFParseError {
    pub message: String,
    pub line: usize,
}

impl std::fmt::Display for HIFParseError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "HIF parse error at line {}: {}", self.line, self.message)
    }
}

impl std::error::Error for HIFParseError {}

/// Result type for HIF parsing.
pub type HIFResult<T> = Result<T, HIFParseError>;

/// Parser for HIF format.
pub struct HIFParser {
    lines: Vec<String>,
    current: usize,
}

impl HIFParser {
    /// Create a new parser for the given input.
    pub fn new(input: &str) -> Self {
        Self {
            lines: input.lines().map(|s| s.to_string()).collect(),
            current: 0,
        }
    }

    /// Parse the input into a HIF file.
    pub fn parse(&mut self) -> HIFResult<HIFFile> {
        let mut file = HIFFile::new();

        // Parse header
        self.parse_header(&mut file)?;

        // Parse definitions
        while self.current < self.lines.len() {
            let line = &self.lines[self.current];
            let trimmed = line.trim();

            if trimmed.is_empty() || trimmed.starts_with('#') {
                self.current += 1;
                continue;
            }

            if trimmed.starts_with("struct ") {
                let s = self.parse_struct()?;
                file.add_struct(s);
            } else if trimmed.starts_with("intent ") {
                let intent = self.parse_intent()?;
                file.add_intent(intent);
            } else {
                return Err(self.error(format!("unexpected token: {}", trimmed)));
            }
        }

        Ok(file)
    }

    /// Parse the file header.
    fn parse_header(&mut self, file: &mut HIFFile) -> HIFResult<()> {
        while self.current < self.lines.len() {
            let line = &self.lines[self.current];
            let trimmed = line.trim();

            if trimmed.is_empty() {
                self.current += 1;
                continue;
            }

            if trimmed.starts_with("# Haira Intent Format v") {
                let version_str = trimmed
                    .strip_prefix("# Haira Intent Format v")
                    .unwrap_or("1");
                file.version = version_str.parse().unwrap_or(1);
                self.current += 1;
                return Ok(());
            } else if trimmed.starts_with('#') {
                // Skip other comments
                self.current += 1;
                continue;
            } else {
                // No header found, use default version
                return Ok(());
            }
        }
        Ok(())
    }

    /// Parse a struct definition.
    fn parse_struct(&mut self) -> HIFResult<HIFStruct> {
        let line = self.lines[self.current].clone();
        let trimmed = line.trim();
        let base_indent = self.get_indent(&line);
        self.current += 1;

        // Parse: struct Name @hash
        let rest = trimmed
            .strip_prefix("struct ")
            .ok_or_else(|| self.error("expected 'struct'"))?;

        let (name, hash) = self.parse_name_and_hash(rest)?;

        let mut fields = Vec::new();

        // Parse fields
        while self.current < self.lines.len() {
            let field_line = &self.lines[self.current];
            let field_indent = self.get_indent(field_line);
            let field_trimmed = field_line.trim();

            // Empty lines or comments
            if field_trimmed.is_empty() || field_trimmed.starts_with('#') {
                self.current += 1;
                continue;
            }

            // Check if we're still in the struct (indented)
            if field_indent <= base_indent {
                break;
            }

            self.current += 1;

            // Parse field: name: type
            if let Some((field_name, field_type)) = field_trimmed.split_once(':') {
                fields.push(HIFField {
                    name: field_name.trim().to_string(),
                    ty: HIFType::parse(field_type.trim()),
                });
            } else {
                return Err(self.error(format!("invalid field definition: {}", field_trimmed)));
            }
        }

        Ok(HIFStruct { name, hash, fields })
    }

    /// Parse an intent definition.
    fn parse_intent(&mut self) -> HIFResult<HIFIntent> {
        let line = self.lines[self.current].clone();
        let trimmed = line.trim();
        let base_indent = self.get_indent(&line);
        self.current += 1;

        // Parse: intent name @hash
        let rest = trimmed
            .strip_prefix("intent ")
            .ok_or_else(|| self.error("expected 'intent'"))?;

        let (name, hash) = self.parse_name_and_hash(rest)?;

        let mut params = Vec::new();
        let mut returns = HIFType::Void;
        let mut body = Vec::new();

        // Parse intent content
        while self.current < self.lines.len() {
            let content_line = self.lines[self.current].clone();
            let content_indent = self.get_indent(&content_line);
            let content_trimmed = content_line.trim().to_string();

            // Empty lines or comments
            if content_trimmed.is_empty() || content_trimmed.starts_with('#') {
                self.current += 1;
                continue;
            }

            // Check if we're still in the intent
            if content_indent <= base_indent {
                break;
            }

            if content_trimmed.starts_with("param ") {
                self.current += 1;
                let param = self.parse_param(&content_trimmed)?;
                params.push(param);
            } else if content_trimmed.starts_with("returns ") {
                self.current += 1;
                let type_str = content_trimmed.strip_prefix("returns ").unwrap().trim();
                returns = HIFType::parse(type_str);
            } else if content_trimmed == "body" {
                self.current += 1;
                body = self.parse_body(content_indent)?;
            } else {
                return Err(self.error(format!("unexpected in intent: {}", content_trimmed)));
            }
        }

        Ok(HIFIntent {
            name,
            hash,
            params,
            returns,
            body,
        })
    }

    /// Parse a parameter definition.
    fn parse_param(&mut self, line: &str) -> HIFResult<HIFParam> {
        // param name: type
        let rest = line.strip_prefix("param ").unwrap();
        if let Some((name, ty)) = rest.split_once(':') {
            Ok(HIFParam {
                name: name.trim().to_string(),
                ty: HIFType::parse(ty.trim()),
            })
        } else {
            Err(self.error(format!("invalid param definition: {}", line)))
        }
    }

    /// Parse the body of an intent.
    fn parse_body(&mut self, base_indent: usize) -> HIFResult<Vec<HIFOperation>> {
        let mut operations = Vec::new();

        while self.current < self.lines.len() {
            let line = self.lines[self.current].clone();
            let indent = self.get_indent(&line);
            let trimmed = line.trim().to_string();

            // Empty lines or comments
            if trimmed.is_empty() || trimmed.starts_with('#') {
                self.current += 1;
                continue;
            }

            // Check if we're still in the body
            if indent <= base_indent {
                break;
            }

            self.current += 1;
            let op = self.parse_operation(&trimmed, indent)?;
            operations.push(op);
        }

        Ok(operations)
    }

    /// Parse a single operation.
    fn parse_operation(&mut self, line: &str, current_indent: usize) -> HIFResult<HIFOperation> {
        // Check for result suffix: -> name: type or -> name
        let (op_part, result, result_type) = self.parse_result_suffix(line);

        let kind = self.parse_op_kind(&op_part, current_indent)?;

        Ok(HIFOperation {
            kind,
            result,
            result_type,
        })
    }

    /// Parse the result suffix from an operation line.
    fn parse_result_suffix(&self, line: &str) -> (String, Option<String>, Option<HIFType>) {
        if let Some(arrow_pos) = line.find(" -> ") {
            let op_part = line[..arrow_pos].to_string();
            let result_part = &line[arrow_pos + 4..];

            if let Some((name, ty)) = result_part.split_once(':') {
                (
                    op_part,
                    Some(name.trim().to_string()),
                    Some(HIFType::parse(ty.trim())),
                )
            } else {
                (op_part, Some(result_part.trim().to_string()), None)
            }
        } else {
            (line.to_string(), None, None)
        }
    }

    /// Parse an operation kind.
    fn parse_op_kind(&mut self, line: &str, current_indent: usize) -> HIFResult<HIFOpKind> {
        let parts: Vec<&str> = line.split_whitespace().collect();
        if parts.is_empty() {
            return Err(self.error("empty operation"));
        }

        let op_name = parts[0];

        match op_name {
            // Literals
            "literal" => {
                let value_str = parts[1..].join(" ");
                let value = self.parse_value(&value_str)?;
                Ok(HIFOpKind::Literal(value))
            }

            // Variable reference
            "var" => {
                let name = parts
                    .get(1)
                    .ok_or_else(|| self.error("var requires name"))?;
                Ok(HIFOpKind::Var(name.to_string()))
            }

            // Return
            "return" => {
                let name = parts
                    .get(1)
                    .ok_or_else(|| self.error("return requires value"))?;
                Ok(HIFOpKind::Return(name.to_string()))
            }

            // Binary operations
            "add" => self.parse_binary_op(&parts, HIFOpKind::Add),
            "sub" => self.parse_binary_op(&parts, HIFOpKind::Sub),
            "mul" => self.parse_binary_op(&parts, HIFOpKind::Mul),
            "div" => self.parse_binary_op(&parts, HIFOpKind::Div),
            "mod" => self.parse_binary_op(&parts, HIFOpKind::Mod),
            "eq" => self.parse_binary_op(&parts, HIFOpKind::Eq),
            "ne" => self.parse_binary_op(&parts, HIFOpKind::Ne),
            "lt" => self.parse_binary_op(&parts, HIFOpKind::Lt),
            "gt" => self.parse_binary_op(&parts, HIFOpKind::Gt),
            "le" => self.parse_binary_op(&parts, HIFOpKind::Le),
            "ge" => self.parse_binary_op(&parts, HIFOpKind::Ge),
            "and" => self.parse_binary_op(&parts, HIFOpKind::And),
            "or" => self.parse_binary_op(&parts, HIFOpKind::Or),

            // Unary operations
            "neg" => {
                let operand = parts
                    .get(1)
                    .ok_or_else(|| self.error("neg requires operand"))?;
                Ok(HIFOpKind::Neg(operand.to_string()))
            }
            "not" => {
                let operand = parts
                    .get(1)
                    .ok_or_else(|| self.error("not requires operand"))?;
                Ok(HIFOpKind::Not(operand.to_string()))
            }

            // Field access
            "get_field" => {
                let field_ref = parts
                    .get(1)
                    .ok_or_else(|| self.error("get_field requires source.field"))?;
                if let Some((source, field)) = field_ref.split_once('.') {
                    Ok(HIFOpKind::GetField(source.to_string(), field.to_string()))
                } else {
                    Err(self.error("get_field requires source.field format"))
                }
            }
            "set_field" => {
                // set_field target.field = value
                let field_ref = parts
                    .get(1)
                    .ok_or_else(|| self.error("set_field requires target.field"))?;
                let value = parts
                    .get(3)
                    .ok_or_else(|| self.error("set_field requires value"))?;
                if let Some((target, field)) = field_ref.split_once('.') {
                    Ok(HIFOpKind::SetField(
                        target.to_string(),
                        field.to_string(),
                        value.to_string(),
                    ))
                } else {
                    Err(self.error("set_field requires target.field format"))
                }
            }

            // Index access
            "get_index" => {
                // get_index source[index]
                let expr = parts
                    .get(1)
                    .ok_or_else(|| self.error("get_index requires source[index]"))?;
                if let Some(bracket_pos) = expr.find('[') {
                    let source = &expr[..bracket_pos];
                    let index = &expr[bracket_pos + 1..expr.len() - 1];
                    Ok(HIFOpKind::GetIndex(source.to_string(), index.to_string()))
                } else {
                    Err(self.error("get_index requires source[index] format"))
                }
            }

            // Collection operations with body
            "map" | "filter" | "find" | "any" | "all" | "loop" => {
                self.parse_collection_op(op_name, &parts, current_indent)
            }

            "reduce" => self.parse_reduce_op(&parts, current_indent),

            // Simple aggregations
            "sum" => {
                let source = parts
                    .get(1)
                    .ok_or_else(|| self.error("sum requires source"))?;
                Ok(HIFOpKind::Sum(source.to_string()))
            }
            "min" => {
                let source = parts
                    .get(1)
                    .ok_or_else(|| self.error("min requires source"))?;
                Ok(HIFOpKind::Min(source.to_string()))
            }
            "max" => {
                let source = parts
                    .get(1)
                    .ok_or_else(|| self.error("max requires source"))?;
                Ok(HIFOpKind::Max(source.to_string()))
            }
            "avg" => {
                let source = parts
                    .get(1)
                    .ok_or_else(|| self.error("avg requires source"))?;
                Ok(HIFOpKind::Avg(source.to_string()))
            }
            "count" => {
                let source = parts
                    .get(1)
                    .ok_or_else(|| self.error("count requires source"))?;
                Ok(HIFOpKind::Count(source.to_string()))
            }
            "take" => {
                let source = parts
                    .get(1)
                    .ok_or_else(|| self.error("take requires source"))?;
                let count = parts
                    .get(2)
                    .ok_or_else(|| self.error("take requires count"))?;
                Ok(HIFOpKind::Take(source.to_string(), count.to_string()))
            }
            "skip" => {
                let source = parts
                    .get(1)
                    .ok_or_else(|| self.error("skip requires source"))?;
                let count = parts
                    .get(2)
                    .ok_or_else(|| self.error("skip requires count"))?;
                Ok(HIFOpKind::Skip(source.to_string(), count.to_string()))
            }

            // Control flow
            "if" => self.parse_if_op(current_indent),

            // Construction
            "construct" => self.parse_construct_op(&parts, current_indent),
            "list" => self.parse_list_op(&parts),

            // Function call
            "call" => self.parse_call_op(&parts),

            // String operations
            "concat" => self.parse_concat_op(&parts),
            "format" => self.parse_format_op(&parts, current_indent),

            _ => Err(self.error(format!("unknown operation: {}", op_name))),
        }
    }

    /// Parse a binary operation.
    fn parse_binary_op<F>(&self, parts: &[&str], constructor: F) -> HIFResult<HIFOpKind>
    where
        F: FnOnce(String, String) -> HIFOpKind,
    {
        let left = parts
            .get(1)
            .ok_or_else(|| self.error("binary op requires left operand"))?;
        let right = parts
            .get(2)
            .ok_or_else(|| self.error("binary op requires right operand"))?;
        Ok(constructor(left.to_string(), right.to_string()))
    }

    /// Parse a collection operation with body.
    fn parse_collection_op(
        &mut self,
        op_name: &str,
        parts: &[&str],
        current_indent: usize,
    ) -> HIFResult<HIFOpKind> {
        // op source as element_var
        let source = parts
            .get(1)
            .ok_or_else(|| self.error(format!("{} requires source", op_name)))?;
        let element_var = parts
            .get(3)
            .ok_or_else(|| self.error(format!("{} requires element variable", op_name)))?;

        let body = self.parse_block_body(current_indent)?;

        match op_name {
            "map" => Ok(HIFOpKind::Map {
                source: source.to_string(),
                element_var: element_var.to_string(),
                body,
            }),
            "filter" => Ok(HIFOpKind::Filter {
                source: source.to_string(),
                element_var: element_var.to_string(),
                body,
            }),
            "find" => Ok(HIFOpKind::Find {
                source: source.to_string(),
                element_var: element_var.to_string(),
                body,
            }),
            "any" => Ok(HIFOpKind::Any {
                source: source.to_string(),
                element_var: element_var.to_string(),
                body,
            }),
            "all" => Ok(HIFOpKind::All {
                source: source.to_string(),
                element_var: element_var.to_string(),
                body,
            }),
            "loop" => Ok(HIFOpKind::Loop {
                source: source.to_string(),
                element_var: element_var.to_string(),
                body,
            }),
            _ => Err(self.error(format!("unknown collection op: {}", op_name))),
        }
    }

    /// Parse a reduce operation.
    fn parse_reduce_op(&mut self, parts: &[&str], current_indent: usize) -> HIFResult<HIFOpKind> {
        // reduce source from initial as acc, elem
        let source = parts
            .get(1)
            .ok_or_else(|| self.error("reduce requires source"))?;
        let initial = parts
            .get(3)
            .ok_or_else(|| self.error("reduce requires initial value"))?;
        let vars = parts
            .get(5)
            .ok_or_else(|| self.error("reduce requires variables"))?;

        let (acc_var, elem_var) = if let Some(comma_pos) = vars.find(',') {
            let acc = vars[..comma_pos].trim();
            let elem = parts.get(6).map(|s| s.trim()).unwrap_or("");
            (acc.to_string(), elem.to_string())
        } else {
            return Err(self.error("reduce requires accumulator and element variables"));
        };

        let body = self.parse_block_body(current_indent)?;

        Ok(HIFOpKind::Reduce {
            source: source.to_string(),
            initial: initial.to_string(),
            accumulator_var: acc_var,
            element_var: elem_var,
            body,
        })
    }

    /// Parse an if operation.
    fn parse_if_op(&mut self, current_indent: usize) -> HIFResult<HIFOpKind> {
        let mut condition = Vec::new();
        let mut then_ops = Vec::new();
        let mut else_ops = Vec::new();

        let mut state = "condition"; // condition, then, else

        while self.current < self.lines.len() {
            let line = self.lines[self.current].clone();
            let indent = self.get_indent(&line);
            let trimmed = line.trim().to_string();

            if trimmed.is_empty() || trimmed.starts_with('#') {
                self.current += 1;
                continue;
            }

            if indent <= current_indent {
                if trimmed == "then" {
                    self.current += 1;
                    state = "then";
                    continue;
                } else if trimmed == "else" {
                    self.current += 1;
                    state = "else";
                    continue;
                } else if trimmed == "end" {
                    self.current += 1;
                    break;
                } else {
                    break;
                }
            }

            self.current += 1;
            let op = self.parse_operation(&trimmed, indent)?;

            match state {
                "condition" => condition.push(op),
                "then" => then_ops.push(op),
                "else" => else_ops.push(op),
                _ => {}
            }
        }

        Ok(HIFOpKind::If {
            condition,
            then_ops,
            else_ops,
        })
    }

    /// Parse a construct operation.
    fn parse_construct_op(
        &mut self,
        parts: &[&str],
        current_indent: usize,
    ) -> HIFResult<HIFOpKind> {
        let ty = parts
            .get(1)
            .ok_or_else(|| self.error("construct requires type"))?;
        let fields = self.parse_field_assignments(current_indent)?;

        Ok(HIFOpKind::Construct {
            ty: ty.to_string(),
            fields,
        })
    }

    /// Parse field assignments in a block.
    fn parse_field_assignments(
        &mut self,
        current_indent: usize,
    ) -> HIFResult<Vec<(String, String)>> {
        let mut fields = Vec::new();

        while self.current < self.lines.len() {
            let line = self.lines[self.current].clone();
            let indent = self.get_indent(&line);
            let trimmed = line.trim().to_string();

            if trimmed.is_empty() || trimmed.starts_with('#') {
                self.current += 1;
                continue;
            }

            if trimmed == "end" {
                self.current += 1;
                break;
            }

            if indent <= current_indent {
                break;
            }

            self.current += 1;

            if let Some((name, value)) = trimmed.split_once(':') {
                fields.push((name.trim().to_string(), value.trim().to_string()));
            }
        }

        Ok(fields)
    }

    /// Parse a list operation.
    fn parse_list_op(&self, parts: &[&str]) -> HIFResult<HIFOpKind> {
        // list [elem1, elem2, ...]
        let list_str = parts[1..].join(" ");
        let inner = list_str.trim_start_matches('[').trim_end_matches(']');
        let elements: Vec<String> = inner
            .split(',')
            .map(|s| s.trim().to_string())
            .filter(|s| !s.is_empty())
            .collect();
        Ok(HIFOpKind::CreateList(elements))
    }

    /// Parse a call operation.
    fn parse_call_op(&self, parts: &[&str]) -> HIFResult<HIFOpKind> {
        // call func(arg1, arg2)
        let call_str = parts[1..].join(" ");
        if let Some(paren_pos) = call_str.find('(') {
            let function = call_str[..paren_pos].to_string();
            let args_str = &call_str[paren_pos + 1..call_str.len() - 1];
            let args: Vec<String> = args_str
                .split(',')
                .map(|s| s.trim().to_string())
                .filter(|s| !s.is_empty())
                .collect();
            Ok(HIFOpKind::Call { function, args })
        } else {
            Err(self.error("call requires function(args) format"))
        }
    }

    /// Parse a concat operation.
    fn parse_concat_op(&self, parts: &[&str]) -> HIFResult<HIFOpKind> {
        // concat [part1, part2, ...]
        let concat_str = parts[1..].join(" ");
        let inner = concat_str.trim_start_matches('[').trim_end_matches(']');
        let parts: Vec<String> = inner
            .split(',')
            .map(|s| s.trim().to_string())
            .filter(|s| !s.is_empty())
            .collect();
        Ok(HIFOpKind::Concat(parts))
    }

    /// Parse a format operation.
    fn parse_format_op(&mut self, parts: &[&str], current_indent: usize) -> HIFResult<HIFOpKind> {
        // format "template"
        let format_str = parts[1..].join(" ");
        let template = format_str
            .trim_matches('"')
            .replace("\\\"", "\"")
            .to_string();
        let values = self.parse_field_assignments(current_indent)?;

        Ok(HIFOpKind::Format { template, values })
    }

    /// Parse a block body until "end".
    fn parse_block_body(&mut self, current_indent: usize) -> HIFResult<Vec<HIFOperation>> {
        let mut operations = Vec::new();

        while self.current < self.lines.len() {
            let line = self.lines[self.current].clone();
            let indent = self.get_indent(&line);
            let trimmed = line.trim().to_string();

            if trimmed.is_empty() || trimmed.starts_with('#') {
                self.current += 1;
                continue;
            }

            if trimmed == "end" {
                self.current += 1;
                break;
            }

            if indent <= current_indent {
                break;
            }

            self.current += 1;
            let op = self.parse_operation(&trimmed, indent)?;
            operations.push(op);
        }

        Ok(operations)
    }

    /// Parse name and hash from "Name @hash" format.
    fn parse_name_and_hash(&self, rest: &str) -> HIFResult<(String, String)> {
        if let Some(at_pos) = rest.find('@') {
            let name = rest[..at_pos].trim().to_string();
            let hash = rest[at_pos + 1..].trim().to_string();
            Ok((name, hash))
        } else {
            // No hash, generate empty
            Ok((rest.trim().to_string(), String::new()))
        }
    }

    /// Parse a literal value.
    fn parse_value(&self, s: &str) -> HIFResult<HIFValue> {
        let s = s.trim();

        if s == "none" {
            return Ok(HIFValue::None);
        }
        if s == "true" {
            return Ok(HIFValue::Bool(true));
        }
        if s == "false" {
            return Ok(HIFValue::Bool(false));
        }

        // String literal
        if s.starts_with('"') && s.ends_with('"') {
            let inner = &s[1..s.len() - 1];
            return Ok(HIFValue::String(inner.replace("\\\"", "\"")));
        }

        // Float
        if s.contains('.') {
            if let Ok(f) = s.parse::<f64>() {
                return Ok(HIFValue::Float(f));
            }
        }

        // Int
        if let Ok(n) = s.parse::<i64>() {
            return Ok(HIFValue::Int(n));
        }

        Err(self.error(format!("invalid value: {}", s)))
    }

    /// Get the indentation level of a line.
    fn get_indent(&self, line: &str) -> usize {
        line.len() - line.trim_start().len()
    }

    /// Create an error at the current line.
    fn error(&self, message: impl Into<String>) -> HIFParseError {
        HIFParseError {
            message: message.into(),
            line: self.current,
        }
    }
}

/// Parse a HIF string into a HIFFile.
pub fn parse_hif(input: &str) -> HIFResult<HIFFile> {
    HIFParser::new(input).parse()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_struct() {
        let input = r#"# Haira Intent Format v1

struct User @abc123
  name: string
  age: int
  email: string?
"#;
        let file = parse_hif(input).unwrap();
        assert_eq!(file.version, 1);

        let user = file.get_struct("User").unwrap();
        assert_eq!(user.name, "User");
        assert_eq!(user.hash, "abc123");
        assert_eq!(user.fields.len(), 3);
        assert_eq!(user.fields[0].name, "name");
        assert_eq!(user.fields[0].ty, HIFType::String);
        assert_eq!(
            user.fields[2].ty,
            HIFType::Optional(Box::new(HIFType::String))
        );
    }

    #[test]
    fn test_parse_intent() {
        let input = r#"# Haira Intent Format v1

intent get_user_name @def456
  param user: User
  returns string
  body
    get_field user.name -> _name: string
    return _name
"#;
        let file = parse_hif(input).unwrap();

        let intent = file.get_intent("get_user_name").unwrap();
        assert_eq!(intent.name, "get_user_name");
        assert_eq!(intent.hash, "def456");
        assert_eq!(intent.params.len(), 1);
        assert_eq!(intent.params[0].name, "user");
        assert_eq!(intent.returns, HIFType::String);
        assert_eq!(intent.body.len(), 2);
    }

    #[test]
    fn test_roundtrip() {
        let input = r#"# Haira Intent Format v1

struct User @abc123
  name: string
  age: int

intent get_user_name @def456
  param user: User
  returns string
  body
    get_field user.name -> _name: string
    return _name
"#;
        let file = parse_hif(input).unwrap();
        let output = super::super::writer::write_hif(&file);

        // Parse the output again
        let file2 = parse_hif(&output).unwrap();

        assert_eq!(file.structs.len(), file2.structs.len());
        assert_eq!(file.intents.len(), file2.intents.len());
    }
}
