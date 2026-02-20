package haira

import "regexp"

func RegexIsMatch(pattern, text string) bool {
	matched, _ := regexp.MatchString(pattern, text)
	return matched
}

func RegexFind(pattern, text string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}
	return re.FindString(text)
}

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

func RegexReplaceAll(pattern, text, repl string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return text
	}
	return re.ReplaceAllString(text, repl)
}

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
