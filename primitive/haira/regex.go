package haira

import (
	"fmt"
	"regexp"
)

// mustCompile compiles a regex pattern, panicking on error (caught by Haira's try/catch).
func mustCompile(pattern string) *regexp.Regexp {
	re, err := regexp.Compile(pattern)
	if err != nil {
		panic(fmt.Sprintf("invalid regex %q: %v", pattern, err))
	}
	return re
}

// RegexIsMatch returns true if the pattern matches the text.
// Panics on invalid pattern (caught by Haira's try/catch).
func RegexIsMatch(pattern, text string) bool {
	re := mustCompile(pattern)
	return re.MatchString(text)
}

// RegexFind returns the first match of the pattern in text, or empty string.
// Panics on invalid pattern (caught by Haira's try/catch).
func RegexFind(pattern, text string) string {
	re := mustCompile(pattern)
	return re.FindString(text)
}

// RegexFindAll returns all matches of the pattern in text.
// Panics on invalid pattern (caught by Haira's try/catch).
func RegexFindAll(pattern, text string) []any {
	re := mustCompile(pattern)
	matches := re.FindAllString(text, -1)
	result := make([]any, len(matches))
	for i, m := range matches {
		result[i] = m
	}
	return result
}

// RegexReplace replaces the first match of pattern in text with repl.
// Panics on invalid pattern (caught by Haira's try/catch).
func RegexReplace(pattern, text, repl string) string {
	re := mustCompile(pattern)
	loc := re.FindStringIndex(text)
	if loc == nil {
		return text
	}
	return text[:loc[0]] + re.ReplaceAllString(text[loc[0]:loc[1]], repl) + text[loc[1]:]
}

// RegexReplaceAll replaces all matches of pattern in text with repl.
// Panics on invalid pattern (caught by Haira's try/catch).
func RegexReplaceAll(pattern, text, repl string) string {
	re := mustCompile(pattern)
	return re.ReplaceAllString(text, repl)
}

// RegexCaptures returns the capture groups of the first match.
// Panics on invalid pattern (caught by Haira's try/catch).
func RegexCaptures(pattern, text string) []any {
	re := mustCompile(pattern)
	match := re.FindStringSubmatch(text)
	if match == nil {
		return nil
	}
	result := make([]any, len(match))
	for i, m := range match {
		result[i] = m
	}
	return result
}

// RegexSplit splits text by pattern.
// Panics on invalid pattern (caught by Haira's try/catch).
func RegexSplit(pattern, text string) []any {
	re := mustCompile(pattern)
	parts := re.Split(text, -1)
	result := make([]any, len(parts))
	for i, p := range parts {
		result[i] = p
	}
	return result
}
