// Package ast defines AST node types for Haira.
package ast

// Span represents a range in the source code.
type Span struct {
	Start int // byte offset start (inclusive)
	End   int // byte offset end (exclusive)
}

// Merge returns a span covering both spans.
func (s Span) Merge(other Span) Span {
	start := s.Start
	if other.Start < start {
		start = other.Start
	}
	end := s.End
	if other.End > end {
		end = other.End
	}
	return Span{Start: start, End: end}
}

// Spanned wraps a value with source location information.
type Spanned[T any] struct {
	Node T
	Span Span
}
