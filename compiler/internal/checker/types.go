// Package checker implements type inference and type checking for Haira.
package checker

import (
	"fmt"
	"strings"
)

// Type represents a Haira type in the checker.
type Type interface {
	typeString() string
	String() string
}

// Primitive types

type IntType struct{}
type FloatType struct{}
type StringType struct{}
type BoolType struct{}
type AnyType struct{}
type VoidType struct{}
type ErrorType struct{}

func (IntType) typeString() string    { return "int" }
func (FloatType) typeString() string  { return "float" }
func (StringType) typeString() string { return "string" }
func (BoolType) typeString() string   { return "bool" }
func (AnyType) typeString() string    { return "any" }
func (VoidType) typeString() string   { return "void" }
func (ErrorType) typeString() string  { return "error" }

func (t IntType) String() string    { return t.typeString() }
func (t FloatType) String() string  { return t.typeString() }
func (t StringType) String() string { return t.typeString() }
func (t BoolType) String() string   { return t.typeString() }
func (t AnyType) String() string    { return t.typeString() }
func (t VoidType) String() string   { return t.typeString() }
func (t ErrorType) String() string  { return t.typeString() }

// Compound types

type ListType struct{ Elem Type }
type MapType struct{ Key, Value Type }

type StructType struct {
	Name   string
	Fields map[string]Type
}

type EnumType struct {
	Name          string
	Variants      []string
	VariantFields map[string]int // variant name -> number of fields
}

type FuncType struct {
	Params []Type
	Return Type
}

type OptionTypeC struct {
	Inner Type
}

type TupleTypeC struct {
	Elems []Type
}

type UnionTypeC struct {
	Variants []Type
}

func (l ListType) typeString() string {
	return fmt.Sprintf("[%s]", l.Elem.typeString())
}
func (m MapType) typeString() string {
	return fmt.Sprintf("{%s: %s}", m.Key.typeString(), m.Value.typeString())
}
func (s StructType) typeString() string {
	return s.Name
}
func (e EnumType) typeString() string {
	return e.Name
}
func (f FuncType) typeString() string {
	return "fn"
}
func (o OptionTypeC) typeString() string {
	return o.Inner.typeString() + "?"
}
func (t TupleTypeC) typeString() string {
	parts := make([]string, len(t.Elems))
	for i, e := range t.Elems {
		parts[i] = e.typeString()
	}
	return "(" + strings.Join(parts, ", ") + ")"
}
func (u UnionTypeC) typeString() string {
	parts := make([]string, len(u.Variants))
	for i, v := range u.Variants {
		parts[i] = v.typeString()
	}
	return strings.Join(parts, " | ")
}

func (l ListType) String() string    { return l.typeString() }
func (m MapType) String() string     { return m.typeString() }
func (s StructType) String() string  { return s.typeString() }
func (e EnumType) String() string    { return e.typeString() }
func (f FuncType) String() string    { return f.typeString() }
func (o OptionTypeC) String() string { return o.typeString() }
func (t TupleTypeC) String() string  { return t.typeString() }
func (u UnionTypeC) String() string  { return u.typeString() }

// TypeEquals checks structural equality of two types.
func TypeEquals(a, b Type) bool {
	if a == nil || b == nil {
		return false
	}
	// any matches everything
	if _, ok := a.(AnyType); ok {
		return true
	}
	if _, ok := b.(AnyType); ok {
		return true
	}
	switch at := a.(type) {
	case IntType:
		_, ok := b.(IntType)
		return ok
	case FloatType:
		_, ok := b.(FloatType)
		return ok
	case StringType:
		_, ok := b.(StringType)
		return ok
	case BoolType:
		_, ok := b.(BoolType)
		return ok
	case VoidType:
		_, ok := b.(VoidType)
		return ok
	case ErrorType:
		_, ok := b.(ErrorType)
		return ok
	case ListType:
		bt, ok := b.(ListType)
		return ok && TypeEquals(at.Elem, bt.Elem)
	case MapType:
		bt, ok := b.(MapType)
		return ok && TypeEquals(at.Key, bt.Key) && TypeEquals(at.Value, bt.Value)
	case StructType:
		bt, ok := b.(StructType)
		return ok && at.Name == bt.Name
	case EnumType:
		bt, ok := b.(EnumType)
		return ok && at.Name == bt.Name
	case FuncType:
		bt, ok := b.(FuncType)
		if !ok || len(at.Params) != len(bt.Params) {
			return false
		}
		for i := range at.Params {
			if !TypeEquals(at.Params[i], bt.Params[i]) {
				return false
			}
		}
		return TypeEquals(at.Return, bt.Return)
	case OptionTypeC:
		bt, ok := b.(OptionTypeC)
		return ok && TypeEquals(at.Inner, bt.Inner)
	case TupleTypeC:
		bt, ok := b.(TupleTypeC)
		if !ok || len(at.Elems) != len(bt.Elems) {
			return false
		}
		for i := range at.Elems {
			if !TypeEquals(at.Elems[i], bt.Elems[i]) {
				return false
			}
		}
		return true
	case UnionTypeC:
		bt, ok := b.(UnionTypeC)
		if !ok || len(at.Variants) != len(bt.Variants) {
			return false
		}
		for i := range at.Variants {
			if !TypeEquals(at.Variants[i], bt.Variants[i]) {
				return false
			}
		}
		return true
	}
	return false
}
