// Haira runtime — Value type system, core operations, agentic runtime.
//
// This crate defines the dynamic value type used at runtime by compiled Haira programs.
// Every Haira value is represented as a `Value` enum. The runtime also provides the
// core built-in operations (get, set, len, to_slice, etc.), a builtin function registry,
// and the agentic runtime (providers, tools, agents, memory, LLM client).

pub mod agent;
pub mod builtins;
pub mod llm;
pub mod memory;
pub mod provider;
pub mod tool;
pub mod value;

// Re-export primary types at crate root for convenience.
pub use agent::{Agent, AgentStep};
pub use builtins::Builtins;
pub use memory::{MemoryConfig, SessionStore};
pub use provider::Provider;
pub use tool::{build_tool_schema, ToolDef, ToolRegistry};
pub use value::{HairaFn, HairaStruct, RuntimeError, Value};
