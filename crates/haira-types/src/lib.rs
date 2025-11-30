//! Type system and inference for the Haira programming language.
//!
//! This crate handles:
//! - Type representation
//! - Type inference via unification
//! - Type checking
//! - Constraint generation and solving

use smol_str::SmolStr;
use std::sync::atomic::{AtomicU32, Ordering};

/// Unique type variable ID.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct TypeVar(u32);

static NEXT_TYPE_VAR: AtomicU32 = AtomicU32::new(0);

impl TypeVar {
    pub fn fresh() -> Self {
        TypeVar(NEXT_TYPE_VAR.fetch_add(1, Ordering::SeqCst))
    }
}

/// Core type representation.
#[derive(Debug, Clone, PartialEq)]
pub enum Type {
    /// Unknown type (type variable for inference).
    Unknown(TypeVar),
    /// Primitive types.
    Int,
    Float,
    String,
    Bool,
    /// Named type (user-defined).
    Named(SmolStr),
    /// Generic type with parameters.
    Generic(SmolStr, Vec<Type>),
    /// Option type.
    Option(Box<Type>),
    /// Array type.
    Array(Box<Type>),
    /// Tuple type.
    Tuple(Vec<Type>),
    /// Function type.
    Function {
        params: Vec<Type>,
        returns: Box<Type>,
    },
    /// Union type.
    Union(Vec<Type>),
    /// Unit type (no value).
    Unit,
    /// Error type (for error recovery).
    Error,
}

impl Type {
    /// Check if type contains any unknown type variables.
    pub fn is_concrete(&self) -> bool {
        match self {
            Type::Unknown(_) => false,
            Type::Option(inner) | Type::Array(inner) => inner.is_concrete(),
            Type::Tuple(types) | Type::Union(types) => types.iter().all(|t| t.is_concrete()),
            Type::Generic(_, args) => args.iter().all(|t| t.is_concrete()),
            Type::Function { params, returns } => {
                params.iter().all(|t| t.is_concrete()) && returns.is_concrete()
            }
            _ => true,
        }
    }
}

/// Type inference context.
pub struct InferenceContext {
    /// Substitution map from type variables to types.
    substitutions: rustc_hash::FxHashMap<TypeVar, Type>,
}

impl InferenceContext {
    pub fn new() -> Self {
        Self {
            substitutions: rustc_hash::FxHashMap::default(),
        }
    }

    /// Unify two types, returning error if incompatible.
    pub fn unify(&mut self, a: &Type, b: &Type) -> Result<(), TypeError> {
        match (a, b) {
            (Type::Unknown(var), other) | (other, Type::Unknown(var)) => {
                if let Type::Unknown(other_var) = other {
                    if var == other_var {
                        return Ok(());
                    }
                }
                self.substitutions.insert(*var, other.clone());
                Ok(())
            }
            (Type::Int, Type::Int)
            | (Type::Float, Type::Float)
            | (Type::String, Type::String)
            | (Type::Bool, Type::Bool)
            | (Type::Unit, Type::Unit) => Ok(()),
            (Type::Named(a), Type::Named(b)) if a == b => Ok(()),
            (Type::Option(a), Type::Option(b)) => self.unify(a, b),
            (Type::Array(a), Type::Array(b)) => self.unify(a, b),
            (Type::Tuple(a), Type::Tuple(b)) if a.len() == b.len() => {
                for (ta, tb) in a.iter().zip(b.iter()) {
                    self.unify(ta, tb)?;
                }
                Ok(())
            }
            (
                Type::Function { params: pa, returns: ra },
                Type::Function { params: pb, returns: rb },
            ) if pa.len() == pb.len() => {
                for (ta, tb) in pa.iter().zip(pb.iter()) {
                    self.unify(ta, tb)?;
                }
                self.unify(ra, rb)
            }
            _ => Err(TypeError::Mismatch {
                expected: a.clone(),
                found: b.clone(),
            }),
        }
    }

    /// Apply substitutions to resolve a type.
    pub fn resolve(&self, ty: &Type) -> Type {
        match ty {
            Type::Unknown(var) => {
                if let Some(resolved) = self.substitutions.get(var) {
                    self.resolve(resolved)
                } else {
                    ty.clone()
                }
            }
            Type::Option(inner) => Type::Option(Box::new(self.resolve(inner))),
            Type::Array(inner) => Type::Array(Box::new(self.resolve(inner))),
            Type::Tuple(types) => Type::Tuple(types.iter().map(|t| self.resolve(t)).collect()),
            Type::Generic(name, args) => {
                Type::Generic(name.clone(), args.iter().map(|t| self.resolve(t)).collect())
            }
            Type::Function { params, returns } => Type::Function {
                params: params.iter().map(|t| self.resolve(t)).collect(),
                returns: Box::new(self.resolve(returns)),
            },
            Type::Union(types) => Type::Union(types.iter().map(|t| self.resolve(t)).collect()),
            _ => ty.clone(),
        }
    }
}

impl Default for InferenceContext {
    fn default() -> Self {
        Self::new()
    }
}

/// Type error.
#[derive(Debug, Clone)]
pub enum TypeError {
    Mismatch { expected: Type, found: Type },
    UnresolvedType(SmolStr),
    InfiniteType(TypeVar),
}
