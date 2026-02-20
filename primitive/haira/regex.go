package haira

import "regexp"

// RegexIsMatch returns true if the pattern matches the text.
func RegexIsMatch(pattern, text string) bool {
	matched, err := regexp.MatchString(pattern, text)
	if err != nil {
		return false
	}
	return matched
}

// RegexFind returns the first match of the pattern in text, or empty string.
func RegexFind(pattern, text string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}
	return re.FindString(text)
}

// RegexFindAll returns all matches of the pattern in text.
func RegexFindAll(pattern, text string) []any {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	matches := re.FindAllString(text, -1)
	result := make([]any, len(matches))
	for i, m := range matches {
		result[i] = m
	}
	return result
}

// RegexReplace replaces the first match of pattern in text with repl.
func RegexReplace(pattern, text, repl string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return text
	}
	loc := re.FindStringIndex(text)
	if loc == nil {
		return text
	}
	return text[:loc[0]] + re.ReplaceAllString(text[loc[0]:loc[1]], repl) + text[loc[1]:]
}

// RegexReplaceAll replaces all matches of pattern in text with repl.
func RegexReplaceAll(pattern, text, repl string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return text
	}
	return re.ReplaceAllString(text, repl)
}

// RegexCaptures returns the capture groups of the first match.
func RegexCaptures(pattern, text string) []any {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
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
func RegexSplit(pattern, text string) []any {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return []any{text}
	}
	parts := re.Split(text, -1)
	result := make([]any, len(parts))
	for i, p := range parts {
		result[i] = p
	}
	return result
}
