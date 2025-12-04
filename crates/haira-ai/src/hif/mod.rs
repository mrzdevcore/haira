//! Haira Intent Format (HIF) - Human-readable cache format for AI-generated code.
//!
//! The HIF format stores AI-inferred struct types and intent functions in a
//! human-readable, Git-friendly format with full type annotations.
//!
//! ## Example
//!
//! ```hif
//! # Haira Intent Format v1
//!
//! struct User @abc123
//!   name: string
//!   age: int
//!   email: string
//!
//! intent get_user_name @def456
//!   params user: User
//!   returns string
//!   body
//!     get_field user.name -> _name: string
//!     return _name
//! ```

mod inference;
mod parser;
mod types;
mod writer;

pub use inference::*;
pub use parser::*;
pub use types::*;
pub use writer::*;

/// Current HIF format version.
pub const HIF_VERSION: u32 = 1;

/// HIF file extension.
pub const HIF_EXTENSION: &str = "hif";

/// Default HIF filename.
pub const HIF_FILENAME: &str = "haira.hif";
