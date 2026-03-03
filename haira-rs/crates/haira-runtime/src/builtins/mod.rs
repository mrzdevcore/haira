// Built-in function registry for the Haira runtime.
//
// Registers core builtins: len, keys, join, panic, type_of, to_string, to_int, to_float.

use std::collections::HashMap;
use std::sync::Arc;

use crate::value::{RuntimeError, Value};

/// Registry of built-in functions available to Haira programs.
pub struct Builtins {
    funcs: HashMap<String, Arc<dyn Fn(Vec<Value>) -> Result<Value, RuntimeError> + Send + Sync>>,
}

impl Builtins {
    /// Create a new builtin registry with all core builtins registered.
    pub fn new() -> Self {
        let mut funcs: HashMap<
            String,
            Arc<dyn Fn(Vec<Value>) -> Result<Value, RuntimeError> + Send + Sync>,
        > = HashMap::new();

        // len(v) -> v.len()
        funcs.insert(
            "len".to_string(),
            Arc::new(|args| {
                check_arity("len", &args, 1)?;
                Ok(args[0].len())
            }),
        );

        // keys(v) -> v.keys()
        funcs.insert(
            "keys".to_string(),
            Arc::new(|args| {
                check_arity("keys", &args, 1)?;
                Ok(args[0].keys())
            }),
        );

        // join(list, sep) -> join list elements with separator
        funcs.insert(
            "join".to_string(),
            Arc::new(|args| {
                check_arity("join", &args, 2)?;
                let items = match &args[0] {
                    Value::List(v) => v,
                    _ => {
                        return Err(RuntimeError::new(format!(
                            "join: expected list as first argument, got {}",
                            args[0].type_name()
                        )));
                    }
                };
                let sep = match &args[1] {
                    Value::Str(s) => s.as_ref().to_string(),
                    other => format!("{other}"),
                };
                let parts: Vec<String> = items.iter().map(|v| format!("{v}")).collect();
                Ok(Value::Str(Arc::from(parts.join(&sep))))
            }),
        );

        // panic(msg) -> RuntimeError
        funcs.insert(
            "panic".to_string(),
            Arc::new(|args| {
                check_arity("panic", &args, 1)?;
                Err(RuntimeError::new(format!("{}", args[0])))
            }),
        );

        // type_of(v) -> Str of type name
        funcs.insert(
            "type_of".to_string(),
            Arc::new(|args| {
                check_arity("type_of", &args, 1)?;
                Ok(Value::Str(Arc::from(args[0].type_name())))
            }),
        );

        // to_string(v) -> converts value to string
        funcs.insert(
            "to_string".to_string(),
            Arc::new(|args| {
                check_arity("to_string", &args, 1)?;
                Ok(Value::Str(Arc::from(format!("{}", args[0]))))
            }),
        );

        // to_int(v) -> converts to int
        funcs.insert(
            "to_int".to_string(),
            Arc::new(|args| {
                check_arity("to_int", &args, 1)?;
                match &args[0] {
                    Value::Int(_) => Ok(args[0].clone()),
                    Value::Float(n) => Ok(Value::Int(*n as i64)),
                    Value::Bool(b) => Ok(Value::Int(if *b { 1 } else { 0 })),
                    Value::Str(s) => match s.parse::<i64>() {
                        Ok(n) => Ok(Value::Int(n)),
                        Err(_) => Err(RuntimeError::new(format!(
                            "to_int: cannot convert string \"{s}\" to int"
                        ))),
                    },
                    other => Err(RuntimeError::new(format!(
                        "to_int: cannot convert {} to int",
                        other.type_name()
                    ))),
                }
            }),
        );

        // to_float(v) -> converts to float
        funcs.insert(
            "to_float".to_string(),
            Arc::new(|args| {
                check_arity("to_float", &args, 1)?;
                match &args[0] {
                    Value::Float(_) => Ok(args[0].clone()),
                    Value::Int(n) => Ok(Value::Float(*n as f64)),
                    Value::Bool(b) => Ok(Value::Float(if *b { 1.0 } else { 0.0 })),
                    Value::Str(s) => match s.parse::<f64>() {
                        Ok(n) => Ok(Value::Float(n)),
                        Err(_) => Err(RuntimeError::new(format!(
                            "to_float: cannot convert string \"{s}\" to float"
                        ))),
                    },
                    other => Err(RuntimeError::new(format!(
                        "to_float: cannot convert {} to float",
                        other.type_name()
                    ))),
                }
            }),
        );

        Self { funcs }
    }

    /// Look up a builtin function by name.
    pub fn get(
        &self,
        name: &str,
    ) -> Option<&Arc<dyn Fn(Vec<Value>) -> Result<Value, RuntimeError> + Send + Sync>> {
        self.funcs.get(name)
    }

    /// Returns an iterator over all registered builtin names.
    pub fn names(&self) -> impl Iterator<Item = &str> {
        self.funcs.keys().map(|s| s.as_str())
    }
}

impl Default for Builtins {
    fn default() -> Self {
        Self::new()
    }
}

/// Helper: check that the argument count matches the expected arity.
fn check_arity(name: &str, args: &[Value], expected: usize) -> Result<(), RuntimeError> {
    if args.len() != expected {
        Err(RuntimeError::new(format!(
            "{name}: expected {expected} argument(s), got {}",
            args.len()
        )))
    } else {
        Ok(())
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::Arc;

    fn builtins() -> Builtins {
        Builtins::new()
    }

    #[test]
    fn test_builtin_len_list() {
        let b = builtins();
        let f = b.get("len").unwrap();
        let result = f(vec![Value::from(vec![Value::Int(1), Value::Int(2)])]).unwrap();
        assert_eq!(result, Value::Int(2));
    }

    #[test]
    fn test_builtin_len_string() {
        let b = builtins();
        let f = b.get("len").unwrap();
        let result = f(vec![Value::from("hello")]).unwrap();
        assert_eq!(result, Value::Int(5));
    }

    #[test]
    fn test_builtin_len_arity_error() {
        let b = builtins();
        let f = b.get("len").unwrap();
        let result = f(vec![]);
        assert!(result.is_err());
    }

    #[test]
    fn test_builtin_keys() {
        let b = builtins();
        let f = b.get("keys").unwrap();
        let mut m = std::collections::HashMap::new();
        m.insert(Arc::from("x"), Value::Int(1));
        let result = f(vec![Value::from(m)]).unwrap();
        if let Value::List(keys) = result {
            assert_eq!(keys.len(), 1);
            assert_eq!(keys[0], Value::from("x"));
        } else {
            panic!("expected list");
        }
    }

    #[test]
    fn test_builtin_join() {
        let b = builtins();
        let f = b.get("join").unwrap();
        let list = Value::from(vec![Value::from("a"), Value::from("b"), Value::from("c")]);
        let result = f(vec![list, Value::from(", ")]).unwrap();
        assert_eq!(result, Value::from("a, b, c"));
    }

    #[test]
    fn test_builtin_join_with_numbers() {
        let b = builtins();
        let f = b.get("join").unwrap();
        let list = Value::from(vec![Value::Int(1), Value::Int(2), Value::Int(3)]);
        let result = f(vec![list, Value::from("-")]).unwrap();
        assert_eq!(result, Value::from("1-2-3"));
    }

    #[test]
    fn test_builtin_panic() {
        let b = builtins();
        let f = b.get("panic").unwrap();
        let result = f(vec![Value::from("something went wrong")]);
        assert!(result.is_err());
        assert_eq!(result.unwrap_err().message, "something went wrong");
    }

    #[test]
    fn test_builtin_type_of() {
        let b = builtins();
        let f = b.get("type_of").unwrap();
        assert_eq!(f(vec![Value::Int(42)]).unwrap(), Value::from("int"));
        assert_eq!(f(vec![Value::from("hi")]).unwrap(), Value::from("string"));
        assert_eq!(f(vec![Value::None]).unwrap(), Value::from("none"));
        assert_eq!(f(vec![Value::Bool(true)]).unwrap(), Value::from("bool"));
    }

    #[test]
    fn test_builtin_to_string() {
        let b = builtins();
        let f = b.get("to_string").unwrap();
        assert_eq!(f(vec![Value::Int(42)]).unwrap(), Value::from("42"));
        assert_eq!(f(vec![Value::Bool(true)]).unwrap(), Value::from("true"));
    }

    #[test]
    fn test_builtin_to_int() {
        let b = builtins();
        let f = b.get("to_int").unwrap();

        assert_eq!(f(vec![Value::Int(42)]).unwrap(), Value::Int(42));
        assert_eq!(f(vec![Value::Float(3.7)]).unwrap(), Value::Int(3));
        assert_eq!(f(vec![Value::Bool(true)]).unwrap(), Value::Int(1));
        assert_eq!(f(vec![Value::from("123")]).unwrap(), Value::Int(123));

        // Invalid string
        assert!(f(vec![Value::from("abc")]).is_err());
    }

    #[test]
    fn test_builtin_to_float() {
        let b = builtins();
        let f = b.get("to_float").unwrap();

        assert_eq!(f(vec![Value::Float(3.14)]).unwrap(), Value::Float(3.14));
        assert_eq!(f(vec![Value::Int(5)]).unwrap(), Value::Float(5.0));
        assert_eq!(f(vec![Value::Bool(false)]).unwrap(), Value::Float(0.0));
        assert_eq!(f(vec![Value::from("2.5")]).unwrap(), Value::Float(2.5));

        // Invalid string
        assert!(f(vec![Value::from("xyz")]).is_err());
    }

    #[test]
    fn test_builtin_names_count() {
        let b = builtins();
        let names: Vec<&str> = b.names().collect();
        assert_eq!(names.len(), 8); // len, keys, join, panic, type_of, to_string, to_int, to_float
    }

    #[test]
    fn test_missing_builtin_returns_none() {
        let b = builtins();
        assert!(b.get("nonexistent").is_none());
    }
}
