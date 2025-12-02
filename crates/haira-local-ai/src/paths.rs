//! Path utilities for Haira data directories.

use std::path::PathBuf;

/// Get the Haira data directory (~/.haira/).
pub fn haira_data_dir() -> PathBuf {
    dirs::home_dir()
        .expect("Could not determine home directory")
        .join(".haira")
}

/// Get the models directory (~/.haira/models/).
pub fn models_dir() -> PathBuf {
    haira_data_dir().join("models")
}

/// Get the bin directory (~/.haira/bin/).
pub fn bin_dir() -> PathBuf {
    haira_data_dir().join("bin")
}

/// Get the path to the llama-server binary.
pub fn llama_server_path() -> PathBuf {
    let binary_name = if cfg!(target_os = "windows") {
        "llama-server.exe"
    } else {
        "llama-server"
    };
    bin_dir().join(binary_name)
}

/// Get the path to a model file.
pub fn model_path(filename: &str) -> PathBuf {
    models_dir().join(filename)
}

/// Ensure the Haira data directories exist.
pub fn ensure_dirs() -> std::io::Result<()> {
    std::fs::create_dir_all(haira_data_dir())?;
    std::fs::create_dir_all(models_dir())?;
    std::fs::create_dir_all(bin_dir())?;
    Ok(())
}
