package haira

import (
	"strings"
	"unicode"
)

// StringLen returns the length of a string.
func StringLen(s string) int {
	return len(s)
}

// StringIsEmpty returns true if the string is empty.
func StringIsEmpty(s string) bool {
	return len(s) == 0
}

// StringTrim removes leading and trailing whitespace.
func StringTrim(s string) string {
	return strings.TrimSpace(s)
}

// StringTrimLeft removes leading whitespace.
func StringTrimLeft(s string) string {
	return strings.TrimLeftFunc(s, unicode.IsSpace)
}

// StringTrimRight removes trailing whitespace.
func StringTrimRight(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// StringToUpper converts a string to uppercase.
func StringToUpper(s string) string {
	return strings.ToUpper(s)
}

// StringToLower converts a string to lowercase.
func StringToLower(s string) string {
	return strings.ToLower(s)
}

// StringContains returns true if s contains substr.
func StringContains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// StringStartsWith returns true if s starts with prefix.
func StringStartsWith(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// StringEndsWith returns true if s ends with suffix.
func StringEndsWith(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// StringIndexOf returns the index of the first occurrence of substr, or -1.
func StringIndexOf(s, substr string) int {
	return strings.Index(s, substr)
}

// StringLastIndexOf returns the index of the last occurrence of substr, or -1.
func StringLastIndexOf(s, substr string) int {
	return strings.LastIndex(s, substr)
}

// StringSplit splits s by separator and returns a slice.
func StringSplit(s, sep string) []any {
	parts := strings.Split(s, sep)
	result := make([]any, len(parts))
	for i, p := range parts {
		result[i] = p
	}
	return result
}

// StringJoin joins a slice of strings with a separator.
func StringJoin(items any, sep string) string {
	return Join(items, sep)
}

// StringSubstring returns a substring from start to end (exclusive).
func StringSubstring(s string, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	if start >= end {
		return ""
	}
	return s[start:end]
}

// StringCharAt returns the character at the given index as a string.
func StringCharAt(s string, index int) string {
	if index < 0 || index >= len(s) {
		return ""
	}
	return string(s[index])
}

// StringReplace replaces the first occurrence of old with new.
func StringReplace(s, old, newStr string) string {
	return strings.Replace(s, old, newStr, 1)
}

// StringReplaceAll replaces all occurrences of old with new.
func StringReplaceAll(s, old, newStr string) string {
	return strings.ReplaceAll(s, old, newStr)
}

// StringRepeat repeats s count times.
func StringRepeat(s string, count int) string {
	return strings.Repeat(s, count)
}
