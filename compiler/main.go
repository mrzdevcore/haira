// Haira — An agentic orchestration programming language compiler.
package main

import (
	"fmt"
	"os"

	"github.com/haira-lang/haira/internal/driver"
	"github.com/haira-lang/haira/internal/lsp"
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

func cmdBuild(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: haira build <file> [-o output]")
	}
	file := args[0]
	output := ""
	for i := 1; i < len(args); i++ {
		if (args[i] == "-o" || args[i] == "--output") && i+1 < len(args) {
			output = args[i+1]
			i++
		}
	}
	return driver.Compile(file, output)
}

func cmdRun(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: haira run <file>")
	}
	return driver.Run(args[0])
}

func cmdParse(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: haira parse <file>")
	}
	return driver.ParseFile(args[0])
}

func cmdCheck(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: haira check <file> [files...]")
	}
	for _, file := range args {
		if err := driver.CheckFile(file); err != nil {
			return err
		}
	}
	return nil
}

func cmdLex(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: haira lex <file>")
	}
	return driver.LexFile(args[0])
}

func cmdEmit(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: haira emit <file>")
	}
	return driver.EmitFile(args[0])
}

func printUsage() {
	fmt.Print(`Haira — An agentic orchestration programming language

Usage: haira <command> [arguments]

Commands:
  build <file> [-o output]   Compile to a native binary
  run <file>                 Compile and execute
  parse <file>               Show the AST
  check <file> [files...]    Parse and report errors
  lex <file>                 Show tokens
  emit <file>                Show generated Go code
  lsp                        Start the language server (LSP)
  version                    Show version
  help                       Show this help

`)
}
