//! HIF Inference - Converts between CIR and HIF formats.
//!
//! This module provides bidirectional conversion between:
//! - CIR (Canonical IR) - JSON-based format from AI
//! - HIF (Haira Intent Format) - Human-readable cache format

use super::types::*;
use haira_cir::{
    BinaryOperator, CIRFunction, CIROperation, CIRParam, CIRType, CIRTypeKind, CIRValue,
    FieldDefinition, TypeDefinition, UnaryOperator,
};

/// Convert a CIR type to HIF type.
pub fn cir_type_to_hif(ty: &CIRType) -> HIFType {
    match ty {
        CIRType::Simple(name) => match name.as_str() {
            "int" | "i32" | "i64" | "integer" => HIFType::Int,
            "float" | "f32" | "f64" | "double" => HIFType::Float,
            "string" | "str" => HIFType::String,
            "bool" | "boolean" => HIFType::Bool,
            "void" | "none" | "()" => HIFType::Void,
            "datetime" | "date" | "time" => HIFType::DateTime,
            other => HIFType::Struct(other.to_string()),
        },
        CIRType::Complex(kind) => match kind {
            CIRTypeKind::List { element } => HIFType::Array(Box::new(cir_type_to_hif(element))),
            CIRTypeKind::Map { key, value } => HIFType::Map(
                Box::new(cir_type_to_hif(key)),
                Box::new(cir_type_to_hif(value)),
            ),
            CIRTypeKind::Option { inner } => HIFType::Optional(Box::new(cir_type_to_hif(inner))),
            CIRTypeKind::Function { .. } => HIFType::Unknown,
            CIRTypeKind::Union { .. } => HIFType::Unknown,
        },
    }
}

/// Convert a HIF type to CIR type.
pub fn hif_type_to_cir(ty: &HIFType) -> CIRType {
    match ty {
        HIFType::Int => CIRType::simple("int"),
        HIFType::Float => CIRType::simple("float"),
        HIFType::String => CIRType::simple("string"),
        HIFType::Bool => CIRType::simple("bool"),
        HIFType::Void => CIRType::simple("none"),
        HIFType::DateTime => CIRType::simple("datetime"),
        HIFType::Array(inner) => CIRType::list(hif_type_to_cir(inner)),
        HIFType::Optional(inner) => CIRType::option(hif_type_to_cir(inner)),
        HIFType::Map(key, value) => CIRType::map(hif_type_to_cir(key), hif_type_to_cir(value)),
        HIFType::Struct(name) => CIRType::simple(name),
        HIFType::Unknown => CIRType::simple("unknown"),
    }
}

/// Convert a CIR function to HIF intent.
pub fn cir_function_to_hif_intent(func: &CIRFunction, hash: &str) -> HIFIntent {
    let params = func
        .params
        .iter()
        .map(|p| HIFParam {
            name: p.name.clone(),
            ty: cir_type_to_hif(&p.ty),
        })
        .collect();

    let body = func.body.iter().map(cir_operation_to_hif).collect();

    HIFIntent {
        name: func.name.clone(),
        hash: hash.to_string(),
        params,
        returns: cir_type_to_hif(&func.returns),
        body,
    }
}

/// Convert a HIF intent to CIR function.
pub fn hif_intent_to_cir_function(intent: &HIFIntent) -> CIRFunction {
    let params = intent
        .params
        .iter()
        .map(|p| CIRParam {
            name: p.name.clone(),
            ty: hif_type_to_cir(&p.ty),
            default: None,
        })
        .collect();

    let body = intent.body.iter().map(hif_operation_to_cir).collect();

    CIRFunction {
        name: intent.name.clone(),
        description: None,
        params,
        returns: hif_type_to_cir(&intent.returns),
        new_types: Vec::new(),
        body,
    }
}

/// Convert a CIR type definition to HIF struct.
pub fn cir_type_def_to_hif_struct(ty: &TypeDefinition, hash: &str) -> HIFStruct {
    let fields = ty
        .fields
        .iter()
        .map(|f| {
            let base_ty = HIFType::parse(&f.ty);
            let field_ty = if f.optional {
                HIFType::Optional(Box::new(base_ty))
            } else {
                base_ty
            };
            HIFField {
                name: f.name.clone(),
                ty: field_ty,
            }
        })
        .collect();

    HIFStruct {
        name: ty.name.clone(),
        hash: hash.to_string(),
        fields,
    }
}

/// Convert a HIF struct to CIR type definition.
pub fn hif_struct_to_cir_type_def(s: &HIFStruct) -> TypeDefinition {
    let fields = s
        .fields
        .iter()
        .map(|f| {
            let (ty_str, optional) = match &f.ty {
                HIFType::Optional(inner) => (inner.to_hif_string(), true),
                other => (other.to_hif_string(), false),
            };
            FieldDefinition {
                name: f.name.clone(),
                ty: ty_str,
                optional,
                default: None,
            }
        })
        .collect();

    TypeDefinition {
        name: s.name.clone(),
        fields,
    }
}

/// Convert a CIR operation to HIF operation.
pub fn cir_operation_to_hif(op: &CIROperation) -> HIFOperation {
    match op {
        CIROperation::GetField {
            source,
            field,
            result,
        } => HIFOperation {
            kind: HIFOpKind::GetField(source.clone(), field.clone()),
            result: Some(result.clone()),
            result_type: None,
        },

        CIROperation::GetIndex {
            source,
            index,
            result,
        } => HIFOperation {
            kind: HIFOpKind::GetIndex(source.clone(), cir_value_to_string(index)),
            result: Some(result.clone()),
            result_type: None,
        },

        CIROperation::SetField {
            target,
            field,
            value,
        } => HIFOperation {
            kind: HIFOpKind::SetField(target.clone(), field.clone(), cir_value_to_string(value)),
            result: None,
            result_type: None,
        },

        CIROperation::Map {
            source,
            element_var,
            transform,
            result,
        } => HIFOperation {
            kind: HIFOpKind::Map {
                source: source.clone(),
                element_var: element_var.clone(),
                body: transform.iter().map(cir_operation_to_hif).collect(),
            },
            result: Some(result.clone()),
            result_type: None,
        },

        CIROperation::Filter {
            source,
            element_var,
            predicate,
            result,
        } => HIFOperation {
            kind: HIFOpKind::Filter {
                source: source.clone(),
                element_var: element_var.clone(),
                body: predicate.iter().map(cir_operation_to_hif).collect(),
            },
            result: Some(result.clone()),
            result_type: None,
        },

        CIROperation::Reduce {
            source,
            initial,
            accumulator_var,
            element_var,
            reducer,
            result,
        } => HIFOperation {
            kind: HIFOpKind::Reduce {
                source: source.clone(),
                initial: cir_value_to_string(initial),
                accumulator_var: accumulator_var.clone(),
                element_var: element_var.clone(),
                body: reducer.iter().map(cir_operation_to_hif).collect(),
            },
            result: Some(result.clone()),
            result_type: None,
        },

        CIROperation::Take {
            source,
            count,
            result,
        } => HIFOperation {
            kind: HIFOpKind::Take(source.clone(), cir_value_to_string(count)),
            result: Some(result.clone()),
            result_type: None,
        },

        CIROperation::Skip {
            source,
            count,
            result,
        } => HIFOperation {
            kind: HIFOpKind::Skip(source.clone(), cir_value_to_string(count)),
            result: Some(result.clone()),
            result_type: None,
        },

        CIROperation::Count { source, result } => HIFOperation {
            kind: HIFOpKind::Count(source.clone()),
            result: Some(result.clone()),
            result_type: Some(HIFType::Int),
        },

        CIROperation::Find {
            source,
            element_var,
            predicate,
            result,
        } => HIFOperation {
            kind: HIFOpKind::Find {
                source: source.clone(),
                element_var: element_var.clone(),
                body: predicate.iter().map(cir_operation_to_hif).collect(),
            },
            result: Some(result.clone()),
            result_type: None,
        },

        CIROperation::Any {
            source,
            element_var,
            predicate,
            result,
        } => HIFOperation {
            kind: HIFOpKind::Any {
                source: source.clone(),
                element_var: element_var.clone(),
                body: predicate.iter().map(cir_operation_to_hif).collect(),
            },
            result: Some(result.clone()),
            result_type: Some(HIFType::Bool),
        },

        CIROperation::All {
            source,
            element_var,
            predicate,
            result,
        } => HIFOperation {
            kind: HIFOpKind::All {
                source: source.clone(),
                element_var: element_var.clone(),
                body: predicate.iter().map(cir_operation_to_hif).collect(),
            },
            result: Some(result.clone()),
            result_type: Some(HIFType::Bool),
        },

        CIROperation::Sum { source, result } => HIFOperation {
            kind: HIFOpKind::Sum(source.clone()),
            result: Some(result.clone()),
            result_type: None,
        },

        CIROperation::Min { source, result } => HIFOperation {
            kind: HIFOpKind::Min(source.clone()),
            result: Some(result.clone()),
            result_type: None,
        },

        CIROperation::Max { source, result } => HIFOperation {
            kind: HIFOpKind::Max(source.clone()),
            result: Some(result.clone()),
            result_type: None,
        },

        CIROperation::Avg { source, result } => HIFOperation {
            kind: HIFOpKind::Avg(source.clone()),
            result: Some(result.clone()),
            result_type: Some(HIFType::Float),
        },

        CIROperation::If {
            condition,
            then_ops,
            else_ops,
            result,
        } => HIFOperation {
            kind: HIFOpKind::If {
                condition: condition.iter().map(cir_operation_to_hif).collect(),
                then_ops: then_ops.iter().map(cir_operation_to_hif).collect(),
                else_ops: else_ops.iter().map(cir_operation_to_hif).collect(),
            },
            result: Some(result.clone()),
            result_type: None,
        },

        CIROperation::Loop {
            source,
            element_var,
            body,
            result,
        } => HIFOperation {
            kind: HIFOpKind::Loop {
                source: source.clone(),
                element_var: element_var.clone(),
                body: body.iter().map(cir_operation_to_hif).collect(),
            },
            result: Some(result.clone()),
            result_type: None,
        },

        CIROperation::Construct { ty, fields, result } => HIFOperation {
            kind: HIFOpKind::Construct {
                ty: ty.clone(),
                fields: fields
                    .iter()
                    .map(|(k, v)| (k.clone(), cir_value_to_string(v)))
                    .collect(),
            },
            result: Some(result.clone()),
            result_type: Some(HIFType::Struct(ty.clone())),
        },

        CIROperation::CreateList { elements, result } => HIFOperation {
            kind: HIFOpKind::CreateList(elements.iter().map(cir_value_to_string).collect()),
            result: Some(result.clone()),
            result_type: None,
        },

        CIROperation::BinaryOp {
            op,
            left,
            right,
            result,
        } => {
            let left_str = cir_value_to_string(left);
            let right_str = cir_value_to_string(right);
            let kind = match op {
                BinaryOperator::Add => HIFOpKind::Add(left_str, right_str),
                BinaryOperator::Sub => HIFOpKind::Sub(left_str, right_str),
                BinaryOperator::Mul => HIFOpKind::Mul(left_str, right_str),
                BinaryOperator::Div => HIFOpKind::Div(left_str, right_str),
                BinaryOperator::Mod => HIFOpKind::Mod(left_str, right_str),
                BinaryOperator::Eq => HIFOpKind::Eq(left_str, right_str),
                BinaryOperator::Ne => HIFOpKind::Ne(left_str, right_str),
                BinaryOperator::Lt => HIFOpKind::Lt(left_str, right_str),
                BinaryOperator::Gt => HIFOpKind::Gt(left_str, right_str),
                BinaryOperator::Le => HIFOpKind::Le(left_str, right_str),
                BinaryOperator::Ge => HIFOpKind::Ge(left_str, right_str),
                BinaryOperator::And => HIFOpKind::And(left_str, right_str),
                BinaryOperator::Or => HIFOpKind::Or(left_str, right_str),
            };
            HIFOperation {
                kind,
                result: Some(result.clone()),
                result_type: None,
            }
        }

        CIROperation::UnaryOp {
            op,
            operand,
            result,
        } => {
            let operand_str = cir_value_to_string(operand);
            let kind = match op {
                UnaryOperator::Neg => HIFOpKind::Neg(operand_str),
                UnaryOperator::Not => HIFOpKind::Not(operand_str),
            };
            HIFOperation {
                kind,
                result: Some(result.clone()),
                result_type: None,
            }
        }

        CIROperation::Call {
            function,
            args,
            result,
        } => HIFOperation {
            kind: HIFOpKind::Call {
                function: function.clone(),
                args: args.iter().map(cir_value_to_string).collect(),
            },
            result: Some(result.clone()),
            result_type: None,
        },

        CIROperation::Literal { value, result } => HIFOperation {
            kind: HIFOpKind::Literal(cir_value_to_hif_value(value)),
            result: Some(result.clone()),
            result_type: Some(cir_value_to_hif_value(value).infer_type()),
        },

        CIROperation::Var { name, result } => HIFOperation {
            kind: HIFOpKind::Var(name.clone()),
            result: Some(result.clone()),
            result_type: None,
        },

        CIROperation::Return { value } => HIFOperation {
            kind: HIFOpKind::Return(cir_value_to_string(value)),
            result: None,
            result_type: None,
        },

        CIROperation::Format {
            template,
            values,
            result,
        } => HIFOperation {
            kind: HIFOpKind::Format {
                template: template.clone(),
                values: values
                    .iter()
                    .map(|(k, v)| (k.clone(), cir_value_to_string(v)))
                    .collect(),
            },
            result: Some(result.clone()),
            result_type: Some(HIFType::String),
        },

        CIROperation::Concat { parts, result } => HIFOperation {
            kind: HIFOpKind::Concat(parts.iter().map(cir_value_to_string).collect()),
            result: Some(result.clone()),
            result_type: Some(HIFType::String),
        },

        // Operations not directly mapped - create a placeholder
        _ => HIFOperation {
            kind: HIFOpKind::Literal(HIFValue::None),
            result: None,
            result_type: None,
        },
    }
}

/// Convert a HIF operation to CIR operation.
pub fn hif_operation_to_cir(op: &HIFOperation) -> CIROperation {
    let result = op.result.clone().unwrap_or_else(|| "_".to_string());

    match &op.kind {
        HIFOpKind::GetField(source, field) => CIROperation::GetField {
            source: source.clone(),
            field: field.clone(),
            result,
        },

        HIFOpKind::GetIndex(source, index) => CIROperation::GetIndex {
            source: source.clone(),
            index: string_to_cir_value(index),
            result,
        },

        HIFOpKind::SetField(target, field, value) => CIROperation::SetField {
            target: target.clone(),
            field: field.clone(),
            value: string_to_cir_value(value),
        },

        HIFOpKind::Map {
            source,
            element_var,
            body,
        } => CIROperation::Map {
            source: source.clone(),
            element_var: element_var.clone(),
            transform: body.iter().map(hif_operation_to_cir).collect(),
            result,
        },

        HIFOpKind::Filter {
            source,
            element_var,
            body,
        } => CIROperation::Filter {
            source: source.clone(),
            element_var: element_var.clone(),
            predicate: body.iter().map(hif_operation_to_cir).collect(),
            result,
        },

        HIFOpKind::Reduce {
            source,
            initial,
            accumulator_var,
            element_var,
            body,
        } => CIROperation::Reduce {
            source: source.clone(),
            initial: string_to_cir_value(initial),
            accumulator_var: accumulator_var.clone(),
            element_var: element_var.clone(),
            reducer: body.iter().map(hif_operation_to_cir).collect(),
            result,
        },

        HIFOpKind::Take(source, count) => CIROperation::Take {
            source: source.clone(),
            count: string_to_cir_value(count),
            result,
        },

        HIFOpKind::Skip(source, count) => CIROperation::Skip {
            source: source.clone(),
            count: string_to_cir_value(count),
            result,
        },

        HIFOpKind::Sum(source) => CIROperation::Sum {
            source: source.clone(),
            result,
        },

        HIFOpKind::Min(source) => CIROperation::Min {
            source: source.clone(),
            result,
        },

        HIFOpKind::Max(source) => CIROperation::Max {
            source: source.clone(),
            result,
        },

        HIFOpKind::Avg(source) => CIROperation::Avg {
            source: source.clone(),
            result,
        },

        HIFOpKind::Count(source) => CIROperation::Count {
            source: source.clone(),
            result,
        },

        HIFOpKind::Find {
            source,
            element_var,
            body,
        } => CIROperation::Find {
            source: source.clone(),
            element_var: element_var.clone(),
            predicate: body.iter().map(hif_operation_to_cir).collect(),
            result,
        },

        HIFOpKind::Any {
            source,
            element_var,
            body,
        } => CIROperation::Any {
            source: source.clone(),
            element_var: element_var.clone(),
            predicate: body.iter().map(hif_operation_to_cir).collect(),
            result,
        },

        HIFOpKind::All {
            source,
            element_var,
            body,
        } => CIROperation::All {
            source: source.clone(),
            element_var: element_var.clone(),
            predicate: body.iter().map(hif_operation_to_cir).collect(),
            result,
        },

        HIFOpKind::If {
            condition,
            then_ops,
            else_ops,
        } => CIROperation::If {
            condition: condition.iter().map(hif_operation_to_cir).collect(),
            then_ops: then_ops.iter().map(hif_operation_to_cir).collect(),
            else_ops: else_ops.iter().map(hif_operation_to_cir).collect(),
            result,
        },

        HIFOpKind::Loop {
            source,
            element_var,
            body,
        } => CIROperation::Loop {
            source: source.clone(),
            element_var: element_var.clone(),
            body: body.iter().map(hif_operation_to_cir).collect(),
            result,
        },

        HIFOpKind::Construct { ty, fields } => CIROperation::Construct {
            ty: ty.clone(),
            fields: fields
                .iter()
                .map(|(k, v)| (k.clone(), string_to_cir_value(v)))
                .collect(),
            result,
        },

        HIFOpKind::CreateList(elements) => CIROperation::CreateList {
            elements: elements.iter().map(|s| string_to_cir_value(s)).collect(),
            result,
        },

        HIFOpKind::Add(left, right) => CIROperation::BinaryOp {
            op: BinaryOperator::Add,
            left: string_to_cir_value(left),
            right: string_to_cir_value(right),
            result,
        },

        HIFOpKind::Sub(left, right) => CIROperation::BinaryOp {
            op: BinaryOperator::Sub,
            left: string_to_cir_value(left),
            right: string_to_cir_value(right),
            result,
        },

        HIFOpKind::Mul(left, right) => CIROperation::BinaryOp {
            op: BinaryOperator::Mul,
            left: string_to_cir_value(left),
            right: string_to_cir_value(right),
            result,
        },

        HIFOpKind::Div(left, right) => CIROperation::BinaryOp {
            op: BinaryOperator::Div,
            left: string_to_cir_value(left),
            right: string_to_cir_value(right),
            result,
        },

        HIFOpKind::Mod(left, right) => CIROperation::BinaryOp {
            op: BinaryOperator::Mod,
            left: string_to_cir_value(left),
            right: string_to_cir_value(right),
            result,
        },

        HIFOpKind::Eq(left, right) => CIROperation::BinaryOp {
            op: BinaryOperator::Eq,
            left: string_to_cir_value(left),
            right: string_to_cir_value(right),
            result,
        },

        HIFOpKind::Ne(left, right) => CIROperation::BinaryOp {
            op: BinaryOperator::Ne,
            left: string_to_cir_value(left),
            right: string_to_cir_value(right),
            result,
        },

        HIFOpKind::Lt(left, right) => CIROperation::BinaryOp {
            op: BinaryOperator::Lt,
            left: string_to_cir_value(left),
            right: string_to_cir_value(right),
            result,
        },

        HIFOpKind::Gt(left, right) => CIROperation::BinaryOp {
            op: BinaryOperator::Gt,
            left: string_to_cir_value(left),
            right: string_to_cir_value(right),
            result,
        },

        HIFOpKind::Le(left, right) => CIROperation::BinaryOp {
            op: BinaryOperator::Le,
            left: string_to_cir_value(left),
            right: string_to_cir_value(right),
            result,
        },

        HIFOpKind::Ge(left, right) => CIROperation::BinaryOp {
            op: BinaryOperator::Ge,
            left: string_to_cir_value(left),
            right: string_to_cir_value(right),
            result,
        },

        HIFOpKind::And(left, right) => CIROperation::BinaryOp {
            op: BinaryOperator::And,
            left: string_to_cir_value(left),
            right: string_to_cir_value(right),
            result,
        },

        HIFOpKind::Or(left, right) => CIROperation::BinaryOp {
            op: BinaryOperator::Or,
            left: string_to_cir_value(left),
            right: string_to_cir_value(right),
            result,
        },

        HIFOpKind::Neg(operand) => CIROperation::UnaryOp {
            op: UnaryOperator::Neg,
            operand: string_to_cir_value(operand),
            result,
        },

        HIFOpKind::Not(operand) => CIROperation::UnaryOp {
            op: UnaryOperator::Not,
            operand: string_to_cir_value(operand),
            result,
        },

        HIFOpKind::Call { function, args } => CIROperation::Call {
            function: function.clone(),
            args: args.iter().map(|s| string_to_cir_value(s)).collect(),
            result,
        },

        HIFOpKind::Literal(value) => CIROperation::Literal {
            value: hif_value_to_cir_value(value),
            result,
        },

        HIFOpKind::Var(name) => CIROperation::Var {
            name: name.clone(),
            result,
        },

        HIFOpKind::Return(value) => CIROperation::Return {
            value: string_to_cir_value(value),
        },

        HIFOpKind::Format { template, values } => CIROperation::Format {
            template: template.clone(),
            values: values
                .iter()
                .map(|(k, v)| (k.clone(), string_to_cir_value(v)))
                .collect(),
            result,
        },

        HIFOpKind::Concat(parts) => CIROperation::Concat {
            parts: parts.iter().map(|s| string_to_cir_value(s)).collect(),
            result,
        },
    }
}

/// Convert a CIR value to a string representation.
fn cir_value_to_string(value: &CIRValue) -> String {
    match value {
        CIRValue::Ref(name) => name.clone(),
        CIRValue::Int(n) => n.to_string(),
        CIRValue::Float(f) => {
            let s = f.to_string();
            if s.contains('.') {
                s
            } else {
                format!("{}.0", s)
            }
        }
        CIRValue::String(s) => format!("\"{}\"", s.replace('"', "\\\"")),
        CIRValue::Bool(b) => b.to_string(),
        CIRValue::None => "none".to_string(),
        CIRValue::Operation(_) => "_op".to_string(),
    }
}

/// Convert a CIR value to a HIF value.
fn cir_value_to_hif_value(value: &CIRValue) -> HIFValue {
    match value {
        CIRValue::Int(n) => HIFValue::Int(*n),
        CIRValue::Float(f) => HIFValue::Float(*f),
        CIRValue::String(s) => HIFValue::String(s.clone()),
        CIRValue::Bool(b) => HIFValue::Bool(*b),
        CIRValue::None => HIFValue::None,
        CIRValue::Ref(_) => HIFValue::None, // Refs aren't literal values
        CIRValue::Operation(_) => HIFValue::None,
    }
}

/// Convert a HIF value to a CIR value.
fn hif_value_to_cir_value(value: &HIFValue) -> CIRValue {
    match value {
        HIFValue::Int(n) => CIRValue::Int(*n),
        HIFValue::Float(f) => CIRValue::Float(*f),
        HIFValue::String(s) => CIRValue::String(s.clone()),
        HIFValue::Bool(b) => CIRValue::Bool(*b),
        HIFValue::None => CIRValue::None,
    }
}

/// Convert a string to a CIR value (parsing literals or treating as ref).
fn string_to_cir_value(s: &str) -> CIRValue {
    let s = s.trim();

    if s == "none" {
        return CIRValue::None;
    }
    if s == "true" {
        return CIRValue::Bool(true);
    }
    if s == "false" {
        return CIRValue::Bool(false);
    }

    // String literal
    if s.starts_with('"') && s.ends_with('"') {
        let inner = &s[1..s.len() - 1];
        return CIRValue::String(inner.replace("\\\"", "\""));
    }

    // Float
    if s.contains('.') {
        if let Ok(f) = s.parse::<f64>() {
            return CIRValue::Float(f);
        }
    }

    // Int
    if let Ok(n) = s.parse::<i64>() {
        return CIRValue::Int(n);
    }

    // Default to variable reference
    CIRValue::Ref(s.to_string())
}

/// Compute a hash for caching purposes.
pub fn compute_context_hash(context: &str) -> String {
    use std::collections::hash_map::DefaultHasher;
    use std::hash::{Hash, Hasher};

    let mut hasher = DefaultHasher::new();
    context.hash(&mut hasher);
    format!("{:x}", hasher.finish())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_type_conversion() {
        // CIR to HIF
        assert_eq!(cir_type_to_hif(&CIRType::simple("int")), HIFType::Int);
        assert_eq!(cir_type_to_hif(&CIRType::simple("string")), HIFType::String);
        assert_eq!(
            cir_type_to_hif(&CIRType::simple("User")),
            HIFType::Struct("User".to_string())
        );

        // HIF to CIR
        let cir = hif_type_to_cir(&HIFType::Int);
        assert!(matches!(cir, CIRType::Simple(s) if s == "int"));

        let cir = hif_type_to_cir(&HIFType::Array(Box::new(HIFType::String)));
        assert!(matches!(cir, CIRType::Complex(CIRTypeKind::List { .. })));
    }

    #[test]
    fn test_function_roundtrip() {
        let func = CIRFunction::new("test_func")
            .with_param("x", "int")
            .returning(CIRType::simple("int"))
            .with_op(CIROperation::Return {
                value: CIRValue::var("x"),
            });

        let hif_intent = cir_function_to_hif_intent(&func, "test123");
        assert_eq!(hif_intent.name, "test_func");
        assert_eq!(hif_intent.params.len(), 1);

        let back = hif_intent_to_cir_function(&hif_intent);
        assert_eq!(back.name, "test_func");
        assert_eq!(back.params.len(), 1);
    }

    #[test]
    fn test_struct_roundtrip() {
        let type_def = TypeDefinition {
            name: "Point".to_string(),
            fields: vec![
                FieldDefinition {
                    name: "x".to_string(),
                    ty: "float".to_string(),
                    optional: false,
                    default: None,
                },
                FieldDefinition {
                    name: "y".to_string(),
                    ty: "float".to_string(),
                    optional: false,
                    default: None,
                },
            ],
        };

        let hif_struct = cir_type_def_to_hif_struct(&type_def, "xyz789");
        assert_eq!(hif_struct.name, "Point");
        assert_eq!(hif_struct.fields.len(), 2);

        let back = hif_struct_to_cir_type_def(&hif_struct);
        assert_eq!(back.name, "Point");
        assert_eq!(back.fields.len(), 2);
    }
}
