use zed_extension_api::{self as zed, Result};

struct HairaExtension;

impl zed::Extension for HairaExtension {
    fn new() -> Self {
        HairaExtension
    }

    fn language_server_command(
        &mut self,
        _language_server_id: &zed::LanguageServerId,
        worktree: &zed::Worktree,
    ) -> Result<zed::Command> {
        let path = worktree
            .which("haira")
            .ok_or_else(|| {
                concat!(
                    "haira binary not found in PATH. ",
                    "Install the Haira compiler and ensure `haira` is on your PATH."
                )
                .to_string()
            })?;

        Ok(zed::Command {
            command: path,
            args: vec!["lsp".to_string()],
            env: worktree.shell_env(),
        })
    }
}

zed::register_extension!(HairaExtension);
