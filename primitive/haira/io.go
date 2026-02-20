// Package haira provides the Haira standard library runtime for Go.
package haira

import (
	"bufio"
	"fmt"
	"os"
)

// Println prints values followed by a newline.
func Println(args ...any) {
	fmt.Println(args...)
}

// Print prints values without a trailing newline.
func Print(args ...any) {
	fmt.Print(args...)
}

// Printf prints a formatted string.
func Printf(format string, args ...any) {
	fmt.Printf(format, args...)
}

// Readln reads a line from stdin and returns it (without trailing newline).
func Readln() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

// ReadFile reads a file and returns its contents as a string.
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Eprintln prints values to stderr followed by a newline.
func Eprintln(args ...any) {
	fmt.Fprintln(os.Stderr, args...)
}

// Eprintf prints a formatted string to stderr.
func Eprintf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
}
