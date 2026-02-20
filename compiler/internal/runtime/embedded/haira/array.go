package haira

import (
	"fmt"
	"sort"
)

func ArrayLen(arr any) int      { return Len(arr) }
func ArrayIsEmpty(arr any) bool { return Len(arr) == 0 }

func ArrayFirst(arr any) any {
	s := ToSlice(arr)
	if len(s) == 0 {
		return nil
	}
	return s[0]
}

func ArrayLast(arr any) any {
	s := ToSlice(arr)
	if len(s) == 0 {
		return nil
	}
	return s[len(s)-1]
}

func ArrayGet(arr any, index int) any {
	s := ToSlice(arr)
	if index < 0 || index >= len(s) {
		return nil
	}
	return s[index]
}

func ArrayPush(arr any, elem any) []any {
	return append(ToSlice(arr), elem)
}

func ArrayPop(arr any) ([]any, any) {
	s := ToSlice(arr)
	if len(s) == 0 {
		return s, nil
	}
	return s[:len(s)-1], s[len(s)-1]
}

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

func ArrayRemove(arr any, index int) []any {
	s := ToSlice(arr)
	if index < 0 || index >= len(s) {
		return s
	}
	return append(s[:index], s[index+1:]...)
}

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

func ArrayReverse(arr any) []any {
	s := ToSlice(arr)
	result := make([]any, len(s))
	for i, item := range s {
		result[len(s)-1-i] = item
	}
	return result
}

func ArrayConcat(a, b any) []any {
	sa := ToSlice(a)
	sb := ToSlice(b)
	result := make([]any, 0, len(sa)+len(sb))
	result = append(result, sa...)
	result = append(result, sb...)
	return result
}

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

func ArraySort(arr any) []any {
	s := ToSlice(arr)
	result := make([]any, len(s))
	copy(result, s)
	sort.Slice(result, func(i, j int) bool {
		return fmt.Sprintf("%v", result[i]) < fmt.Sprintf("%v", result[j])
	})
	return result
}

func ArrayJoin(arr any, sep string) string { return Join(arr, sep) }

func ArrayMap(arr any, fn func(any) any) []any {
	s := ToSlice(arr)
	result := make([]any, len(s))
	for i, item := range s {
		result[i] = fn(item)
	}
	return result
}

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

func ArrayReduce(arr any, fn func(any, any) any, initial any) any {
	s := ToSlice(arr)
	acc := initial
	for _, item := range s {
		acc = fn(acc, item)
	}
	return acc
}

func ArrayFind(arr any, fn func(any) any) any {
	s := ToSlice(arr)
	for _, item := range s {
		if isTruthy(fn(item)) {
			return item
		}
	}
	return nil
}

func ArrayFindIndex(arr any, fn func(any) any) int {
	s := ToSlice(arr)
	for i, item := range s {
		if isTruthy(fn(item)) {
			return i
		}
	}
	return -1
}

func ArraySortBy(arr any, fn func(any) any) []any {
	s := ToSlice(arr)
	result := make([]any, len(s))
	copy(result, s)
	sort.SliceStable(result, func(i, j int) bool {
		return fmt.Sprintf("%v", fn(result[i])) < fmt.Sprintf("%v", fn(result[j]))
	})
	return result
}

func ArrayEvery(arr any, fn func(any) any) bool {
	for _, item := range ToSlice(arr) {
		if !isTruthy(fn(item)) {
			return false
		}
	}
	return true
}

func ArraySome(arr any, fn func(any) any) bool {
	for _, item := range ToSlice(arr) {
		if isTruthy(fn(item)) {
			return true
		}
	}
	return false
}

func ArrayForEach(arr any, fn func(any) any) {
	for _, item := range ToSlice(arr) {
		fn(item)
	}
}

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
