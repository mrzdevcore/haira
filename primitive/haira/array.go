package haira

import (
	"fmt"
	"sort"
)

// ArrayLen returns the length of an array.
func ArrayLen(arr any) int {
	return Len(arr)
}

// ArrayIsEmpty returns true if the array is empty.
func ArrayIsEmpty(arr any) bool {
	return Len(arr) == 0
}

// ArrayFirst returns the first element or nil.
func ArrayFirst(arr any) any {
	s := ToSlice(arr)
	if len(s) == 0 {
		return nil
	}
	return s[0]
}

// ArrayLast returns the last element or nil.
func ArrayLast(arr any) any {
	s := ToSlice(arr)
	if len(s) == 0 {
		return nil
	}
	return s[len(s)-1]
}

// ArrayGet returns the element at index, or nil if out of bounds.
func ArrayGet(arr any, index int) any {
	s := ToSlice(arr)
	if index < 0 || index >= len(s) {
		return nil
	}
	return s[index]
}

// ArrayPush appends an element to the array and returns the new array.
func ArrayPush(arr any, elem any) []any {
	s := ToSlice(arr)
	return append(s, elem)
}

// ArrayPop removes and returns the last element. Returns (new_array, popped_element).
func ArrayPop(arr any) ([]any, any) {
	s := ToSlice(arr)
	if len(s) == 0 {
		return s, nil
	}
	return s[:len(s)-1], s[len(s)-1]
}

// ArrayInsert inserts an element at the given index.
func ArrayInsert(arr any, index int, elem any) []any {
	s := ToSlice(arr)
	if index < 0 {
		index = 0
	}
	if index >= len(s) {
		return append(s, elem)
	}
	s = append(s, nil)
	copy(s[index+1:], s[index:])
	s[index] = elem
	return s
}

// ArrayRemove removes the element at the given index.
func ArrayRemove(arr any, index int) []any {
	s := ToSlice(arr)
	if index < 0 || index >= len(s) {
		return s
	}
	return append(s[:index], s[index+1:]...)
}

// ArraySlice returns a sub-array from start to end (exclusive).
func ArraySlice(arr any, start, end int) []any {
	s := ToSlice(arr)
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	if start >= end {
		return []any{}
	}
	result := make([]any, end-start)
	copy(result, s[start:end])
	return result
}

// ArrayTake returns the first n elements.
func ArrayTake(arr any, n int) []any {
	s := ToSlice(arr)
	if n > len(s) {
		n = len(s)
	}
	if n < 0 {
		n = 0
	}
	result := make([]any, n)
	copy(result, s[:n])
	return result
}

// ArrayDrop returns the array without the first n elements.
func ArrayDrop(arr any, n int) []any {
	s := ToSlice(arr)
	if n > len(s) {
		n = len(s)
	}
	if n < 0 {
		n = 0
	}
	result := make([]any, len(s)-n)
	copy(result, s[n:])
	return result
}

// ArrayContains returns true if the array contains the element.
func ArrayContains(arr any, elem any) bool {
	s := ToSlice(arr)
	target := fmt.Sprintf("%v", elem)
	for _, item := range s {
		if fmt.Sprintf("%v", item) == target {
			return true
		}
	}
	return false
}

// ArrayIndexOf returns the index of the first occurrence, or -1.
func ArrayIndexOf(arr any, elem any) int {
	s := ToSlice(arr)
	target := fmt.Sprintf("%v", elem)
	for i, item := range s {
		if fmt.Sprintf("%v", item) == target {
			return i
		}
	}
	return -1
}

// ArrayReverse returns a new array with elements in reverse order.
func ArrayReverse(arr any) []any {
	s := ToSlice(arr)
	result := make([]any, len(s))
	for i, item := range s {
		result[len(s)-1-i] = item
	}
	return result
}

// ArrayConcat concatenates two arrays.
func ArrayConcat(a, b any) []any {
	sa := ToSlice(a)
	sb := ToSlice(b)
	result := make([]any, 0, len(sa)+len(sb))
	result = append(result, sa...)
	result = append(result, sb...)
	return result
}

// ArrayFlatten flattens one level of nesting.
func ArrayFlatten(arr any) []any {
	s := ToSlice(arr)
	result := make([]any, 0)
	for _, item := range s {
		if inner := ToSlice(item); inner != nil {
			result = append(result, inner...)
		} else {
			result = append(result, item)
		}
	}
	return result
}

// ArrayUnique returns a new array with duplicates removed.
func ArrayUnique(arr any) []any {
	s := ToSlice(arr)
	seen := make(map[string]bool)
	result := make([]any, 0)
	for _, item := range s {
		key := fmt.Sprintf("%v", item)
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}
	return result
}

// ArraySort sorts an array of strings or numbers in ascending order.
func ArraySort(arr any) []any {
	s := ToSlice(arr)
	result := make([]any, len(s))
	copy(result, s)
	sort.Slice(result, func(i, j int) bool {
		return fmt.Sprintf("%v", result[i]) < fmt.Sprintf("%v", result[j])
	})
	return result
}

// ArrayJoin joins array elements into a string with a separator.
func ArrayJoin(arr any, sep string) string {
	return Join(arr, sep)
}

// ArrayMap applies a function to each element and returns a new array.
func ArrayMap(arr any, fn func(any) any) []any {
	s := ToSlice(arr)
	result := make([]any, len(s))
	for i, item := range s {
		result[i] = fn(item)
	}
	return result
}

// ArrayFilter returns elements for which fn returns a truthy value.
func ArrayFilter(arr any, fn func(any) any) []any {
	s := ToSlice(arr)
	var result []any
	for _, item := range s {
		if isTruthy(fn(item)) {
			result = append(result, item)
		}
	}
	if result == nil {
		return []any{}
	}
	return result
}

// ArrayReduce reduces an array to a single value using an accumulator function.
func ArrayReduce(arr any, fn func(any, any) any, initial any) any {
	s := ToSlice(arr)
	acc := initial
	for _, item := range s {
		acc = fn(acc, item)
	}
	return acc
}

// ArrayFind returns the first element for which fn returns truthy, or nil.
func ArrayFind(arr any, fn func(any) any) any {
	s := ToSlice(arr)
	for _, item := range s {
		if isTruthy(fn(item)) {
			return item
		}
	}
	return nil
}

// ArrayFindIndex returns the index of the first element for which fn returns truthy, or -1.
func ArrayFindIndex(arr any, fn func(any) any) int {
	s := ToSlice(arr)
	for i, item := range s {
		if isTruthy(fn(item)) {
			return i
		}
	}
	return -1
}

// ArraySortBy sorts by a key function. The key function should return a comparable value.
func ArraySortBy(arr any, fn func(any) any) []any {
	s := ToSlice(arr)
	result := make([]any, len(s))
	copy(result, s)
	sort.SliceStable(result, func(i, j int) bool {
		ki := fmt.Sprintf("%v", fn(result[i]))
		kj := fmt.Sprintf("%v", fn(result[j]))
		return ki < kj
	})
	return result
}

// ArrayEvery returns true if fn returns truthy for every element.
func ArrayEvery(arr any, fn func(any) any) bool {
	s := ToSlice(arr)
	for _, item := range s {
		if !isTruthy(fn(item)) {
			return false
		}
	}
	return true
}

// ArraySome returns true if fn returns truthy for at least one element.
func ArraySome(arr any, fn func(any) any) bool {
	s := ToSlice(arr)
	for _, item := range s {
		if isTruthy(fn(item)) {
			return true
		}
	}
	return false
}

// ArrayForEach calls fn for each element (side-effects only).
func ArrayForEach(arr any, fn func(any) any) {
	s := ToSlice(arr)
	for _, item := range s {
		fn(item)
	}
}

// ArrayFlatMap applies fn to each element and flattens the result by one level.
func ArrayFlatMap(arr any, fn func(any) any) []any {
	s := ToSlice(arr)
	var result []any
	for _, item := range s {
		mapped := fn(item)
		if inner := ToSlice(mapped); inner != nil {
			result = append(result, inner...)
		} else {
			result = append(result, mapped)
		}
	}
	if result == nil {
		return []any{}
	}
	return result
}

// Range generates a slice of integers from start to end.
// If inclusive is true, the range includes end (start..=end).
// Otherwise, it excludes end (start..end).
func Range(start, end any, inclusive bool) []any {
	s := toInt(start)
	e := toInt(end)
	if inclusive {
		e++
	}
	if s >= e {
		return []any{}
	}
	result := make([]any, 0, e-s)
	for i := s; i < e; i++ {
		result = append(result, i)
	}
	return result
}

func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

// isTruthy converts a value to a boolean for filter/find/every/some.
func isTruthy(v any) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case float64:
		return val != 0
	case string:
		return val != ""
	default:
		return true
	}
}
