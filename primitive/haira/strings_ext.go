package haira

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

// StringSlugify converts a string to a URL-safe slug.
// e.g., "Hello World!.xlsx" → "hello-world-xlsx"
func StringSlugify(s string) string {
	s = strings.ToLower(s)
	s = slugRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// StringBasename returns the last element of a path.
// e.g., "/path/to/file.xlsx" → "file.xlsx"
func StringBasename(s string) string {
	return filepath.Base(s)
}

// StringStripExt removes the file extension from a string.
// e.g., "file.xlsx" → "file"
func StringStripExt(s string) string {
	ext := filepath.Ext(s)
	if ext == "" {
		return s
	}
	return s[:len(s)-len(ext)]
}

// StringExt returns the file extension including the dot.
// e.g., "file.xlsx" → ".xlsx"
func StringExt(s string) string {
	return filepath.Ext(s)
}

// StringPadLeft pads a string on the left to the given length.
// e.g., StringPadLeft("42", 10, "0") → "0000000042"
func StringPadLeft(s string, length int, pad string) string {
	if pad == "" {
		pad = " "
	}
	for len(s) < length {
		s = pad + s
	}
	if len(s) > length {
		s = s[len(s)-length:]
	}
	return s
}

// StringPadRight pads a string on the right to the given length.
func StringPadRight(s string, length int, pad string) string {
	if pad == "" {
		pad = " "
	}
	for len(s) < length {
		s = s + pad
	}
	if len(s) > length {
		s = s[:length]
	}
	return s
}

// StringTruncate truncates a string to maxLen, appending suffix if truncated.
// e.g., StringTruncate("long text here", 8, "...") → "long ..."
func StringTruncate(s string, maxLen int, suffix string) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 0 {
		return ""
	}
	if len(suffix) >= maxLen {
		return suffix[:maxLen]
	}
	return s[:maxLen-len(suffix)] + suffix
}

// StringExtractBetween extracts the first substring between two delimiters.
// e.g., StringExtractBetween(`key="value"`, `"`, `"`) → "value"
func StringExtractBetween(s, start, end string) string {
	i := strings.Index(s, start)
	if i < 0 {
		return ""
	}
	s = s[i+len(start):]
	j := strings.Index(s, end)
	if j < 0 {
		return ""
	}
	return s[:j]
}

// StringCapitalize capitalizes the first letter of a string.
func StringCapitalize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// StringTitle capitalizes the first letter of each word.
func StringTitle(s string) string {
	return strings.Title(s)
}

// StringReverse reverses a string.
func StringReverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// StringCount counts non-overlapping occurrences of substr in s.
func StringCount(s, substr string) int {
	return strings.Count(s, substr)
}

// StringLines splits a string into lines.
func StringLines(s string) []any {
	parts := strings.Split(s, "\n")
	result := make([]any, len(parts))
	for i, p := range parts {
		result[i] = p
	}
	return result
}

// StringWords splits a string into words (whitespace-separated).
func StringWords(s string) []any {
	parts := strings.Fields(s)
	result := make([]any, len(parts))
	for i, p := range parts {
		result[i] = p
	}
	return result
}

// StringShellEscape escapes a string for safe use as a shell argument.
// Wraps the value in single quotes and escapes embedded single quotes.
func StringShellEscape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// DetectLanguage returns the programming language name for a file extension.
func DetectLanguage(path string) string {
	idx := strings.LastIndex(path, ".")
	if idx < 0 {
		return "text"
	}
	switch strings.ToLower(path[idx+1:]) {
	case "go":
		return "go"
	case "ts", "tsx":
		return "typescript"
	case "js", "jsx":
		return "javascript"
	case "py":
		return "python"
	case "rs":
		return "rust"
	case "haira":
		return "haira"
	case "json":
		return "json"
	case "yaml", "yml":
		return "yaml"
	case "md":
		return "markdown"
	case "html":
		return "html"
	case "css":
		return "css"
	case "sh", "bash":
		return "bash"
	case "sql":
		return "sql"
	case "toml":
		return "toml"
	case "xml":
		return "xml"
	default:
		return "text"
	}
}
