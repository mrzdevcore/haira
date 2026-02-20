package haira

import (
	"fmt"
	"os"
)

// LogPrint prints a log message with a level prefix.
func LogPrint(level string, message any) {
	switch level {
	case "error":
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", message)
	case "warn":
		fmt.Fprintf(os.Stderr, "[WARN] %v\n", message)
	default:
		fmt.Printf("[INFO] %v\n", message)
	}
}
