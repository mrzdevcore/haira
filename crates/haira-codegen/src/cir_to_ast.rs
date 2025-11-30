//! CIR to AST converter.
//!
//! This module converts AI-generated CIR (Canonical Intermediate Representation)
//! into Haira AST nodes that can be compiled normally.

use haira_ast::{
    AssignTarget, Assignment, BinaryExpr, BinaryOp, Block, CallExpr, ElseBranch, Expr, ExprKind,
    Field, FieldExpr, FunctionDef, IfStatement, InstanceExpr, InstanceField, Literal, Param,
    ReturnStatement, Span, Spanned, StatementKind, Type, TypeDef,
};
use haira_cir::{
    BinaryOperator, CIRFunction, CIROperation, CIRParam, CIRType, CIRTypeKind, CIRValue,
    TypeDefinition, UnaryOperator,
};
use smol_str::SmolStr;

/// Error during CIR to AST conversion.
#[derive(Debug, thiserror::Error)]
pub enum ConversionError {
    #[error("unsupported CIR operation: {0}")]
    UnsupportedOperation(String),
    #[error("invalid CIR value: {0}")]
    InvalidValue(String),
    #[error("invalid type: {0}")]
    InvalidType(String),
}

/// Create a dummy span for generated code.
fn dummy_span() -> Span {
    Span::new(0, 0)
}

/// Check if a CIR type represents "none" (void return).
fn is_none_type(ty: &CIRType) -> bool {
    match ty {
        CIRType::Simple(name) => name == "none" || name.is_empty(),
        CIRType::Complex(_) => false,
    }
}

/// Convert a CIR function to a Haira AST FunctionDef.
pub fn cir_to_function_def(cir: &CIRFunction) -> Result<FunctionDef, ConversionError> {
    // Convert parameters
    let params: Vec<Param> = cir
        .params
        .iter()
        .map(cir_param_to_ast)
        .collect::<Result<Vec<_>, _>>()?;

    // Convert return type
    let return_ty = if is_none_type(&cir.returns) {
        None
    } else {
        Some(cir_type_to_ast(&cir.returns)?)
    };

    // Convert body operations to statements
    let statements = cir_body_to_statements(&cir.body)?;

    Ok(FunctionDef {
        is_public: false,
        name: Spanned::new(SmolStr::from(&cir.name), dummy_span()),
        params,
        return_ty,
        body: Block {
            statements,
            span: dummy_span(),
        },
    })
}

/// Convert CIR type definitions to Haira AST TypeDefs.
pub fn cir_types_to_ast(types: &[TypeDefinition]) -> Result<Vec<TypeDef>, ConversionError> {
    types
        .iter()
        .map(|t| {
            let fields = t
                .fields
                .iter()
                .map(|f| Field {
                    name: Spanned::new(SmolStr::from(&f.name), dummy_span()),
                    ty: Some(Spanned::new(
                        Type::Named(SmolStr::from(&f.ty)),
                        dummy_span(),
                    )),
                    default: None,
                    span: dummy_span(),
                })
                .collect();

            Ok(TypeDef {
                is_public: false,
                name: Spanned::new(SmolStr::from(&t.name), dummy_span()),
                fields,
            })
        })
        .collect()
}

fn cir_param_to_ast(param: &CIRParam) -> Result<Param, ConversionError> {
    Ok(Param {
        name: Spanned::new(SmolStr::from(&param.name), dummy_span()),
        ty: Some(cir_type_to_ast(&param.ty)?),
        default: None,
        is_rest: false,
        span: dummy_span(),
    })
}

fn cir_type_to_ast(ty: &CIRType) -> Result<Spanned<Type>, ConversionError> {
    let ast_type = match ty {
        CIRType::Simple(name) => Type::Named(SmolStr::from(name)),
        CIRType::Complex(kind) => match kind {
            CIRTypeKind::List { element } => Type::List(Box::new(cir_type_to_ast(element)?)),
            CIRTypeKind::Option { inner } => Type::Option(Box::new(cir_type_to_ast(inner)?)),
            CIRTypeKind::Map { key, value } => Type::Map {
                key: Box::new(cir_type_to_ast(key)?),
                value: Box::new(cir_type_to_ast(value)?),
            },
            CIRTypeKind::Function { params, returns } => {
                let ast_params = params
                    .iter()
                    .map(cir_type_to_ast)
                    .collect::<Result<Vec<_>, _>>()?;
                Type::Function {
                    params: ast_params,
                    ret: Box::new(cir_type_to_ast(returns)?),
                }
            }
            CIRTypeKind::Union { variants } => {
                let ast_variants = variants
                    .iter()
                    .map(cir_type_to_ast)
                    .collect::<Result<Vec<_>, _>>()?;
                Type::Union(ast_variants)
            }
        },
    };

    Ok(Spanned::new(ast_type, dummy_span()))
}

fn cir_body_to_statements(
    ops: &[CIROperation],
) -> Result<Vec<Spanned<StatementKind>>, ConversionError> {
    let mut statements = Vec::new();

    for op in ops {
        match op {
            CIROperation::Return { value } => {
                let expr = cir_value_to_expr(value)?;
                statements.push(Spanned::new(
                    StatementKind::Return(ReturnStatement { values: vec![expr] }),
                    dummy_span(),
                ));
            }

            CIROperation::Literal { value, result } => {
                let expr = cir_value_to_expr(value)?;
                statements.push(make_assignment(result, expr));
            }

            CIROperation::Var { name, result } => {
                let expr = make_ident(name);
                statements.push(make_assignment(result, expr));
            }

            CIROperation::GetField {
                source,
                field,
                result,
            } => {
                let obj = make_ident(source);
                let expr = Spanned::new(
                    ExprKind::Field(FieldExpr {
                        object: Box::new(obj),
                        field: Spanned::new(SmolStr::from(field.as_str()), dummy_span()),
                    }),
                    dummy_span(),
                );
                statements.push(make_assignment(result, expr));
            }

            CIROperation::BinaryOp {
                op,
                left,
                right,
                result,
            } => {
                let left_expr = cir_value_to_expr(left)?;
                let right_expr = cir_value_to_expr(right)?;
                let bin_op = cir_binop_to_ast(op);

                let expr = Spanned::new(
                    ExprKind::Binary(BinaryExpr {
                        left: Box::new(left_expr),
                        op: Spanned::new(bin_op, dummy_span()),
                        right: Box::new(right_expr),
                    }),
                    dummy_span(),
                );
                statements.push(make_assignment(result, expr));
            }

            CIROperation::UnaryOp {
                op,
                operand,
                result,
            } => {
                let operand_expr = cir_value_to_expr(operand)?;
                let unary_op = match op {
                    UnaryOperator::Neg => haira_ast::UnaryOp::Neg,
                    UnaryOperator::Not => haira_ast::UnaryOp::Not,
                };

                let expr = Spanned::new(
                    ExprKind::Unary(haira_ast::UnaryExpr {
                        op: Spanned::new(unary_op, dummy_span()),
                        operand: Box::new(operand_expr),
                    }),
                    dummy_span(),
                );
                statements.push(make_assignment(result, expr));
            }

            CIROperation::Call {
                function,
                args,
                result,
            } => {
                let callee = make_ident(function);
                let call_args = args
                    .iter()
                    .map(|a| {
                        Ok(haira_ast::Argument {
                            name: None,
                            value: cir_value_to_expr(a)?,
                            span: dummy_span(),
                        })
                    })
                    .collect::<Result<Vec<_>, ConversionError>>()?;

                let expr = Spanned::new(
                    ExprKind::Call(CallExpr {
                        callee: Box::new(callee),
                        args: call_args,
                    }),
                    dummy_span(),
                );
                statements.push(make_assignment(result, expr));
            }

            CIROperation::Construct { ty, fields, result } => {
                let instance_fields = fields
                    .iter()
                    .map(|(name, value)| {
                        Ok(InstanceField {
                            name: Some(Spanned::new(SmolStr::from(name.as_str()), dummy_span())),
                            value: cir_value_to_expr(value)?,
                            span: dummy_span(),
                        })
                    })
                    .collect::<Result<Vec<_>, ConversionError>>()?;

                let expr = Spanned::new(
                    ExprKind::Instance(InstanceExpr {
                        type_name: Spanned::new(SmolStr::from(ty.as_str()), dummy_span()),
                        fields: instance_fields,
                    }),
                    dummy_span(),
                );
                statements.push(make_assignment(result, expr));
            }

            CIROperation::CreateList { elements, result } => {
                let list_elements = elements
                    .iter()
                    .map(cir_value_to_expr)
                    .collect::<Result<Vec<_>, ConversionError>>()?;

                let expr = Spanned::new(ExprKind::List(list_elements), dummy_span());
                statements.push(make_assignment(result, expr));
            }

            CIROperation::Count { source, result } => {
                // Convert count to a method call: source.len()
                let obj = make_ident(source);
                let expr = Spanned::new(
                    ExprKind::MethodCall(haira_ast::MethodCallExpr {
                        receiver: Box::new(obj),
                        method: Spanned::new(SmolStr::from("len"), dummy_span()),
                        args: vec![],
                    }),
                    dummy_span(),
                );
                statements.push(make_assignment(result, expr));
            }

            CIROperation::Sum { source, result } => {
                let obj = make_ident(source);
                let expr = Spanned::new(
                    ExprKind::MethodCall(haira_ast::MethodCallExpr {
                        receiver: Box::new(obj),
                        method: Spanned::new(SmolStr::from("sum"), dummy_span()),
                        args: vec![],
                    }),
                    dummy_span(),
                );
                statements.push(make_assignment(result, expr));
            }

            CIROperation::If {
                condition,
                then_ops,
                else_ops,
                result,
            } => {
                // Build condition from nested operations
                let (cond_stmts, cond_expr) = build_condition_expr(condition)?;
                statements.extend(cond_stmts);

                let then_stmts = cir_body_to_statements(then_ops)?;
                let else_stmts = cir_body_to_statements(else_ops)?;

                let if_stmt = IfStatement {
                    condition: cond_expr,
                    then_branch: Block {
                        statements: then_stmts,
                        span: dummy_span(),
                    },
                    else_branch: if else_stmts.is_empty() {
                        None
                    } else {
                        Some(ElseBranch::Block(Block {
                            statements: else_stmts,
                            span: dummy_span(),
                        }))
                    },
                };

                // If there's a result, wrap in expression context
                if !result.is_empty() {
                    let if_expr = Spanned::new(ExprKind::If(Box::new(if_stmt)), dummy_span());
                    statements.push(make_assignment(result, if_expr));
                } else {
                    statements.push(Spanned::new(StatementKind::If(if_stmt), dummy_span()));
                }
            }

            // For operations we don't fully support yet, generate a placeholder
            _ => {
                return Err(ConversionError::UnsupportedOperation(format!(
                    "{:?}",
                    std::mem::discriminant(op)
                )));
            }
        }
    }

    Ok(statements)
}

fn build_condition_expr(
    ops: &[CIROperation],
) -> Result<(Vec<Spanned<StatementKind>>, Expr), ConversionError> {
    if ops.is_empty() {
        return Ok((
            vec![],
            Spanned::new(ExprKind::Literal(Literal::Bool(true)), dummy_span()),
        ));
    }

    // For simple cases, just process and return last expression
    let mut stmts = Vec::new();
    let mut last_result = String::new();

    for op in ops {
        match op {
            CIROperation::BinaryOp {
                op: bin_op,
                left,
                right,
                result,
            } => {
                let left_expr = cir_value_to_expr(left)?;
                let right_expr = cir_value_to_expr(right)?;
                let ast_op = cir_binop_to_ast(bin_op);

                let expr = Spanned::new(
                    ExprKind::Binary(BinaryExpr {
                        left: Box::new(left_expr),
                        op: Spanned::new(ast_op, dummy_span()),
                        right: Box::new(right_expr),
                    }),
                    dummy_span(),
                );
                stmts.push(make_assignment(result, expr));
                last_result = result.clone();
            }
            CIROperation::Literal { value, result } => {
                let expr = cir_value_to_expr(value)?;
                stmts.push(make_assignment(result, expr));
                last_result = result.clone();
            }
            _ => {}
        }
    }

    // Return the last result as an identifier
    let expr = if last_result.is_empty() {
        Spanned::new(ExprKind::Literal(Literal::Bool(true)), dummy_span())
    } else {
        make_ident(&last_result)
    };

    Ok((stmts, expr))
}

fn cir_value_to_expr(value: &CIRValue) -> Result<Expr, ConversionError> {
    let kind = match value {
        CIRValue::Ref(name) => ExprKind::Identifier(SmolStr::from(name.as_str())),
        CIRValue::Int(n) => ExprKind::Literal(Literal::Int(*n)),
        CIRValue::Float(n) => ExprKind::Literal(Literal::Float(*n)),
        CIRValue::String(s) => ExprKind::Literal(Literal::String(SmolStr::from(s.as_str()))),
        CIRValue::Bool(b) => ExprKind::Literal(Literal::Bool(*b)),
        CIRValue::None => ExprKind::None,
        CIRValue::Operation(op) => {
            // Handle inline operations
            return cir_operation_to_expr(op);
        }
    };

    Ok(Spanned::new(kind, dummy_span()))
}

/// Convert an inline CIR operation to an expression.
fn cir_operation_to_expr(op: &CIROperation) -> Result<Expr, ConversionError> {
    match op {
        CIROperation::Format {
            template, values, ..
        } => {
            // Convert format to string interpolation or concatenation
            // For simplicity, use a format call or string concat
            if values.is_empty() {
                // Just a plain string
                return Ok(Spanned::new(
                    ExprKind::Literal(Literal::String(SmolStr::from(template.as_str()))),
                    dummy_span(),
                ));
            }

            // Build string concatenation: "Hello, " + name + "!"
            // Parse template and interleave with values
            let parts: Vec<&str> = template.split("{}").collect();
            let mut exprs: Vec<Expr> = Vec::new();

            for (i, part) in parts.iter().enumerate() {
                if !part.is_empty() {
                    exprs.push(Spanned::new(
                        ExprKind::Literal(Literal::String(SmolStr::from(*part))),
                        dummy_span(),
                    ));
                }
                // Add the value if we have one for this position
                if let Some(val) = values.get(&i.to_string()) {
                    exprs.push(cir_value_to_expr(val)?);
                }
            }

            // Combine all exprs with + operator
            if exprs.is_empty() {
                return Ok(Spanned::new(
                    ExprKind::Literal(Literal::String(SmolStr::from(""))),
                    dummy_span(),
                ));
            }

            let mut result = exprs.remove(0);
            for expr in exprs {
                result = Spanned::new(
                    ExprKind::Binary(BinaryExpr {
                        left: Box::new(result),
                        op: Spanned::new(BinaryOp::Add, dummy_span()),
                        right: Box::new(expr),
                    }),
                    dummy_span(),
                );
            }

            Ok(result)
        }

        CIROperation::Concat { parts, .. } => {
            // Convert to string concatenation
            let mut exprs: Vec<Expr> = parts
                .iter()
                .map(cir_value_to_expr)
                .collect::<Result<Vec<_>, _>>()?;

            if exprs.is_empty() {
                return Ok(Spanned::new(
                    ExprKind::Literal(Literal::String(SmolStr::from(""))),
                    dummy_span(),
                ));
            }

            let mut result = exprs.remove(0);
            for expr in exprs {
                result = Spanned::new(
                    ExprKind::Binary(BinaryExpr {
                        left: Box::new(result),
                        op: Spanned::new(BinaryOp::Add, dummy_span()),
                        right: Box::new(expr),
                    }),
                    dummy_span(),
                );
            }

            Ok(result)
        }

        CIROperation::GetField { source, field, .. } => {
            // If field is empty, this is just a variable reference
            if field.is_empty() {
                return Ok(make_ident(source));
            }
            let obj = make_ident(source);
            Ok(Spanned::new(
                ExprKind::Field(FieldExpr {
                    object: Box::new(obj),
                    field: Spanned::new(SmolStr::from(field.as_str()), dummy_span()),
                }),
                dummy_span(),
            ))
        }

        CIROperation::Var { name, .. } => Ok(make_ident(name)),

        CIROperation::Literal { value, .. } => cir_value_to_expr(value),

        CIROperation::BinaryOp {
            op, left, right, ..
        } => {
            let left_expr = cir_value_to_expr(left)?;
            let right_expr = cir_value_to_expr(right)?;
            let bin_op = cir_binop_to_ast(op);

            Ok(Spanned::new(
                ExprKind::Binary(BinaryExpr {
                    left: Box::new(left_expr),
                    op: Spanned::new(bin_op, dummy_span()),
                    right: Box::new(right_expr),
                }),
                dummy_span(),
            ))
        }

        CIROperation::Call { function, args, .. } => {
            let callee = make_ident(function);
            let call_args = args
                .iter()
                .map(|a| {
                    Ok(haira_ast::Argument {
                        name: None,
                        value: cir_value_to_expr(a)?,
                        span: dummy_span(),
                    })
                })
                .collect::<Result<Vec<_>, ConversionError>>()?;

            Ok(Spanned::new(
                ExprKind::Call(CallExpr {
                    callee: Box::new(callee),
                    args: call_args,
                }),
                dummy_span(),
            ))
        }

        _ => Err(ConversionError::UnsupportedOperation(format!(
            "inline operation {:?}",
            std::mem::discriminant(op)
        ))),
    }
}

fn cir_binop_to_ast(op: &BinaryOperator) -> BinaryOp {
    match op {
        BinaryOperator::Add => BinaryOp::Add,
        BinaryOperator::Sub => BinaryOp::Sub,
        BinaryOperator::Mul => BinaryOp::Mul,
        BinaryOperator::Div => BinaryOp::Div,
        BinaryOperator::Mod => BinaryOp::Mod,
        BinaryOperator::Eq => BinaryOp::Eq,
        BinaryOperator::Ne => BinaryOp::Ne,
        BinaryOperator::Lt => BinaryOp::Lt,
        BinaryOperator::Gt => BinaryOp::Gt,
        BinaryOperator::Le => BinaryOp::Le,
        BinaryOperator::Ge => BinaryOp::Ge,
        BinaryOperator::And => BinaryOp::And,
        BinaryOperator::Or => BinaryOp::Or,
    }
}

fn make_ident(name: &str) -> Expr {
    Spanned::new(ExprKind::Identifier(SmolStr::from(name)), dummy_span())
}

fn make_assignment(target: &str, value: Expr) -> Spanned<StatementKind> {
    Spanned::new(
        StatementKind::Assignment(Assignment {
            targets: vec![AssignTarget {
                name: Spanned::new(SmolStr::from(target), dummy_span()),
                ty: None,
            }],
            value,
        }),
        dummy_span(),
    )
}

#[cfg(test)]
mod tests {
    use super::*;
    use haira_cir::{CIRFunction, CIROperation, CIRType, CIRValue};

    #[test]
    fn test_simple_function_conversion() {
        let cir = CIRFunction::new("add_numbers")
            .with_param("a", CIRType::simple("int"))
            .with_param("b", CIRType::simple("int"))
            .returning(CIRType::simple("int"))
            .with_op(CIROperation::BinaryOp {
                op: BinaryOperator::Add,
                left: CIRValue::Ref("a".to_string()),
                right: CIRValue::Ref("b".to_string()),
                result: "sum".to_string(),
            })
            .with_op(CIROperation::Return {
                value: CIRValue::Ref("sum".to_string()),
            });

        let func_def = cir_to_function_def(&cir).unwrap();
        assert_eq!(func_def.name.node.as_str(), "add_numbers");
        assert_eq!(func_def.params.len(), 2);
        assert_eq!(func_def.body.statements.len(), 2);
    }

    #[test]
    fn test_literal_values() {
        let cir = CIRFunction::new("return_constant")
            .returning(CIRType::simple("int"))
            .with_op(CIROperation::Return {
                value: CIRValue::Int(42),
            });

        let func_def = cir_to_function_def(&cir).unwrap();
        assert_eq!(func_def.body.statements.len(), 1);
    }
}
