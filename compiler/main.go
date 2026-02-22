// Haira — An agentic orchestration programming language compiler.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/haira-lang/haira/internal/console"
	"github.com/haira-lang/haira/internal/driver"
	"github.com/haira-lang/haira/internal/lsp"
	"github.com/haira-lang/haira/internal/manifest"
)

var version = "dev"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]
	rest := args[1:]

	var err error
	switch cmd {
	case "build":
		err = cmdBuild(rest)
	case "run":
		err = cmdRun(rest)
	case "parse":
		err = cmdParse(rest)
	case "check":
		err = cmdCheck(rest)
	case "lex":
		err = cmdLex(rest)
	case "emit":
		err = cmdEmit(rest)
	case "test":
		err = cmdTest(rest)
	case "fmt":
		err = cmdFmt(rest)
	case "init":
		err = cmdInit(rest)
	case "console":
		err = cmdConsole(rest)
	case "lsp":
		err = lsp.RunStdio()
	case "version", "--version", "-v":
		fmt.Printf("haira %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

// resolveEntryFile resolves the source file from args or package.haira.
// Returns the file path and remaining args.
func resolveEntryFile(args []string) (string, []string) {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		return args[0], args[1:]
	}
	// Look for package.haira in current directory
	if pkg := manifest.Find("."); pkg != "" {
		p, err := manifest.Load(pkg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", err)
			return "", args
		}
		return p.Entry, args
	}
	return "", args
}

func cmdBuild(args []string) error {
	file, rest := resolveEntryFile(args)
	if file == "" {
		return fmt.Errorf("usage: haira build [file] [-o output]\n  No file specified and no package.haira found")
	}
	output := ""
	for i := 0; i < len(rest); i++ {
		if (rest[i] == "-o" || rest[i] == "--output") && i+1 < len(rest) {
			output = rest[i+1]
			i++
		}
	}
	return driver.Compile(file, output)
}

func cmdRun(args []string) error {
	file, _ := resolveEntryFile(args)
	if file == "" {
		return fmt.Errorf("usage: haira run [file]\n  No file specified and no package.haira found")
	}
	return driver.Run(file)
}

func cmdParse(args []string) error {
	file, _ := resolveEntryFile(args)
	if file == "" {
		return fmt.Errorf("usage: haira parse [file]\n  No file specified and no package.haira found")
	}
	return driver.ParseFile(file)
}

func cmdCheck(args []string) error {
	if len(args) == 0 {
		file, _ := resolveEntryFile(args)
		if file == "" {
			return fmt.Errorf("usage: haira check [file] [files...]\n  No file specified and no package.haira found")
		}
		return driver.CheckFile(file)
	}
	for _, file := range args {
		if err := driver.CheckFile(file); err != nil {
			return err
		}
	}
	return nil
}

func cmdLex(args []string) error {
	file, _ := resolveEntryFile(args)
	if file == "" {
		return fmt.Errorf("usage: haira lex [file]\n  No file specified and no package.haira found")
	}
	return driver.LexFile(file)
}

func cmdEmit(args []string) error {
	file, _ := resolveEntryFile(args)
	if file == "" {
		return fmt.Errorf("usage: haira emit [file]\n  No file specified and no package.haira found")
	}
	return driver.EmitFile(file)
}

func cmdTest(args []string) error {
	file, rest := resolveEntryFile(args)
	if file == "" {
		return fmt.Errorf("usage: haira test [file] [flags...]\n  No file specified and no package.haira found")
	}
	return driver.Test(file, rest)
}

func cmdFmt(args []string) error {
	if len(args) == 0 {
		file, _ := resolveEntryFile(args)
		if file == "" {
			return fmt.Errorf("usage: haira fmt [file] [files...]\n  No file specified and no package.haira found")
		}
		return driver.FormatFile(file)
	}
	for _, file := range args {
		if err := driver.FormatFile(file); err != nil {
			return err
		}
	}
	return nil
}

func cmdConsole(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: haira console <host:port>\n  Connect to a running Haira server")
	}
	return console.Run(args[0], args[1:])
}

func cmdInit(args []string) error {
	// Check if package.haira already exists
	if manifest.Find(".") != "" {
		return fmt.Errorf("package.haira already exists in current directory")
	}

	// Use directory name as project name
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	name := filepath.Base(dir)

	content := manifest.DefaultManifest(name)
	if err := os.WriteFile(manifest.Filename, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", manifest.Filename, err)
	}

	fmt.Printf("Created %s\n", manifest.Filename)
	return nil
}

func printUsage() {
	fmt.Print(`Haira — An agentic orchestration programming language

Usage: haira <command> [arguments]

Commands:
  build [file] [-o output]   Compile to a native binary
  run [file]                  Compile and execute
  parse [file]                Show the AST
  check [file] [files...]     Parse and report errors
  lex [file]                  Show tokens
  emit [file]                 Show generated Go code
  test [file] [flags...]      Run test blocks
  fmt [file] [files...]       Format source files in-place
  init                        Create a package.haira manifest
  console <host:port>         Connect to a Haira server (interactive terminal)
  lsp                         Start the language server (LSP)
  version                     Show version
  help                        Show this help

If no file is specified, haira looks for package.haira in the current directory.

`)
}
