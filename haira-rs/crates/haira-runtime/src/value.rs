// Dynamic Value type for the Haira runtime.
//
// Every value in a running Haira program is represented as a `Value` enum.
// Uses `Arc` for reference-counted sharing — no tracing GC needed.

use std::collections::HashMap;
use std::fmt;
use std::sync::Arc;

use serde::ser::{SerializeMap, SerializeSeq};
use serde::{Serialize, Serializer};

// ---------------------------------------------------------------------------
// RuntimeError
// ---------------------------------------------------------------------------

/// A runtime error produced by value operations.
#[derive(Clone, Debug)]
pub struct RuntimeError {
    pub message: String,
}

impl RuntimeError {
    pub fn new(msg: impl Into<String>) -> Self {
        Self {
            message: msg.into(),
        }
    }
}

impl fmt::Display for RuntimeError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(&self.message)
    }
}

impl std::error::Error for RuntimeError {}

// ---------------------------------------------------------------------------
// HairaStruct
// ---------------------------------------------------------------------------

/// A struct value — a named collection of fields.
#[derive(Clone, Debug)]
pub struct HairaStruct {
    pub type_name: Arc<str>,
    pub fields: Arc<HashMap<Arc<str>, Value>>,
}

// ---------------------------------------------------------------------------
// HairaFn
// ---------------------------------------------------------------------------

/// A function value — can be a native Rust function or a closure.
#[derive(Clone)]
pub enum HairaFn {
    /// A native Rust function (for builtins).
    Native {
        name: Arc<str>,
        func: Arc<dyn Fn(Vec<Value>) -> Result<Value, RuntimeError> + Send + Sync>,
    },
    /// A closure that captures variables.
    Closure {
        name: Arc<str>,
        params: Vec<String>,
        captures: Arc<HashMap<String, Value>>,
        /// The body is opaque at this level — the codegen fills it in.
        body_id: usize,
    },
}

impl fmt::Debug for HairaFn {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            HairaFn::Native { name, .. } => write!(f, "Native({name})"),
            HairaFn::Closure {
                name,
                params,
                body_id,
                ..
            } => write!(f, "Closure({name}, params={params:?}, body={body_id})"),
        }
    }
}

// ---------------------------------------------------------------------------
// Value
// ---------------------------------------------------------------------------

/// The core dynamic value type for the Haira runtime.
/// All values in a running Haira program are represented as `Value`.
#[derive(Clone, Debug)]
pub enum Value {
    None,
    Bool(bool),
    Int(i64),
    Float(f64),
    Str(Arc<str>),
    List(Arc<Vec<Value>>),
    Map(Arc<HashMap<Arc<str>, Value>>),
    Struct(HairaStruct),
    Fn(HairaFn),
    Error(Arc<str>),
}

// ---------------------------------------------------------------------------
// Type inspection
// ---------------------------------------------------------------------------

impl Value {
    /// Returns the type name of this value as a static string.
    pub fn type_name(&self) -> &str {
        match self {
            Value::None => "none",
            Value::Bool(_) => "bool",
            Value::Int(_) => "int",
            Value::Float(_) => "float",
            Value::Str(_) => "string",
            Value::List(_) => "list",
            Value::Map(_) => "map",
            Value::Struct(_) => "struct",
            Value::Fn(_) => "fn",
            Value::Error(_) => "error",
        }
    }

    /// Returns `true` if this value is considered truthy.
    ///
    /// Falsy values: `None`, `Bool(false)`, `Int(0)`, `Float(0.0)`, empty string,
    /// empty list, empty map. Everything else is truthy.
    pub fn is_truthy(&self) -> bool {
        match self {
            Value::None => false,
            Value::Bool(b) => *b,
            Value::Int(n) => *n != 0,
            Value::Float(n) => *n != 0.0,
            Value::Str(s) => !s.is_empty(),
            Value::List(v) => !v.is_empty(),
            Value::Map(m) => !m.is_empty(),
            Value::Error(_) => false,
            Value::Struct(_) | Value::Fn(_) => true,
        }
    }
}

// ---------------------------------------------------------------------------
// Conversions (accessors)
// ---------------------------------------------------------------------------

impl Value {
    pub fn as_bool(&self) -> Option<bool> {
        match self {
            Value::Bool(b) => Some(*b),
            _ => None,
        }
    }

    pub fn as_int(&self) -> Option<i64> {
        match self {
            Value::Int(n) => Some(*n),
            _ => None,
        }
    }

    pub fn as_float(&self) -> Option<f64> {
        match self {
            Value::Float(n) => Some(*n),
            Value::Int(n) => Some(*n as f64),
            _ => None,
        }
    }

    pub fn as_str(&self) -> Option<&str> {
        match self {
            Value::Str(s) => Some(s),
            _ => None,
        }
    }

    pub fn as_list(&self) -> Option<&[Value]> {
        match self {
            Value::List(v) => Some(v.as_slice()),
            _ => None,
        }
    }

    pub fn as_map(&self) -> Option<&HashMap<Arc<str>, Value>> {
        match self {
            Value::Map(m) => Some(m.as_ref()),
            _ => None,
        }
    }

    /// Convert any value to its string representation (consumes self).
    pub fn into_string(self) -> String {
        format!("{self}")
    }
}

// ---------------------------------------------------------------------------
// Core operations (matching Go runtime core.go semantics)
// ---------------------------------------------------------------------------

impl Value {
    /// Field / index access.
    ///
    /// - Map + Str key → look up field
    /// - List + Int key → look up index
    /// - Struct + Str key → look up field
    /// - Everything else → `Value::None`
    pub fn get(&self, key: &Value) -> Value {
        match (self, key) {
            (Value::Map(m), Value::Str(k)) => m.get(k.as_ref()).cloned().unwrap_or(Value::None),
            (Value::List(v), Value::Int(i)) => {
                let idx = *i;
                if idx >= 0 && (idx as usize) < v.len() {
                    v[idx as usize].clone()
                } else {
                    Value::None
                }
            }
            (Value::Struct(s), Value::Str(k)) => {
                s.fields.get(k.as_ref()).cloned().unwrap_or(Value::None)
            }
            (Value::Str(s), Value::Int(i)) => {
                let idx = *i;
                if idx >= 0 && (idx as usize) < s.len() {
                    let ch = s.as_bytes()[idx as usize];
                    Value::Str(Arc::from(String::from(ch as char)))
                } else {
                    Value::None
                }
            }
            _ => Value::None,
        }
    }

    /// Returns a new value with the given key/index set to `val`.
    ///
    /// This is a functional update — the original value is not mutated.
    pub fn set(&self, key: &Value, val: Value) -> Value {
        match (self, key) {
            (Value::Map(m), Value::Str(k)) => {
                let mut new_map = (**m).clone();
                new_map.insert(k.clone(), val);
                Value::Map(Arc::new(new_map))
            }
            (Value::List(v), Value::Int(i)) => {
                let idx = *i;
                if idx >= 0 && (idx as usize) < v.len() {
                    let mut new_vec = (**v).clone();
                    new_vec[idx as usize] = val;
                    Value::List(Arc::new(new_vec))
                } else {
                    // Out of bounds — return self unchanged.
                    self.clone()
                }
            }
            (Value::Struct(s), Value::Str(k)) => {
                let mut new_fields = (*s.fields).clone();
                new_fields.insert(k.clone(), val);
                Value::Struct(HairaStruct {
                    type_name: s.type_name.clone(),
                    fields: Arc::new(new_fields),
                })
            }
            _ => self.clone(),
        }
    }

    /// Returns `Int(length)` for strings, lists, and maps. `Int(0)` for everything else.
    pub fn len(&self) -> Value {
        match self {
            Value::Str(s) => Value::Int(s.len() as i64),
            Value::List(v) => Value::Int(v.len() as i64),
            Value::Map(m) => Value::Int(m.len() as i64),
            Value::None => Value::Int(0),
            _ => Value::Int(0),
        }
    }

    /// Convert to a list. Lists pass through, maps become a list of `[key, value]` pairs,
    /// strings become a list of single-character strings.
    pub fn to_slice(&self) -> Value {
        match self {
            Value::List(_) => self.clone(),
            Value::Map(m) => {
                let items: Vec<Value> = m
                    .iter()
                    .map(|(k, v)| {
                        Value::List(Arc::new(vec![Value::Str(k.clone()), v.clone()]))
                    })
                    .collect();
                Value::List(Arc::new(items))
            }
            Value::Str(s) => {
                let items: Vec<Value> = s
                    .chars()
                    .map(|c| Value::Str(Arc::from(c.to_string())))
                    .collect();
                Value::List(Arc::new(items))
            }
            _ => Value::List(Arc::new(Vec::new())),
        }
    }

    /// Convert to a map. Maps pass through, structs become a map of their fields,
    /// lists of `[key, value]` pairs become a map.
    pub fn to_map(&self) -> Value {
        match self {
            Value::Map(_) => self.clone(),
            Value::Struct(s) => Value::Map(s.fields.clone()),
            Value::List(items) => {
                let mut map = HashMap::new();
                for item in items.iter() {
                    if let Value::List(pair) = item {
                        if pair.len() == 2 {
                            let key: Arc<str> = Arc::from(format!("{}", pair[0]));
                            map.insert(key, pair[1].clone());
                        }
                    }
                }
                Value::Map(Arc::new(map))
            }
            _ => Value::Map(Arc::new(HashMap::new())),
        }
    }

    /// Membership test.
    ///
    /// - List: checks if any element equals `item`
    /// - Map: checks if key exists (item must be Str)
    /// - Str: checks if substring exists (item must be Str)
    pub fn contains(&self, item: &Value) -> Value {
        match self {
            Value::List(v) => Value::Bool(v.iter().any(|el| el == item)),
            Value::Map(m) => {
                if let Value::Str(k) = item {
                    Value::Bool(m.contains_key(k.as_ref()))
                } else {
                    Value::Bool(false)
                }
            }
            Value::Str(s) => {
                if let Value::Str(sub) = item {
                    Value::Bool(s.contains(sub.as_ref()))
                } else {
                    Value::Bool(false)
                }
            }
            _ => Value::Bool(false),
        }
    }

    /// Returns the keys of a map as a list of strings.
    pub fn keys(&self) -> Value {
        match self {
            Value::Map(m) => {
                let keys: Vec<Value> = m.keys().map(|k| Value::Str(k.clone())).collect();
                Value::List(Arc::new(keys))
            }
            Value::Struct(s) => {
                let keys: Vec<Value> = s.fields.keys().map(|k| Value::Str(k.clone())).collect();
                Value::List(Arc::new(keys))
            }
            _ => Value::List(Arc::new(Vec::new())),
        }
    }
}

// ---------------------------------------------------------------------------
// Arithmetic
// ---------------------------------------------------------------------------

/// Helper: extract numeric pair, promoting Int to Float when mixed.
fn numeric_pair(a: &Value, b: &Value) -> Option<NumPair> {
    match (a, b) {
        (Value::Int(x), Value::Int(y)) => Some(NumPair::Ints(*x, *y)),
        (Value::Float(x), Value::Float(y)) => Some(NumPair::Floats(*x, *y)),
        (Value::Int(x), Value::Float(y)) => Some(NumPair::Floats(*x as f64, *y)),
        (Value::Float(x), Value::Int(y)) => Some(NumPair::Floats(*x, *y as f64)),
        _ => None,
    }
}

enum NumPair {
    Ints(i64, i64),
    Floats(f64, f64),
}

impl Value {
    /// Addition. Supports numeric addition, string concatenation, list concatenation,
    /// and `Str + anything` (auto-converts RHS to string).
    pub fn add(&self, other: &Value) -> Result<Value, RuntimeError> {
        // String concatenation: Str + Str
        if let (Value::Str(a), Value::Str(b)) = (self, other) {
            let mut s = String::with_capacity(a.len() + b.len());
            s.push_str(a);
            s.push_str(b);
            return Ok(Value::Str(Arc::from(s)));
        }

        // Str + anything → string concat
        if let Value::Str(a) = self {
            let mut s = String::with_capacity(a.len() + 16);
            s.push_str(a);
            s.push_str(&format!("{other}"));
            return Ok(Value::Str(Arc::from(s)));
        }

        // anything + Str → string concat
        if let Value::Str(_) = other {
            let mut s = format!("{self}");
            s.push_str(&format!("{other}"));
            return Ok(Value::Str(Arc::from(s)));
        }

        // List + List → concatenation
        if let (Value::List(a), Value::List(b)) = (self, other) {
            let mut new_vec = Vec::with_capacity(a.len() + b.len());
            new_vec.extend_from_slice(a);
            new_vec.extend_from_slice(b);
            return Ok(Value::List(Arc::new(new_vec)));
        }

        // Numeric addition
        if let Some(pair) = numeric_pair(self, other) {
            return Ok(match pair {
                NumPair::Ints(a, b) => Value::Int(a.wrapping_add(b)),
                NumPair::Floats(a, b) => Value::Float(a + b),
            });
        }

        Err(RuntimeError::new(format!(
            "cannot add {} and {}",
            self.type_name(),
            other.type_name()
        )))
    }

    pub fn sub(&self, other: &Value) -> Result<Value, RuntimeError> {
        if let Some(pair) = numeric_pair(self, other) {
            return Ok(match pair {
                NumPair::Ints(a, b) => Value::Int(a.wrapping_sub(b)),
                NumPair::Floats(a, b) => Value::Float(a - b),
            });
        }
        Err(RuntimeError::new(format!(
            "cannot subtract {} from {}",
            other.type_name(),
            self.type_name()
        )))
    }

    pub fn mul(&self, other: &Value) -> Result<Value, RuntimeError> {
        if let Some(pair) = numeric_pair(self, other) {
            return Ok(match pair {
                NumPair::Ints(a, b) => Value::Int(a.wrapping_mul(b)),
                NumPair::Floats(a, b) => Value::Float(a * b),
            });
        }
        Err(RuntimeError::new(format!(
            "cannot multiply {} and {}",
            self.type_name(),
            other.type_name()
        )))
    }

    pub fn div(&self, other: &Value) -> Result<Value, RuntimeError> {
        if let Some(pair) = numeric_pair(self, other) {
            return match pair {
                NumPair::Ints(a, b) => {
                    if b == 0 {
                        Err(RuntimeError::new("division by zero"))
                    } else {
                        Ok(Value::Int(a / b))
                    }
                }
                NumPair::Floats(a, b) => {
                    if b == 0.0 {
                        Err(RuntimeError::new("division by zero"))
                    } else {
                        Ok(Value::Float(a / b))
                    }
                }
            };
        }
        Err(RuntimeError::new(format!(
            "cannot divide {} by {}",
            self.type_name(),
            other.type_name()
        )))
    }

    pub fn rem(&self, other: &Value) -> Result<Value, RuntimeError> {
        if let Some(pair) = numeric_pair(self, other) {
            return match pair {
                NumPair::Ints(a, b) => {
                    if b == 0 {
                        Err(RuntimeError::new("remainder by zero"))
                    } else {
                        Ok(Value::Int(a % b))
                    }
                }
                NumPair::Floats(a, b) => {
                    if b == 0.0 {
                        Err(RuntimeError::new("remainder by zero"))
                    } else {
                        Ok(Value::Float(a % b))
                    }
                }
            };
        }
        Err(RuntimeError::new(format!(
            "cannot compute remainder of {} and {}",
            self.type_name(),
            other.type_name()
        )))
    }

    pub fn neg(&self) -> Result<Value, RuntimeError> {
        match self {
            Value::Int(n) => Ok(Value::Int(-n)),
            Value::Float(n) => Ok(Value::Float(-n)),
            _ => Err(RuntimeError::new(format!(
                "cannot negate {}",
                self.type_name()
            ))),
        }
    }
}

// ---------------------------------------------------------------------------
// Comparison
// ---------------------------------------------------------------------------

impl Value {
    /// Deep equality — returns `Bool`.
    pub fn eq(&self, other: &Value) -> Value {
        Value::Bool(self == other)
    }

    pub fn ne(&self, other: &Value) -> Value {
        Value::Bool(self != other)
    }

    pub fn lt(&self, other: &Value) -> Value {
        Value::Bool(self.partial_cmp_impl(other) == Some(std::cmp::Ordering::Less))
    }

    pub fn gt(&self, other: &Value) -> Value {
        Value::Bool(self.partial_cmp_impl(other) == Some(std::cmp::Ordering::Greater))
    }

    pub fn le(&self, other: &Value) -> Value {
        Value::Bool(matches!(
            self.partial_cmp_impl(other),
            Some(std::cmp::Ordering::Less | std::cmp::Ordering::Equal)
        ))
    }

    pub fn ge(&self, other: &Value) -> Value {
        Value::Bool(matches!(
            self.partial_cmp_impl(other),
            Some(std::cmp::Ordering::Greater | std::cmp::Ordering::Equal)
        ))
    }

    fn partial_cmp_impl(&self, other: &Value) -> Option<std::cmp::Ordering> {
        match (self, other) {
            (Value::Int(a), Value::Int(b)) => a.partial_cmp(b),
            (Value::Float(a), Value::Float(b)) => a.partial_cmp(b),
            (Value::Int(a), Value::Float(b)) => (*a as f64).partial_cmp(b),
            (Value::Float(a), Value::Int(b)) => a.partial_cmp(&(*b as f64)),
            (Value::Str(a), Value::Str(b)) => a.as_ref().partial_cmp(b.as_ref()),
            (Value::Bool(a), Value::Bool(b)) => a.partial_cmp(b),
            _ => None,
        }
    }
}

// ---------------------------------------------------------------------------
// Logical
// ---------------------------------------------------------------------------

impl Value {
    /// Short-circuit AND — returns `other` if self is truthy, else self.
    pub fn and(&self, other: &Value) -> Value {
        if self.is_truthy() {
            other.clone()
        } else {
            self.clone()
        }
    }

    /// Short-circuit OR — returns self if truthy, else `other`.
    pub fn or(&self, other: &Value) -> Value {
        if self.is_truthy() {
            self.clone()
        } else {
            other.clone()
        }
    }

    /// Logical NOT.
    pub fn not(&self) -> Value {
        Value::Bool(!self.is_truthy())
    }
}

// ---------------------------------------------------------------------------
// Bitwise
// ---------------------------------------------------------------------------

impl Value {
    pub fn bit_and(&self, other: &Value) -> Result<Value, RuntimeError> {
        match (self, other) {
            (Value::Int(a), Value::Int(b)) => Ok(Value::Int(a & b)),
            _ => Err(RuntimeError::new(format!(
                "bitwise AND requires int operands, got {} and {}",
                self.type_name(),
                other.type_name()
            ))),
        }
    }

    pub fn bit_or(&self, other: &Value) -> Result<Value, RuntimeError> {
        match (self, other) {
            (Value::Int(a), Value::Int(b)) => Ok(Value::Int(a | b)),
            _ => Err(RuntimeError::new(format!(
                "bitwise OR requires int operands, got {} and {}",
                self.type_name(),
                other.type_name()
            ))),
        }
    }

    pub fn bit_xor(&self, other: &Value) -> Result<Value, RuntimeError> {
        match (self, other) {
            (Value::Int(a), Value::Int(b)) => Ok(Value::Int(a ^ b)),
            _ => Err(RuntimeError::new(format!(
                "bitwise XOR requires int operands, got {} and {}",
                self.type_name(),
                other.type_name()
            ))),
        }
    }

    pub fn bit_not(&self) -> Result<Value, RuntimeError> {
        match self {
            Value::Int(n) => Ok(Value::Int(!n)),
            _ => Err(RuntimeError::new(format!(
                "bitwise NOT requires int operand, got {}",
                self.type_name()
            ))),
        }
    }

    pub fn shl(&self, other: &Value) -> Result<Value, RuntimeError> {
        match (self, other) {
            (Value::Int(a), Value::Int(b)) => {
                if *b < 0 || *b > 63 {
                    Err(RuntimeError::new(format!(
                        "shift amount {b} out of range (0..63)"
                    )))
                } else {
                    Ok(Value::Int(a << b))
                }
            }
            _ => Err(RuntimeError::new(format!(
                "left shift requires int operands, got {} and {}",
                self.type_name(),
                other.type_name()
            ))),
        }
    }

    pub fn shr(&self, other: &Value) -> Result<Value, RuntimeError> {
        match (self, other) {
            (Value::Int(a), Value::Int(b)) => {
                if *b < 0 || *b > 63 {
                    Err(RuntimeError::new(format!(
                        "shift amount {b} out of range (0..63)"
                    )))
                } else {
                    Ok(Value::Int(a >> b))
                }
            }
            _ => Err(RuntimeError::new(format!(
                "right shift requires int operands, got {} and {}",
                self.type_name(),
                other.type_name()
            ))),
        }
    }
}

// ---------------------------------------------------------------------------
// Display
// ---------------------------------------------------------------------------

impl fmt::Display for Value {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Value::None => f.write_str("none"),
            Value::Bool(b) => write!(f, "{b}"),
            Value::Int(n) => write!(f, "{n}"),
            Value::Float(n) => {
                // Ensure at least one decimal place so it looks like a float.
                if n.fract() == 0.0 && !n.is_nan() && !n.is_infinite() {
                    write!(f, "{n:.1}")
                } else {
                    write!(f, "{n}")
                }
            }
            Value::Str(s) => f.write_str(s),
            Value::List(items) => {
                f.write_str("[")?;
                for (i, item) in items.iter().enumerate() {
                    if i > 0 {
                        f.write_str(", ")?;
                    }
                    // Strings inside collections are displayed with quotes for clarity.
                    display_repr(item, f)?;
                }
                f.write_str("]")
            }
            Value::Map(m) => {
                f.write_str("{")?;
                let mut entries: Vec<_> = m.iter().collect();
                entries.sort_by_key(|(k, _)| (*k).clone());
                for (i, (k, v)) in entries.iter().enumerate() {
                    if i > 0 {
                        f.write_str(", ")?;
                    }
                    write!(f, "{k}: ")?;
                    display_repr(v, f)?;
                }
                f.write_str("}")
            }
            Value::Struct(s) => {
                write!(f, "{}", s.type_name)?;
                f.write_str("{")?;
                let mut fields: Vec<_> = s.fields.iter().collect();
                fields.sort_by_key(|(k, _)| (*k).clone());
                for (i, (k, v)) in fields.iter().enumerate() {
                    if i > 0 {
                        f.write_str(", ")?;
                    }
                    write!(f, "{k}: ")?;
                    display_repr(v, f)?;
                }
                f.write_str("}")
            }
            Value::Fn(func) => match func {
                HairaFn::Native { name, .. } => write!(f, "<fn {name}>"),
                HairaFn::Closure { name, .. } => write!(f, "<fn {name}>"),
            },
            Value::Error(msg) => write!(f, "error: {msg}"),
        }
    }
}

/// Helper: display a value as its repr (strings get quotes).
fn display_repr(val: &Value, f: &mut fmt::Formatter<'_>) -> fmt::Result {
    match val {
        Value::Str(s) => write!(f, "\"{s}\""),
        other => write!(f, "{other}"),
    }
}

// ---------------------------------------------------------------------------
// PartialEq — deep structural equality
// ---------------------------------------------------------------------------

impl PartialEq for Value {
    fn eq(&self, other: &Self) -> bool {
        match (self, other) {
            (Value::None, Value::None) => true,
            (Value::Bool(a), Value::Bool(b)) => a == b,
            (Value::Int(a), Value::Int(b)) => a == b,
            (Value::Float(a), Value::Float(b)) => a == b,
            // Cross-type numeric comparison
            (Value::Int(a), Value::Float(b)) => (*a as f64) == *b,
            (Value::Float(a), Value::Int(b)) => *a == (*b as f64),
            (Value::Str(a), Value::Str(b)) => a == b,
            (Value::List(a), Value::List(b)) => a == b,
            (Value::Map(a), Value::Map(b)) => a == b,
            (Value::Struct(a), Value::Struct(b)) => {
                a.type_name == b.type_name && a.fields == b.fields
            }
            (Value::Error(a), Value::Error(b)) => a == b,
            _ => false,
        }
    }
}

// ---------------------------------------------------------------------------
// Serde Serialize
// ---------------------------------------------------------------------------

impl Serialize for Value {
    fn serialize<S: Serializer>(&self, serializer: S) -> Result<S::Ok, S::Error> {
        match self {
            Value::None => serializer.serialize_none(),
            Value::Bool(b) => serializer.serialize_bool(*b),
            Value::Int(n) => serializer.serialize_i64(*n),
            Value::Float(n) => serializer.serialize_f64(*n),
            Value::Str(s) => serializer.serialize_str(s),
            Value::List(items) => {
                let mut seq = serializer.serialize_seq(Some(items.len()))?;
                for item in items.iter() {
                    seq.serialize_element(item)?;
                }
                seq.end()
            }
            Value::Map(m) => {
                let mut map = serializer.serialize_map(Some(m.len()))?;
                for (k, v) in m.iter() {
                    map.serialize_entry(k.as_ref(), v)?;
                }
                map.end()
            }
            Value::Struct(s) => {
                let mut map = serializer.serialize_map(Some(s.fields.len()))?;
                for (k, v) in s.fields.iter() {
                    map.serialize_entry(k.as_ref(), v)?;
                }
                map.end()
            }
            Value::Fn(_) => serializer.serialize_str("<fn>"),
            Value::Error(msg) => {
                let mut map = serializer.serialize_map(Some(1))?;
                map.serialize_entry("error", msg.as_ref())?;
                map.end()
            }
        }
    }
}

// ---------------------------------------------------------------------------
// From conversions
// ---------------------------------------------------------------------------

impl From<bool> for Value {
    fn from(b: bool) -> Self {
        Value::Bool(b)
    }
}

impl From<i64> for Value {
    fn from(n: i64) -> Self {
        Value::Int(n)
    }
}

impl From<i32> for Value {
    fn from(n: i32) -> Self {
        Value::Int(n as i64)
    }
}

impl From<f64> for Value {
    fn from(n: f64) -> Self {
        Value::Float(n)
    }
}

impl From<&str> for Value {
    fn from(s: &str) -> Self {
        Value::Str(Arc::from(s))
    }
}

impl From<String> for Value {
    fn from(s: String) -> Self {
        Value::Str(Arc::from(s.as_str()))
    }
}

impl From<Vec<Value>> for Value {
    fn from(v: Vec<Value>) -> Self {
        Value::List(Arc::new(v))
    }
}

impl From<HashMap<Arc<str>, Value>> for Value {
    fn from(m: HashMap<Arc<str>, Value>) -> Self {
        Value::Map(Arc::new(m))
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    // -- type_name ----------------------------------------------------------

    #[test]
    fn test_type_name_all_variants() {
        assert_eq!(Value::None.type_name(), "none");
        assert_eq!(Value::Bool(true).type_name(), "bool");
        assert_eq!(Value::Int(42).type_name(), "int");
        assert_eq!(Value::Float(3.14).type_name(), "float");
        assert_eq!(Value::from("hello").type_name(), "string");
        assert_eq!(Value::from(vec![]).type_name(), "list");
        assert_eq!(Value::from(HashMap::new()).type_name(), "map");
        assert_eq!(
            Value::Struct(HairaStruct {
                type_name: Arc::from("User"),
                fields: Arc::new(HashMap::new()),
            })
            .type_name(),
            "struct"
        );
        assert_eq!(
            Value::Fn(HairaFn::Native {
                name: Arc::from("test"),
                func: Arc::new(|_| Ok(Value::None)),
            })
            .type_name(),
            "fn"
        );
        assert_eq!(Value::Error(Arc::from("oops")).type_name(), "error");
    }

    // -- is_truthy ----------------------------------------------------------

    #[test]
    fn test_is_truthy() {
        // Falsy values
        assert!(!Value::None.is_truthy());
        assert!(!Value::Bool(false).is_truthy());
        assert!(!Value::Int(0).is_truthy());
        assert!(!Value::Float(0.0).is_truthy());
        assert!(!Value::from("").is_truthy());
        assert!(!Value::from(vec![]).is_truthy());
        assert!(!Value::from(HashMap::new()).is_truthy());
        assert!(!Value::Error(Arc::from("err")).is_truthy());

        // Truthy values
        assert!(Value::Bool(true).is_truthy());
        assert!(Value::Int(1).is_truthy());
        assert!(Value::Int(-1).is_truthy());
        assert!(Value::Float(0.1).is_truthy());
        assert!(Value::from("hello").is_truthy());
        assert!(Value::from(vec![Value::Int(1)]).is_truthy());
    }

    // -- Display ------------------------------------------------------------

    #[test]
    fn test_display_none() {
        assert_eq!(format!("{}", Value::None), "none");
    }

    #[test]
    fn test_display_bool() {
        assert_eq!(format!("{}", Value::Bool(true)), "true");
        assert_eq!(format!("{}", Value::Bool(false)), "false");
    }

    #[test]
    fn test_display_int() {
        assert_eq!(format!("{}", Value::Int(42)), "42");
        assert_eq!(format!("{}", Value::Int(-7)), "-7");
    }

    #[test]
    fn test_display_float() {
        assert_eq!(format!("{}", Value::Float(3.14)), "3.14");
        // Whole-number floats get ".0"
        assert_eq!(format!("{}", Value::Float(5.0)), "5.0");
    }

    #[test]
    fn test_display_string() {
        assert_eq!(format!("{}", Value::from("hello world")), "hello world");
    }

    #[test]
    fn test_display_list() {
        let list = Value::from(vec![Value::Int(1), Value::from("two"), Value::Bool(true)]);
        assert_eq!(format!("{list}"), "[1, \"two\", true]");
    }

    #[test]
    fn test_display_map() {
        let mut m = HashMap::new();
        m.insert(Arc::from("name"), Value::from("Alice"));
        let map = Value::from(m);
        assert_eq!(format!("{map}"), "{name: \"Alice\"}");
    }

    #[test]
    fn test_display_struct() {
        let mut fields = HashMap::new();
        fields.insert(Arc::from("age"), Value::Int(30));
        let s = Value::Struct(HairaStruct {
            type_name: Arc::from("User"),
            fields: Arc::new(fields),
        });
        assert_eq!(format!("{s}"), "User{age: 30}");
    }

    #[test]
    fn test_display_error() {
        assert_eq!(
            format!("{}", Value::Error(Arc::from("something failed"))),
            "error: something failed"
        );
    }

    // -- Arithmetic ---------------------------------------------------------

    #[test]
    fn test_add_int_int() {
        let result = Value::Int(3).add(&Value::Int(4)).unwrap();
        assert_eq!(result, Value::Int(7));
    }

    #[test]
    fn test_add_float_float() {
        let result = Value::Float(1.5).add(&Value::Float(2.5)).unwrap();
        assert_eq!(result, Value::Float(4.0));
    }

    #[test]
    fn test_add_int_float_promotion() {
        let result = Value::Int(2).add(&Value::Float(3.5)).unwrap();
        assert_eq!(result, Value::Float(5.5));
    }

    #[test]
    fn test_add_string_concat() {
        let result = Value::from("hello ").add(&Value::from("world")).unwrap();
        assert_eq!(result, Value::from("hello world"));
    }

    #[test]
    fn test_add_string_auto_convert() {
        let result = Value::from("count: ").add(&Value::Int(42)).unwrap();
        assert_eq!(result, Value::from("count: 42"));
    }

    #[test]
    fn test_add_list_concat() {
        let a = Value::from(vec![Value::Int(1)]);
        let b = Value::from(vec![Value::Int(2)]);
        let result = a.add(&b).unwrap();
        assert_eq!(result, Value::from(vec![Value::Int(1), Value::Int(2)]));
    }

    #[test]
    fn test_sub() {
        assert_eq!(Value::Int(10).sub(&Value::Int(3)).unwrap(), Value::Int(7));
        assert_eq!(
            Value::Float(5.5).sub(&Value::Float(1.5)).unwrap(),
            Value::Float(4.0)
        );
    }

    #[test]
    fn test_mul() {
        assert_eq!(Value::Int(3).mul(&Value::Int(4)).unwrap(), Value::Int(12));
        assert_eq!(
            Value::Float(2.0).mul(&Value::Float(3.5)).unwrap(),
            Value::Float(7.0)
        );
    }

    #[test]
    fn test_div() {
        assert_eq!(Value::Int(10).div(&Value::Int(3)).unwrap(), Value::Int(3));
        assert_eq!(
            Value::Float(7.0).div(&Value::Float(2.0)).unwrap(),
            Value::Float(3.5)
        );
    }

    #[test]
    fn test_div_by_zero() {
        let result = Value::Int(1).div(&Value::Int(0));
        assert!(result.is_err());
        assert!(result.unwrap_err().message.contains("division by zero"));

        let result = Value::Float(1.0).div(&Value::Float(0.0));
        assert!(result.is_err());
    }

    #[test]
    fn test_rem() {
        assert_eq!(Value::Int(10).rem(&Value::Int(3)).unwrap(), Value::Int(1));
    }

    #[test]
    fn test_neg() {
        assert_eq!(Value::Int(5).neg().unwrap(), Value::Int(-5));
        assert_eq!(Value::Float(3.14).neg().unwrap(), Value::Float(-3.14));
        assert!(Value::from("nope").neg().is_err());
    }

    // -- Comparison ---------------------------------------------------------

    #[test]
    fn test_comparison_operators() {
        let three = Value::Int(3);
        let five = Value::Int(5);

        assert_eq!(three.eq(&three), Value::Bool(true));
        assert_eq!(three.ne(&five), Value::Bool(true));
        assert_eq!(three.lt(&five), Value::Bool(true));
        assert_eq!(five.gt(&three), Value::Bool(true));
        assert_eq!(three.le(&three), Value::Bool(true));
        assert_eq!(five.ge(&three), Value::Bool(true));
        assert_eq!(five.lt(&three), Value::Bool(false));
    }

    #[test]
    fn test_comparison_cross_type_numeric() {
        assert_eq!(Value::Int(3).eq(&Value::Float(3.0)), Value::Bool(true));
        assert_eq!(Value::Float(2.5).lt(&Value::Int(3)), Value::Bool(true));
    }

    #[test]
    fn test_comparison_strings() {
        let a = Value::from("apple");
        let b = Value::from("banana");
        assert_eq!(a.lt(&b), Value::Bool(true));
    }

    // -- Logical ------------------------------------------------------------

    #[test]
    fn test_logical_and() {
        // truthy AND other → other
        let result = Value::Int(1).and(&Value::from("yes"));
        assert_eq!(result, Value::from("yes"));

        // falsy AND other → self
        let result = Value::None.and(&Value::from("yes"));
        assert_eq!(result, Value::None);
    }

    #[test]
    fn test_logical_or() {
        // truthy OR other → self
        let result = Value::Int(1).or(&Value::from("fallback"));
        assert_eq!(result, Value::Int(1));

        // falsy OR other → other
        let result = Value::None.or(&Value::from("fallback"));
        assert_eq!(result, Value::from("fallback"));
    }

    #[test]
    fn test_logical_not() {
        assert_eq!(Value::Bool(true).not(), Value::Bool(false));
        assert_eq!(Value::None.not(), Value::Bool(true));
        assert_eq!(Value::Int(0).not(), Value::Bool(true));
        assert_eq!(Value::Int(1).not(), Value::Bool(false));
    }

    // -- Bitwise ------------------------------------------------------------

    #[test]
    fn test_bitwise_ops() {
        assert_eq!(
            Value::Int(0b1100).bit_and(&Value::Int(0b1010)).unwrap(),
            Value::Int(0b1000)
        );
        assert_eq!(
            Value::Int(0b1100).bit_or(&Value::Int(0b1010)).unwrap(),
            Value::Int(0b1110)
        );
        assert_eq!(
            Value::Int(0b1100).bit_xor(&Value::Int(0b1010)).unwrap(),
            Value::Int(0b0110)
        );
        assert_eq!(Value::Int(1).shl(&Value::Int(3)).unwrap(), Value::Int(8));
        assert_eq!(Value::Int(8).shr(&Value::Int(2)).unwrap(), Value::Int(2));
    }

    #[test]
    fn test_bitwise_not() {
        // !0 == -1 in two's complement
        assert_eq!(Value::Int(0).bit_not().unwrap(), Value::Int(-1));
    }

    #[test]
    fn test_bitwise_type_error() {
        assert!(Value::from("nope").bit_and(&Value::Int(1)).is_err());
    }

    // -- List operations ----------------------------------------------------

    #[test]
    fn test_list_get() {
        let list = Value::from(vec![Value::from("a"), Value::from("b"), Value::from("c")]);
        assert_eq!(list.get(&Value::Int(0)), Value::from("a"));
        assert_eq!(list.get(&Value::Int(2)), Value::from("c"));
        assert_eq!(list.get(&Value::Int(5)), Value::None);
        assert_eq!(list.get(&Value::Int(-1)), Value::None);
    }

    #[test]
    fn test_list_set() {
        let list = Value::from(vec![Value::Int(1), Value::Int(2), Value::Int(3)]);
        let updated = list.set(&Value::Int(1), Value::Int(42));
        assert_eq!(
            updated,
            Value::from(vec![Value::Int(1), Value::Int(42), Value::Int(3)])
        );
    }

    #[test]
    fn test_list_len() {
        let list = Value::from(vec![Value::Int(1), Value::Int(2)]);
        assert_eq!(list.len(), Value::Int(2));
    }

    #[test]
    fn test_list_contains() {
        let list = Value::from(vec![Value::Int(1), Value::Int(2), Value::Int(3)]);
        assert_eq!(list.contains(&Value::Int(2)), Value::Bool(true));
        assert_eq!(list.contains(&Value::Int(5)), Value::Bool(false));
    }

    // -- Map operations -----------------------------------------------------

    #[test]
    fn test_map_get() {
        let mut m = HashMap::new();
        m.insert(Arc::from("name"), Value::from("Alice"));
        let map = Value::from(m);
        assert_eq!(map.get(&Value::from("name")), Value::from("Alice"));
        assert_eq!(map.get(&Value::from("missing")), Value::None);
    }

    #[test]
    fn test_map_set() {
        let map = Value::from(HashMap::new());
        let updated = map.set(&Value::from("key"), Value::Int(42));
        assert_eq!(updated.get(&Value::from("key")), Value::Int(42));
    }

    #[test]
    fn test_map_len() {
        let mut m = HashMap::new();
        m.insert(Arc::from("a"), Value::Int(1));
        m.insert(Arc::from("b"), Value::Int(2));
        assert_eq!(Value::from(m).len(), Value::Int(2));
    }

    #[test]
    fn test_map_keys() {
        let mut m = HashMap::new();
        m.insert(Arc::from("x"), Value::Int(1));
        let map = Value::from(m);
        let keys = map.keys();
        if let Value::List(k) = keys {
            assert_eq!(k.len(), 1);
            assert_eq!(k[0], Value::from("x"));
        } else {
            panic!("keys should return a list");
        }
    }

    // -- Struct field access ------------------------------------------------

    #[test]
    fn test_struct_field_access() {
        let mut fields = HashMap::new();
        fields.insert(Arc::from("name"), Value::from("Bob"));
        fields.insert(Arc::from("age"), Value::Int(25));
        let s = Value::Struct(HairaStruct {
            type_name: Arc::from("Person"),
            fields: Arc::new(fields),
        });

        assert_eq!(s.get(&Value::from("name")), Value::from("Bob"));
        assert_eq!(s.get(&Value::from("age")), Value::Int(25));
        assert_eq!(s.get(&Value::from("missing")), Value::None);
    }

    // -- PartialEq ----------------------------------------------------------

    #[test]
    fn test_partial_eq_same_type() {
        assert_eq!(Value::None, Value::None);
        assert_eq!(Value::Bool(true), Value::Bool(true));
        assert_ne!(Value::Bool(true), Value::Bool(false));
        assert_eq!(Value::Int(42), Value::Int(42));
        assert_ne!(Value::Int(1), Value::Int(2));
        assert_eq!(Value::Float(3.14), Value::Float(3.14));
        assert_eq!(Value::from("hi"), Value::from("hi"));
        assert_ne!(Value::from("hi"), Value::from("bye"));
    }

    #[test]
    fn test_partial_eq_cross_type_numeric() {
        assert_eq!(Value::Int(5), Value::Float(5.0));
        assert_eq!(Value::Float(3.0), Value::Int(3));
        assert_ne!(Value::Int(5), Value::Float(5.1));
    }

    #[test]
    fn test_partial_eq_lists() {
        let a = Value::from(vec![Value::Int(1), Value::Int(2)]);
        let b = Value::from(vec![Value::Int(1), Value::Int(2)]);
        let c = Value::from(vec![Value::Int(1), Value::Int(3)]);
        assert_eq!(a, b);
        assert_ne!(a, c);
    }

    #[test]
    fn test_partial_eq_maps() {
        let mut m1 = HashMap::new();
        m1.insert(Arc::from("k"), Value::Int(1));
        let mut m2 = HashMap::new();
        m2.insert(Arc::from("k"), Value::Int(1));
        assert_eq!(Value::from(m1), Value::from(m2));
    }

    #[test]
    fn test_partial_eq_different_types() {
        assert_ne!(Value::Int(1), Value::from("1"));
        assert_ne!(Value::Bool(true), Value::Int(1));
    }

    // -- From conversions ---------------------------------------------------

    #[test]
    fn test_from_conversions() {
        let _ = Value::from(true);
        let _ = Value::from(42_i64);
        let _ = Value::from(3.14_f64);
        let _ = Value::from("hello");
        let _ = Value::from(String::from("world"));
        let _ = Value::from(vec![Value::Int(1)]);
        let _ = Value::from(HashMap::<Arc<str>, Value>::new());
    }

    // -- Serde serialization ------------------------------------------------

    #[test]
    fn test_serialize_none() {
        let json = serde_json::to_string(&Value::None).unwrap();
        assert_eq!(json, "null");
    }

    #[test]
    fn test_serialize_bool() {
        assert_eq!(serde_json::to_string(&Value::Bool(true)).unwrap(), "true");
    }

    #[test]
    fn test_serialize_int() {
        assert_eq!(serde_json::to_string(&Value::Int(42)).unwrap(), "42");
    }

    #[test]
    fn test_serialize_float() {
        assert_eq!(serde_json::to_string(&Value::Float(3.14)).unwrap(), "3.14");
    }

    #[test]
    fn test_serialize_string() {
        assert_eq!(
            serde_json::to_string(&Value::from("hello")).unwrap(),
            "\"hello\""
        );
    }

    #[test]
    fn test_serialize_list() {
        let list = Value::from(vec![Value::Int(1), Value::Bool(true)]);
        assert_eq!(serde_json::to_string(&list).unwrap(), "[1,true]");
    }

    #[test]
    fn test_serialize_map() {
        let mut m = HashMap::new();
        m.insert(Arc::from("x"), Value::Int(10));
        let json = serde_json::to_string(&Value::from(m)).unwrap();
        assert_eq!(json, "{\"x\":10}");
    }

    #[test]
    fn test_serialize_struct() {
        let mut fields = HashMap::new();
        fields.insert(Arc::from("age"), Value::Int(30));
        let s = Value::Struct(HairaStruct {
            type_name: Arc::from("User"),
            fields: Arc::new(fields),
        });
        let json = serde_json::to_string(&s).unwrap();
        assert_eq!(json, "{\"age\":30}");
    }

    // -- to_slice / to_map / contains  --------------------------------------

    #[test]
    fn test_to_slice_list_passthrough() {
        let list = Value::from(vec![Value::Int(1)]);
        assert_eq!(list.to_slice(), list);
    }

    #[test]
    fn test_to_slice_string() {
        let s = Value::from("ab");
        let result = s.to_slice();
        assert_eq!(
            result,
            Value::from(vec![Value::from("a"), Value::from("b")])
        );
    }

    #[test]
    fn test_to_map_struct() {
        let mut fields = HashMap::new();
        fields.insert(Arc::from("a"), Value::Int(1));
        let s = Value::Struct(HairaStruct {
            type_name: Arc::from("T"),
            fields: Arc::new(fields.clone()),
        });
        assert_eq!(s.to_map(), Value::from(fields));
    }

    #[test]
    fn test_contains_string() {
        let s = Value::from("hello world");
        assert_eq!(s.contains(&Value::from("world")), Value::Bool(true));
        assert_eq!(s.contains(&Value::from("xyz")), Value::Bool(false));
    }

    // -- into_string --------------------------------------------------------

    #[test]
    fn test_into_string() {
        assert_eq!(Value::Int(42).into_string(), "42");
        assert_eq!(Value::from("hello").into_string(), "hello");
        assert_eq!(Value::None.into_string(), "none");
    }
}
