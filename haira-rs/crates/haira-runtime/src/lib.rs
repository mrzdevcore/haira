// Haira runtime — Value type system, core operations, and built-in functions.
//
// This crate defines the dynamic value type used at runtime by compiled Haira programs.
// Every Haira value is represented as a `Value` enum. The runtime also provides the
// core built-in operations (get, set, len, to_slice, etc.) and a builtin function registry.

pub mod builtins;
pub mod value;

// Re-export primary types at crate root for convenience.
pub use builtins::Builtins;
pub use value::{HairaFn, HairaStruct, RuntimeError, Value};
