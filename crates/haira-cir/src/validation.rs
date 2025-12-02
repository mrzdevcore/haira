//! CIR Validation - ensure AI output is well-formed.

use crate::{CIRFunction, CIROperation, CIRValue};
use std::collections::HashSet;
use thiserror::Error;

/// Validation errors.
#[derive(Debug, Error)]
pub enum ValidationError {
    #[error("undefined variable: {0}")]
    UndefinedVariable(String),

    #[error("duplicate result variable: {0}")]
    DuplicateResult(String),

    #[error("missing return statement")]
    MissingReturn,

    #[error("invalid operation: {0}")]
    InvalidOperation(String),

    #[error("type mismatch: expected {expected}, found {found}")]
    TypeMismatch { expected: String, found: String },

    #[error("empty body")]
    EmptyBody,
}

/// Validate a CIR function.
pub fn validate(func: &CIRFunction) -> Result<(), Vec<ValidationError>> {
    let mut errors = Vec::new();
    let mut defined_vars: HashSet<String> = HashSet::new();

    // Parameters are initially defined
    for param in &func.params {
        defined_vars.insert(param.name.clone());
    }

    // Check for empty body
    if func.body.is_empty() {
        errors.push(ValidationError::EmptyBody);
        return Err(errors);
    }

    // Validate each operation
    for op in &func.body {
        validate_operation(op, &mut defined_vars, &mut errors);
    }

    // Check that last operation is a return (or the function returns none)
    let has_return = func
        .body
        .iter()
        .any(|op| matches!(op, CIROperation::Return { .. }));
    if !has_return && !matches!(&func.returns, crate::CIRType::Simple(s) if s == "none") {
        errors.push(ValidationError::MissingReturn);
    }

    if errors.is_empty() {
        Ok(())
    } else {
        Err(errors)
    }
}

fn validate_operation(
    op: &CIROperation,
    defined: &mut HashSet<String>,
    errors: &mut Vec<ValidationError>,
) {
    match op {
        CIROperation::GetField { source, result, .. } => {
            check_defined(source, defined, errors);
            define_var(result, defined, errors);
        }
        CIROperation::GetIndex {
            source,
            index,
            result,
        } => {
            check_defined(source, defined, errors);
            check_value(index, defined, errors);
            define_var(result, defined, errors);
        }
        CIROperation::SetField { target, value, .. } => {
            check_defined(target, defined, errors);
            check_value(value, defined, errors);
        }
        CIROperation::Map {
            source,
            element_var,
            transform,
            result,
        } => {
            check_defined(source, defined, errors);
            let mut inner_defined = defined.clone();
            inner_defined.insert(element_var.clone());
            for inner_op in transform {
                validate_operation(inner_op, &mut inner_defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::Filter {
            source,
            element_var,
            predicate,
            result,
        } => {
            check_defined(source, defined, errors);
            let mut inner_defined = defined.clone();
            inner_defined.insert(element_var.clone());
            for inner_op in predicate {
                validate_operation(inner_op, &mut inner_defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::Reduce {
            source,
            initial,
            accumulator_var,
            element_var,
            reducer,
            result,
        } => {
            check_defined(source, defined, errors);
            check_value(initial, defined, errors);
            let mut inner_defined = defined.clone();
            inner_defined.insert(accumulator_var.clone());
            inner_defined.insert(element_var.clone());
            for inner_op in reducer {
                validate_operation(inner_op, &mut inner_defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::GroupBy {
            source,
            element_var,
            key,
            result,
        } => {
            check_defined(source, defined, errors);
            let mut inner_defined = defined.clone();
            inner_defined.insert(element_var.clone());
            for inner_op in key {
                validate_operation(inner_op, &mut inner_defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::Sort {
            source,
            element_var,
            key,
            result,
            ..
        } => {
            check_defined(source, defined, errors);
            let mut inner_defined = defined.clone();
            inner_defined.insert(element_var.clone());
            for inner_op in key {
                validate_operation(inner_op, &mut inner_defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::Take {
            source,
            count,
            result,
        } => {
            check_defined(source, defined, errors);
            check_value(count, defined, errors);
            define_var(result, defined, errors);
        }
        CIROperation::Skip {
            source,
            count,
            result,
        } => {
            check_defined(source, defined, errors);
            check_value(count, defined, errors);
            define_var(result, defined, errors);
        }
        CIROperation::Count { source, result } => {
            check_defined(source, defined, errors);
            define_var(result, defined, errors);
        }
        CIROperation::Find {
            source,
            element_var,
            predicate,
            result,
        } => {
            check_defined(source, defined, errors);
            let mut inner_defined = defined.clone();
            inner_defined.insert(element_var.clone());
            for inner_op in predicate {
                validate_operation(inner_op, &mut inner_defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::Any {
            source,
            element_var,
            predicate,
            result,
        }
        | CIROperation::All {
            source,
            element_var,
            predicate,
            result,
        } => {
            check_defined(source, defined, errors);
            let mut inner_defined = defined.clone();
            inner_defined.insert(element_var.clone());
            for inner_op in predicate {
                validate_operation(inner_op, &mut inner_defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::Sum { source, result }
        | CIROperation::Min { source, result }
        | CIROperation::Max { source, result }
        | CIROperation::Avg { source, result } => {
            check_defined(source, defined, errors);
            define_var(result, defined, errors);
        }
        CIROperation::MaxBy {
            source,
            element_var,
            key,
            result,
        }
        | CIROperation::MinBy {
            source,
            element_var,
            key,
            result,
        } => {
            check_defined(source, defined, errors);
            let mut inner_defined = defined.clone();
            inner_defined.insert(element_var.clone());
            for inner_op in key {
                validate_operation(inner_op, &mut inner_defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::If {
            condition,
            then_ops,
            else_ops,
            result,
        } => {
            for inner_op in condition {
                validate_operation(inner_op, defined, errors);
            }
            for inner_op in then_ops {
                validate_operation(inner_op, defined, errors);
            }
            for inner_op in else_ops {
                validate_operation(inner_op, defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::Match {
            subject,
            arms,
            result,
        } => {
            check_defined(subject, defined, errors);
            for arm in arms {
                let mut inner_defined = defined.clone();
                // Add bindings from pattern
                if let crate::CIRPattern::Binding { name } = &arm.pattern {
                    inner_defined.insert(name.clone());
                }
                if let crate::CIRPattern::Constructor { fields, .. } = &arm.pattern {
                    for field in fields {
                        inner_defined.insert(field.clone());
                    }
                }
                for inner_op in &arm.body {
                    validate_operation(inner_op, &mut inner_defined, errors);
                }
            }
            define_var(result, defined, errors);
        }
        CIROperation::Loop {
            source,
            element_var,
            body,
            result,
        } => {
            check_defined(source, defined, errors);
            let mut inner_defined = defined.clone();
            inner_defined.insert(element_var.clone());
            for inner_op in body {
                validate_operation(inner_op, &mut inner_defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::Construct { fields, result, .. } => {
            for value in fields.values() {
                check_value(value, defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::CreateList { elements, result } => {
            for elem in elements {
                check_value(elem, defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::CreateMap { entries, result } => {
            for (k, v) in entries {
                check_value(k, defined, errors);
                check_value(v, defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::BinaryOp {
            left,
            right,
            result,
            ..
        } => {
            check_value(left, defined, errors);
            check_value(right, defined, errors);
            define_var(result, defined, errors);
        }
        CIROperation::UnaryOp {
            operand, result, ..
        } => {
            check_value(operand, defined, errors);
            define_var(result, defined, errors);
        }
        CIROperation::Call { args, result, .. } => {
            for arg in args {
                check_value(arg, defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::Literal { result, .. } => {
            define_var(result, defined, errors);
        }
        CIROperation::Var { name, result } => {
            check_defined(name, defined, errors);
            define_var(result, defined, errors);
        }
        CIROperation::DbQuery { result, .. } => {
            define_var(result, defined, errors);
        }
        CIROperation::HttpRequest {
            url, body, result, ..
        } => {
            check_value(url, defined, errors);
            if let Some(b) = body {
                check_value(b, defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::FileRead { path, result } => {
            check_value(path, defined, errors);
            define_var(result, defined, errors);
        }
        CIROperation::FileWrite { path, content } => {
            check_value(path, defined, errors);
            check_value(content, defined, errors);
        }
        CIROperation::Format { values, result, .. } => {
            for value in values.values() {
                check_value(value, defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::Concat { parts, result } => {
            for part in parts {
                check_value(part, defined, errors);
            }
            define_var(result, defined, errors);
        }
        CIROperation::Return { value } => {
            check_value(value, defined, errors);
        }
    }
}

fn check_defined(name: &str, defined: &HashSet<String>, errors: &mut Vec<ValidationError>) {
    if !defined.contains(name) {
        errors.push(ValidationError::UndefinedVariable(name.to_string()));
    }
}

fn define_var(name: &str, defined: &mut HashSet<String>, _errors: &mut Vec<ValidationError>) {
    if defined.contains(name) {
        // Allow shadowing - this is fine
    }
    defined.insert(name.to_string());
}

fn check_value(value: &CIRValue, defined: &HashSet<String>, errors: &mut Vec<ValidationError>) {
    match value {
        CIRValue::Ref(name) => check_defined(name, defined, errors),
        CIRValue::Operation(op) => {
            let mut inner = defined.clone();
            validate_operation(op, &mut inner, errors);
        }
        _ => {}
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::CIRType;

    #[test]
    fn test_valid_function() {
        let func = CIRFunction {
            name: "test".to_string(),
            description: None,
            params: vec![crate::CIRParam {
                name: "x".to_string(),
                ty: CIRType::simple("int"),
                default: None,
            }],
            returns: CIRType::simple("int"),
            new_types: vec![],
            body: vec![
                CIROperation::BinaryOp {
                    op: crate::BinaryOperator::Add,
                    left: CIRValue::var("x"),
                    right: CIRValue::Int(1),
                    result: "result".to_string(),
                },
                CIROperation::Return {
                    value: CIRValue::var("result"),
                },
            ],
        };

        assert!(validate(&func).is_ok());
    }

    #[test]
    fn test_undefined_variable() {
        let func = CIRFunction {
            name: "test".to_string(),
            description: None,
            params: vec![],
            returns: CIRType::simple("int"),
            new_types: vec![],
            body: vec![CIROperation::Return {
                value: CIRValue::var("undefined"),
            }],
        };

        let result = validate(&func);
        assert!(result.is_err());
        let errors = result.unwrap_err();
        assert!(errors
            .iter()
            .any(|e| matches!(e, ValidationError::UndefinedVariable(_))));
    }

    #[test]
    fn test_missing_return() {
        let func = CIRFunction {
            name: "test".to_string(),
            description: None,
            params: vec![crate::CIRParam {
                name: "x".to_string(),
                ty: CIRType::simple("int"),
                default: None,
            }],
            returns: CIRType::simple("int"),
            new_types: vec![],
            body: vec![CIROperation::Literal {
                value: CIRValue::Int(42),
                result: "x".to_string(),
            }],
        };

        let result = validate(&func);
        assert!(result.is_err());
        let errors = result.unwrap_err();
        assert!(errors
            .iter()
            .any(|e| matches!(e, ValidationError::MissingReturn)));
    }
}
