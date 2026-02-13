// Package haira provides the Haira standard library runtime for Go.
package haira

import "fmt"

// Println prints values followed by a newline.
func Println(args ...any) {
	fmt.Println(args...)
}

// Print prints values without a trailing newline.
func Print(args ...any) {
	fmt.Print(args...)
}
