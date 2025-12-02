//! Model download and management.

use futures_util::StreamExt;
use indicatif::{ProgressBar, ProgressStyle};
use sha2::{Digest, Sha256};
use std::fs::{self, File};
use std::io::Write;
use std::path::PathBuf;
use tracing::{debug, info};

use crate::error::LocalAIError;
use crate::paths::{ensure_dirs, model_path, models_dir};
use crate::{DEFAULT_MODEL_FILENAME, DEFAULT_MODEL_NAME};

/// Model registry entry.
#[derive(Debug, Clone)]
pub struct ModelInfo {
    /// Display name of the model.
    pub name: String,
    /// Filename on disk.
    pub filename: String,
    /// Download URL.
    pub url: String,
    /// Expected SHA256 checksum (optional).
    pub sha256: Option<String>,
    /// Size in bytes (for progress display).
    pub size_bytes: Option<u64>,
}

/// Manager for downloading and managing models.
pub struct ModelManager {
    client: reqwest::Client,
}

impl ModelManager {
    /// Create a new model manager.
    pub fn new() -> Self {
        Self {
            client: reqwest::Client::new(),
        }
    }

    /// Get the default Haira model info.
    ///
    /// Note: Update this when the fine-tuned model is available.
    pub fn default_model() -> ModelInfo {
        ModelInfo {
            name: DEFAULT_MODEL_NAME.to_string(),
            filename: DEFAULT_MODEL_FILENAME.to_string(),
            // Placeholder URL - update when model is hosted
            url: format!(
                "https://github.com/haira-lang/haira/releases/download/models-v1/{}",
                DEFAULT_MODEL_FILENAME
            ),
            sha256: None, // Add checksum when model is released
            size_bytes: None,
        }
    }

    /// List all installed models.
    pub fn list_installed(&self) -> Result<Vec<String>, LocalAIError> {
        let dir = models_dir();

        if !dir.exists() {
            return Ok(vec![]);
        }

        let models: Vec<String> = fs::read_dir(&dir)?
            .filter_map(|entry| entry.ok())
            .filter(|entry| {
                entry
                    .path()
                    .extension()
                    .map(|ext| ext == "gguf")
                    .unwrap_or(false)
            })
            .filter_map(|entry| {
                entry
                    .file_name()
                    .to_str()
                    .map(|s| s.trim_end_matches(".gguf").to_string())
            })
            .collect();

        Ok(models)
    }

    /// Check if a model is installed.
    pub fn is_installed(&self, filename: &str) -> bool {
        model_path(filename).exists()
    }

    /// Get the path to an installed model.
    pub fn get_model_path(&self, filename: &str) -> Option<PathBuf> {
        let path = model_path(filename);
        if path.exists() {
            Some(path)
        } else {
            None
        }
    }

    /// Download a model from URL.
    pub async fn download(&self, model: &ModelInfo) -> Result<PathBuf, LocalAIError> {
        ensure_dirs()?;

        let dest_path = model_path(&model.filename);

        info!("Downloading model '{}' to {:?}", model.name, dest_path);

        let response = self
            .client
            .get(&model.url)
            .send()
            .await
            .map_err(|e| LocalAIError::DownloadFailed(e.to_string()))?;

        if !response.status().is_success() {
            return Err(LocalAIError::DownloadFailed(format!(
                "HTTP {}: {}",
                response.status(),
                model.url
            )));
        }

        let total_size = response.content_length().or(model.size_bytes);

        // Create progress bar
        let pb = if let Some(size) = total_size {
            let pb = ProgressBar::new(size);
            pb.set_style(
                ProgressStyle::default_bar()
                    .template("{spinner:.green} [{elapsed_precise}] [{bar:40.cyan/blue}] {bytes}/{total_bytes} ({eta})")
                    .expect("Invalid progress bar template")
                    .progress_chars("#>-"),
            );
            pb
        } else {
            let pb = ProgressBar::new_spinner();
            pb.set_style(
                ProgressStyle::default_spinner()
                    .template("{spinner:.green} [{elapsed_precise}] {bytes} downloaded")
                    .expect("Invalid progress bar template"),
            );
            pb
        };

        // Download with progress
        let mut file = File::create(&dest_path)?;
        let mut hasher = Sha256::new();
        let mut stream = response.bytes_stream();
        let mut downloaded: u64 = 0;

        while let Some(chunk) = stream.next().await {
            let chunk = chunk.map_err(|e| LocalAIError::DownloadFailed(e.to_string()))?;
            file.write_all(&chunk)?;
            hasher.update(&chunk);
            downloaded += chunk.len() as u64;
            pb.set_position(downloaded);
        }

        pb.finish_with_message("Download complete");

        // Verify checksum if provided
        if let Some(expected) = &model.sha256 {
            let actual = hex::encode(hasher.finalize());
            if actual != *expected {
                // Remove corrupted file
                let _ = fs::remove_file(&dest_path);
                return Err(LocalAIError::ChecksumMismatch {
                    expected: expected.clone(),
                    actual,
                });
            }
            debug!("Checksum verified: {}", actual);
        }

        info!("Model '{}' downloaded successfully", model.name);
        Ok(dest_path)
    }

    /// Download the default Haira model.
    pub async fn download_default(&self) -> Result<PathBuf, LocalAIError> {
        let model = Self::default_model();
        self.download(&model).await
    }

    /// Install a model from a local file path.
    pub fn install_from_path(&self, source: &PathBuf) -> Result<PathBuf, LocalAIError> {
        ensure_dirs()?;

        let filename = source
            .file_name()
            .and_then(|n| n.to_str())
            .ok_or_else(|| LocalAIError::ModelNotFound("Invalid path".to_string()))?;

        let dest_path = model_path(filename);

        if source == &dest_path {
            // Already in the right place
            return Ok(dest_path);
        }

        info!("Installing model from {:?} to {:?}", source, dest_path);
        fs::copy(source, &dest_path)?;

        Ok(dest_path)
    }

    /// Remove an installed model.
    pub fn remove(&self, filename: &str) -> Result<(), LocalAIError> {
        let path = model_path(filename);
        if path.exists() {
            fs::remove_file(&path)?;
            info!("Removed model: {}", filename);
        }
        Ok(())
    }
}

impl Default for ModelManager {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_model_info() {
        let model = ModelManager::default_model();
        assert_eq!(model.name, DEFAULT_MODEL_NAME);
        assert_eq!(model.filename, DEFAULT_MODEL_FILENAME);
    }
}
