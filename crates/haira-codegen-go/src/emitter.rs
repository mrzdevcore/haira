//! GoEmitter — indented Go source code writer.

use std::collections::HashSet;

/// Writes Go source code with proper indentation.
pub struct GoEmitter {
    buf: String,
    indent: usize,
    /// Track declared variables to emit `:=` only on first assignment.
    declared_vars: HashSet<String>,
}

impl GoEmitter {
    pub fn new() -> Self {
        Self {
            buf: String::with_capacity(4096),
            indent: 0,
            declared_vars: HashSet::new(),
        }
    }

    /// Check if a variable has been declared; if not, mark it as declared.
    /// Returns true if this is a new declaration (use `:=`), false if already declared (use `=`).
    pub fn declare_var(&mut self, name: &str) -> bool {
        self.declared_vars.insert(name.to_string())
    }

    /// Reset declared variables — call at the start of each function/workflow.
    pub fn reset_vars(&mut self) {
        self.declared_vars.clear();
    }

    /// Write a line with current indentation.
    pub fn line(&mut self, s: &str) {
        if s.is_empty() {
            self.buf.push('\n');
        } else {
            for _ in 0..self.indent {
                self.buf.push('\t');
            }
            self.buf.push_str(s);
            self.buf.push('\n');
        }
    }

    /// Write text without a trailing newline, with indentation.
    pub fn write(&mut self, s: &str) {
        for _ in 0..self.indent {
            self.buf.push('\t');
        }
        self.buf.push_str(s);
    }

    /// Write text without indentation or newline (append to current line).
    pub fn append(&mut self, s: &str) {
        self.buf.push_str(s);
    }

    /// Write a formatted line.
    pub fn linef(&mut self, args: std::fmt::Arguments<'_>) {
        let s = std::fmt::format(args);
        self.line(&s);
    }

    /// Increase indentation.
    pub fn indent(&mut self) {
        self.indent += 1;
    }

    /// Decrease indentation.
    pub fn dedent(&mut self) {
        self.indent = self.indent.saturating_sub(1);
    }

    /// Write an opening brace and indent.
    pub fn open_block(&mut self, prefix: &str) {
        self.line(&format!("{} {{", prefix));
        self.indent();
    }

    /// Dedent and write a closing brace.
    pub fn close_block(&mut self) {
        self.dedent();
        self.line("}");
    }

    /// Write a blank line.
    pub fn blank(&mut self) {
        self.buf.push('\n');
    }

    /// Consume the emitter and return the generated source.
    pub fn finish(self) -> String {
        self.buf
    }
}
