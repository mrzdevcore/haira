//! Haira Go transpiler backend.
//!
//! Transforms a Haira AST into a Go project, then uses `go build`
//! to produce a native binary. Users never see the generated Go code.

pub mod agentic;
pub mod codegen;
pub mod emitter;
pub mod expressions;
pub mod goproject;
pub mod statements;
pub mod stdlib;
pub mod types;

use haira_ast::SourceFile;
use std::path::{Path, PathBuf};
use std::process::Command;

pub use goproject::{GoProject, GoProjectError};

/// Error type for Go codegen.
#[derive(Debug, thiserror::Error)]
pub enum GoCodegenError {
    #[error("Failed to write Go project: {0}")]
    Project(#[from] GoProjectError),
    #[error("go build failed: {0}")]
    GoBuild(String),
    #[error("go not found in PATH — install Go from https://go.dev/dl/")]
    GoNotFound,
    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),
}

/// Compile a Haira AST to a native binary via Go transpilation.
///
/// 1. Generate Go source files in a temp directory
/// 2. Run `go build -o <output> .`
/// 3. Clean up temp directory
pub fn compile_to_binary(
    ast: &SourceFile,
    output: &Path,
    runtime_path: &Path,
) -> Result<(), GoCodegenError> {
    // Create temp directory for Go project
    let tmp_dir = std::env::temp_dir().join(format!("haira-build-{}", std::process::id()));

    let project = codegen::generate_go_project(ast, runtime_path.to_path_buf(), tmp_dir.clone());
    project.write()?;

    // Resolve transitive dependencies
    run_go_mod_tidy(&tmp_dir)?;

    // Run go build
    let result = run_go_build(&tmp_dir, output);

    // Clean up temp directory (skip on error for debugging)
    if result.is_ok() {
        let _ = std::fs::remove_dir_all(&tmp_dir);
    } else {
        eprintln!("Debug: generated Go project at {}", tmp_dir.display());
    }

    result
}

/// Run a Haira program via `go run`.
pub fn run_program(
    ast: &SourceFile,
    runtime_path: &Path,
) -> Result<std::process::ExitStatus, GoCodegenError> {
    let tmp_dir = std::env::temp_dir().join(format!("haira-run-{}", std::process::id()));

    let project = codegen::generate_go_project(ast, runtime_path.to_path_buf(), tmp_dir.clone());
    project.write()?;

    // Resolve transitive dependencies
    run_go_mod_tidy(&tmp_dir)?;

    // Run go run .
    let status = Command::new("go")
        .arg("run")
        .arg(".")
        .current_dir(&tmp_dir)
        .status()
        .map_err(|e| {
            if e.kind() == std::io::ErrorKind::NotFound {
                GoCodegenError::GoNotFound
            } else {
                GoCodegenError::Io(e)
            }
        })?;

    // Clean up
    let _ = std::fs::remove_dir_all(&tmp_dir);

    Ok(status)
}

/// Find the go-runtime directory relative to the haira binary.
pub fn find_runtime_path() -> Option<PathBuf> {
    // Try relative to current exe (for development: project_root/go-runtime)
    if let Ok(exe) = std::env::current_exe() {
        // In development: exe is in target/debug/haira
        // go-runtime is at project_root/go-runtime
        let mut path = exe.clone();
        // Go up from target/debug/haira
        for _ in 0..3 {
            path.pop();
        }
        let runtime = path.join("go-runtime");
        if runtime.exists() {
            return Some(runtime);
        }

        // Also try target/release/haira
        let mut path = exe;
        for _ in 0..3 {
            path.pop();
        }
        let runtime = path.join("go-runtime");
        if runtime.exists() {
            return Some(runtime);
        }
    }

    // Try current working directory
    let cwd = std::env::current_dir().ok()?;
    let runtime = cwd.join("go-runtime");
    if runtime.exists() {
        return Some(runtime);
    }

    None
}

fn run_go_mod_tidy(project_dir: &Path) -> Result<(), GoCodegenError> {
    let result = Command::new("go")
        .arg("mod")
        .arg("tidy")
        .current_dir(project_dir)
        .output()
        .map_err(|e| {
            if e.kind() == std::io::ErrorKind::NotFound {
                GoCodegenError::GoNotFound
            } else {
                GoCodegenError::Io(e)
            }
        })?;

    if !result.status.success() {
        let stderr = String::from_utf8_lossy(&result.stderr);
        return Err(GoCodegenError::GoBuild(format!(
            "go mod tidy failed: {}",
            stderr
        )));
    }

    Ok(())
}

fn run_go_build(project_dir: &Path, output: &Path) -> Result<(), GoCodegenError> {
    let output_abs = if output.is_absolute() {
        output.to_path_buf()
    } else {
        std::env::current_dir()
            .map_err(GoCodegenError::Io)?
            .join(output)
    };

    let result = Command::new("go")
        .arg("build")
        .arg("-o")
        .arg(&output_abs)
        .arg(".")
        .current_dir(project_dir)
        .output()
        .map_err(|e| {
            if e.kind() == std::io::ErrorKind::NotFound {
                GoCodegenError::GoNotFound
            } else {
                GoCodegenError::Io(e)
            }
        })?;

    if !result.status.success() {
        let stderr = String::from_utf8_lossy(&result.stderr);
        return Err(GoCodegenError::GoBuild(stderr.to_string()));
    }

    Ok(())
}
