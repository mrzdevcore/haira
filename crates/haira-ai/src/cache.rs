//! AI response caching for reproducibility.

use sha2::{Digest, Sha256};
use std::collections::HashMap;
use std::fs;
use std::path::PathBuf;
use thiserror::Error;

use haira_cir::CIRFunction;

/// Cache for AI-generated functions.
pub struct AICache {
    /// Cache directory
    cache_dir: PathBuf,
    /// In-memory cache
    memory: HashMap<String, CIRFunction>,
}

/// Cache errors.
#[derive(Debug, Error)]
pub enum CacheError {
    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),
    #[error("JSON error: {0}")]
    Json(#[from] serde_json::Error),
}

impl AICache {
    /// Create a new cache.
    pub fn new(cache_dir: PathBuf) -> Self {
        Self {
            cache_dir,
            memory: HashMap::new(),
        }
    }

    /// Generate a cache key from function name and context.
    pub fn cache_key(function_name: &str, context_json: &str) -> String {
        let mut hasher = Sha256::new();
        hasher.update(function_name.as_bytes());
        hasher.update(b":");
        hasher.update(context_json.as_bytes());
        let result = hasher.finalize();
        base64::Engine::encode(&base64::engine::general_purpose::URL_SAFE_NO_PAD, result)
    }

    /// Get a cached function.
    pub fn get(&self, key: &str) -> Option<CIRFunction> {
        // Check memory cache first
        if let Some(func) = self.memory.get(key) {
            return Some(func.clone());
        }

        // Check disk cache
        let path = self.cache_dir.join(format!("{}.cir.json", key));
        if path.exists() {
            if let Ok(content) = fs::read_to_string(&path) {
                if let Ok(func) = serde_json::from_str(&content) {
                    return Some(func);
                }
            }
        }

        None
    }

    /// Store a function in the cache.
    pub fn set(&mut self, key: &str, func: &CIRFunction) -> Result<(), CacheError> {
        // Store in memory
        self.memory.insert(key.to_string(), func.clone());

        // Store on disk
        fs::create_dir_all(&self.cache_dir)?;
        let path = self.cache_dir.join(format!("{}.cir.json", key));
        let content = serde_json::to_string_pretty(func)?;
        fs::write(path, content)?;

        Ok(())
    }

    /// Check if a key exists in the cache.
    pub fn contains(&self, key: &str) -> bool {
        if self.memory.contains_key(key) {
            return true;
        }

        let path = self.cache_dir.join(format!("{}.cir.json", key));
        path.exists()
    }

    /// Clear the cache.
    pub fn clear(&mut self) -> Result<(), CacheError> {
        self.memory.clear();

        if self.cache_dir.exists() {
            for entry in fs::read_dir(&self.cache_dir)? {
                let entry = entry?;
                if entry.path().extension().map_or(false, |e| e == "json") {
                    fs::remove_file(entry.path())?;
                }
            }
        }

        Ok(())
    }

    /// List all cached keys.
    pub fn list_keys(&self) -> Result<Vec<String>, CacheError> {
        let mut keys: Vec<String> = self.memory.keys().cloned().collect();

        if self.cache_dir.exists() {
            for entry in fs::read_dir(&self.cache_dir)? {
                let entry = entry?;
                let path = entry.path();
                if path.extension().map_or(false, |e| e == "json") {
                    if let Some(stem) = path.file_stem() {
                        let stem = stem.to_string_lossy();
                        if let Some(key) = stem.strip_suffix(".cir") {
                            if !keys.contains(&key.to_string()) {
                                keys.push(key.to_string());
                            }
                        }
                    }
                }
            }
        }

        Ok(keys)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::tempdir;

    #[test]
    fn test_cache_key_deterministic() {
        let key1 = AICache::cache_key("get_users", r#"{"types":[]}"#);
        let key2 = AICache::cache_key("get_users", r#"{"types":[]}"#);
        assert_eq!(key1, key2);
    }

    #[test]
    fn test_cache_key_different() {
        let key1 = AICache::cache_key("get_users", r#"{"types":[]}"#);
        let key2 = AICache::cache_key("get_posts", r#"{"types":[]}"#);
        assert_ne!(key1, key2);
    }

    #[test]
    fn test_cache_roundtrip() {
        let dir = tempdir().unwrap();
        let mut cache = AICache::new(dir.path().to_path_buf());

        let func = CIRFunction::new("test")
            .with_param("x", "int")
            .returning("int");

        cache.set("test_key", &func).unwrap();

        let loaded = cache.get("test_key").unwrap();
        assert_eq!(loaded.name, "test");
    }
}
