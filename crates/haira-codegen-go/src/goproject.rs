//! GoProject — writes the generated Go project to disk.

use std::fs;
use std::path::PathBuf;

/// Represents a generated Go project ready to be written to disk.
pub struct GoProject {
    /// Directory to write the project into.
    pub dir: PathBuf,
    /// Contents of main.go.
    pub main_go: String,
    /// Path to the go-runtime module (for replace directive).
    pub runtime_path: PathBuf,
}

impl GoProject {
    /// Write all project files to disk.
    pub fn write(&self) -> Result<(), GoProjectError> {
        // Create directory
        fs::create_dir_all(&self.dir)
            .map_err(|e| GoProjectError::Io(format!("create dir: {}", e)))?;

        // Write go.mod
        let runtime_rel = self.runtime_path.display();
        let go_mod = format!(
            "module haira-generated\n\
             \n\
             go 1.22\n\
             \n\
             require haira-go-runtime v0.0.0\n\
             \n\
             replace haira-go-runtime => {}\n",
            runtime_rel
        );
        fs::write(self.dir.join("go.mod"), &go_mod)
            .map_err(|e| GoProjectError::Io(format!("write go.mod: {}", e)))?;

        // Write main.go
        fs::write(self.dir.join("main.go"), &self.main_go)
            .map_err(|e| GoProjectError::Io(format!("write main.go: {}", e)))?;

        Ok(())
    }
}

#[derive(Debug, thiserror::Error)]
pub enum GoProjectError {
    #[error("IO error: {0}")]
    Io(String),
}
