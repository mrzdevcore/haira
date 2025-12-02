//! Haira Runtime Library
//!
//! This crate provides the runtime functions for compiled Haira programs.
//! All functions use C ABI for compatibility with Cranelift-generated code.

#![allow(clippy::not_unsafe_ptr_arg_deref)]

mod concurrency;
mod env;
mod error;
mod io;
mod math;
mod memory;
mod strings;
mod testing;
mod time;

// Re-export all runtime functions
pub use concurrency::*;
pub use env::*;
pub use error::*;
pub use io::*;
pub use math::*;
pub use memory::*;
pub use strings::*;
pub use testing::*;
pub use time::*;
