//! Process manager for llama-server.

use std::process::{Child, Command, Stdio};
use std::time::Duration;
use tokio::time::sleep;
use tracing::{debug, info, warn};

use crate::client::LlamaCppClient;
use crate::error::LocalAIError;
use crate::paths::{llama_server_path, model_path};
use crate::DEFAULT_PORT;

/// Manager for the llama-server process.
pub struct LlamaCppServer {
    port: u16,
    model_filename: String,
    process: Option<Child>,
}

impl LlamaCppServer {
    /// Create a new server manager with default settings.
    pub fn new(model_filename: impl Into<String>) -> Self {
        Self {
            port: DEFAULT_PORT,
            model_filename: model_filename.into(),
            process: None,
        }
    }

    /// Set a custom port.
    pub fn with_port(mut self, port: u16) -> Self {
        self.port = port;
        self
    }

    /// Get the port this server is configured to use.
    pub fn port(&self) -> u16 {
        self.port
    }

    /// Check if the server binary exists.
    pub fn binary_exists(&self) -> bool {
        llama_server_path().exists()
    }

    /// Check if the model file exists.
    pub fn model_exists(&self) -> bool {
        model_path(&self.model_filename).exists()
    }

    /// Start the llama-server process.
    pub fn start(&mut self) -> Result<(), LocalAIError> {
        let server_path = llama_server_path();
        let model = model_path(&self.model_filename);

        // Check prerequisites
        if !server_path.exists() {
            return Err(LocalAIError::ServerBinaryNotFound(
                server_path.display().to_string(),
            ));
        }

        if !model.exists() {
            return Err(LocalAIError::ModelNotFound(self.model_filename.clone()));
        }

        info!(
            "Starting llama-server on port {} with model {}",
            self.port, self.model_filename
        );

        let child = Command::new(&server_path)
            .arg("--model")
            .arg(&model)
            .arg("--host")
            .arg("127.0.0.1")
            .arg("--port")
            .arg(self.port.to_string())
            .arg("--ctx-size")
            .arg("8192")
            .arg("--n-predict")
            .arg("4096")
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .spawn()
            .map_err(|e| LocalAIError::ServerStartFailed(e.to_string()))?;

        debug!("llama-server process started with PID: {}", child.id());
        self.process = Some(child);

        Ok(())
    }

    /// Wait for the server to become ready.
    pub async fn wait_ready(&self, timeout: Duration) -> Result<(), LocalAIError> {
        let client = LlamaCppClient::with_port(self.port);
        let start = std::time::Instant::now();
        let check_interval = Duration::from_millis(500);

        info!("Waiting for llama-server to become ready...");

        while start.elapsed() < timeout {
            match client.check_health().await {
                Ok(()) => {
                    info!("llama-server is ready");
                    return Ok(());
                }
                Err(_) => {
                    // Check if process died
                    if let Some(ref process) = self.process {
                        // Try to check if process is still running
                        // Note: We can't use try_wait without &mut self
                        debug!("Server not ready yet, PID: {}", process.id());
                    }
                    sleep(check_interval).await;
                }
            }
        }

        Err(LocalAIError::ServerStartTimeout)
    }

    /// Stop the server process.
    pub fn stop(&mut self) -> Result<(), LocalAIError> {
        if let Some(mut child) = self.process.take() {
            info!("Stopping llama-server (PID: {})", child.id());

            // Try graceful shutdown first
            #[cfg(unix)]
            {
                unsafe {
                    libc::kill(child.id() as i32, libc::SIGTERM);
                }
                // Give it a moment to shut down gracefully
                std::thread::sleep(Duration::from_millis(500));
            }

            // Force kill if still running
            match child.try_wait() {
                Ok(Some(status)) => {
                    debug!("Server exited with status: {:?}", status);
                }
                Ok(None) => {
                    warn!("Server didn't exit gracefully, killing...");
                    let _ = child.kill();
                    let _ = child.wait();
                }
                Err(e) => {
                    warn!("Error checking server status: {}", e);
                    let _ = child.kill();
                }
            }
        }
        Ok(())
    }

    /// Check if the server process is running.
    pub fn is_running(&mut self) -> bool {
        if let Some(ref mut child) = self.process {
            match child.try_wait() {
                Ok(Some(_)) => {
                    // Process has exited
                    self.process = None;
                    false
                }
                Ok(None) => true, // Still running
                Err(_) => false,
            }
        } else {
            false
        }
    }

    /// Get a client connected to this server.
    pub fn client(&self) -> LlamaCppClient {
        LlamaCppClient::with_port(self.port)
    }
}

impl Drop for LlamaCppServer {
    fn drop(&mut self) {
        if self.process.is_some() {
            let _ = self.stop();
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_server_config() {
        let server = LlamaCppServer::new("test-model.gguf").with_port(9999);
        assert_eq!(server.port(), 9999);
    }
}
