package haira

import (
	"strings"
	"unicode"
)

func StringLen(s string) int              { return len(s) }
func StringIsEmpty(s string) bool         { return len(s) == 0 }
func StringTrim(s string) string          { return strings.TrimSpace(s) }
func StringTrimLeft(s string) string      { return strings.TrimLeftFunc(s, unicode.IsSpace) }
func StringTrimRight(s string) string     { return strings.TrimRightFunc(s, unicode.IsSpace) }
func StringToUpper(s string) string       { return strings.ToUpper(s) }
func StringToLower(s string) string       { return strings.ToLower(s) }
func StringContains(s, sub string) bool   { return strings.Contains(s, sub) }
func StringStartsWith(s, pre string) bool { return strings.HasPrefix(s, pre) }
func StringEndsWith(s, suf string) bool   { return strings.HasSuffix(s, suf) }
func StringIndexOf(s, sub string) int     { return strings.Index(s, sub) }
func StringLastIndexOf(s, sub string) int { return strings.LastIndex(s, sub) }

func StringSplit(s, sep string) []any {
	parts := strings.Split(s, sep)
	result := make([]any, len(parts))
	for i, p := range parts {
		result[i] = p
	}
	return result
}

func StringJoin(items any, sep string) string { return Join(items, sep) }

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

func StringCharAt(s string, index int) string {
	if index < 0 || index >= len(s) {
		return ""
	}
	return string(s[index])
}

func StringReplace(s, old, newStr string) string    { return strings.Replace(s, old, newStr, 1) }
func StringReplaceAll(s, old, newStr string) string { return strings.ReplaceAll(s, old, newStr) }
func StringRepeat(s string, count int) string       { return strings.Repeat(s, count) }
