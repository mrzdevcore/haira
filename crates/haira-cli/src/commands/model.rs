//! Model management commands.

use haira_ai::{local_ai_paths, ModelManager};
use std::path::Path;

/// List installed models.
pub(crate) fn list() -> miette::Result<()> {
    let manager = ModelManager::new();
    let models = manager
        .list_installed()
        .map_err(|e| miette::miette!("Failed to list models: {}", e))?;

    if models.is_empty() {
        println!("No models installed.");
        println!();
        println!("To install the default model, run:");
        println!("  haira model pull");
        return Ok(());
    }

    println!("Installed models:");
    for model in models {
        println!("  - {}", model);
    }

    println!();
    println!(
        "Models directory: {}",
        local_ai_paths::models_dir().display()
    );

    Ok(())
}

/// Pull/download a model.
pub(crate) async fn pull(path: Option<&Path>) -> miette::Result<()> {
    let manager = ModelManager::new();

    // Ensure directories exist
    local_ai_paths::ensure_dirs()
        .map_err(|e| miette::miette!("Failed to create data directories: {}", e))?;

    if let Some(source_path) = path {
        // Install from local file
        if !source_path.exists() {
            return Err(miette::miette!("File not found: {}", source_path.display()));
        }

        let ext = source_path.extension().and_then(|e| e.to_str());
        if ext != Some("gguf") {
            return Err(miette::miette!(
                "Expected a .gguf file, got: {}",
                source_path.display()
            ));
        }

        println!("Installing model from: {}", source_path.display());
        let dest = manager
            .install_from_path(&source_path.to_path_buf())
            .map_err(|e| miette::miette!("Failed to install model: {}", e))?;

        println!("Model installed to: {}", dest.display());
    } else {
        // Download default model
        let model = ModelManager::default_model();

        if manager.is_installed(&model.filename) {
            println!("Model '{}' is already installed.", model.name);
            return Ok(());
        }

        println!("Downloading model: {}", model.name);
        println!("This may take a while depending on your connection...");
        println!();

        match manager.download_default().await {
            Ok(path) => {
                println!();
                println!("Model downloaded successfully!");
                println!("Location: {}", path.display());
            }
            Err(e) => {
                return Err(miette::miette!("Failed to download model: {}", e));
            }
        }
    }

    Ok(())
}

/// Show information about models and paths.
pub(crate) fn info() -> miette::Result<()> {
    println!("Haira Local AI Configuration");
    println!("=============================");
    println!();
    println!(
        "Data directory:   {}",
        local_ai_paths::haira_data_dir().display()
    );
    println!(
        "Models directory: {}",
        local_ai_paths::models_dir().display()
    );
    println!("Binaries:         {}", local_ai_paths::bin_dir().display());
    println!();
    println!(
        "Server binary:    {}",
        local_ai_paths::llama_server_path().display()
    );
    println!("  Exists: {}", local_ai_paths::llama_server_path().exists());
    println!();

    let manager = ModelManager::new();
    let default_model = ModelManager::default_model();

    println!("Default model:    {}", default_model.filename);
    println!(
        "  Installed: {}",
        manager.is_installed(&default_model.filename)
    );
    println!();

    println!("Default port:     {}", haira_ai::DEFAULT_LOCAL_AI_PORT);

    Ok(())
}
