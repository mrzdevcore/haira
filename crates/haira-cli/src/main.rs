//! Haira CLI - Command-line interface for the Haira programming language.

use clap::{Parser, Subcommand};
use std::path::PathBuf;

mod commands;

/// Haira - A programming language for expressing intent
#[derive(Parser)]
#[command(name = "haira")]
#[command(author, version, about, long_about = None)]
struct Cli {
    /// Enable verbose output
    #[arg(short, long, global = true)]
    verbose: bool,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Build a Haira file to a native binary
    Build {
        /// Input file
        file: PathBuf,
        /// Output file (default: input filename without extension)
        #[arg(short, long)]
        output: Option<PathBuf>,
        /// Interpret AI blocks using Claude API (requires ANTHROPIC_API_KEY)
        #[arg(long)]
        interpret_ai: bool,
        /// Use local Ollama for AI interpretation (requires ollama server running)
        #[arg(long)]
        ollama: bool,
        /// Ollama model to use (default: deepseek-coder:6.7b)
        #[arg(long, default_value = "deepseek-coder:6.7b")]
        ollama_model: String,
        /// Use mock AI interpretation for testing (generates stub implementations)
        #[arg(long)]
        mock_ai: bool,
    },

    /// Build and run a Haira file
    Run {
        /// Input file
        file: PathBuf,
    },

    /// Parse a Haira file and show the AST
    Parse {
        /// Input file
        file: PathBuf,
        /// Output as JSON
        #[arg(long)]
        json: bool,
    },

    /// Check a Haira file for errors
    Check {
        /// Input file(s)
        files: Vec<PathBuf>,
    },

    /// Tokenize a Haira file and show tokens
    Lex {
        /// Input file
        file: PathBuf,
    },

    /// Show information about the Haira installation
    Info,

    /// Interpret a function name (test AI interpretation)
    Interpret {
        /// Function name to interpret
        name: String,
        /// Type context (JSON file)
        #[arg(long)]
        context: Option<PathBuf>,
    },
}

fn main() -> miette::Result<()> {
    let cli = Cli::parse();

    // Set up logging
    let filter = if cli.verbose { "debug" } else { "warn" };
    let subscriber = tracing_subscriber::fmt()
        .with_env_filter(tracing_subscriber::EnvFilter::new(filter))
        .without_time()
        .finish();
    tracing::subscriber::set_global_default(subscriber).ok();

    match cli.command {
        Commands::Build {
            file,
            output,
            interpret_ai,
            ollama,
            ollama_model,
            mock_ai,
        } => commands::build::run(
            &file,
            output.as_deref(),
            interpret_ai,
            ollama,
            &ollama_model,
            mock_ai,
        ),
        Commands::Run { file } => commands::run::run(&file),
        Commands::Parse { file, json } => commands::parse::run(&file, json),
        Commands::Check { files } => commands::check::run(&files),
        Commands::Lex { file } => commands::lex::run(&file),
        Commands::Info => commands::info::run(),
        Commands::Interpret { name, context } => tokio::runtime::Runtime::new()
            .unwrap()
            .block_on(commands::interpret::run(&name, context.as_deref())),
    }
}
