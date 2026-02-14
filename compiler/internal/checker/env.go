package checker

// Env is a scoped symbol table for type checking.
type Env struct {
	parent *Env
	vars   map[string]Type
	types  map[string]Type
	funcs  map[string]*FuncType
}

// NewEnv creates a root environment with stdlib pre-registered.
func NewEnv() *Env {
	env := &Env{
		vars:  make(map[string]Type),
		types: make(map[string]Type),
		funcs: make(map[string]*FuncType),
	}
	env.registerStdlib()
	return env
}

// Child creates a new scope with this env as parent.
func (e *Env) Child() *Env {
	return &Env{
		parent: e,
		vars:   make(map[string]Type),
		types:  make(map[string]Type),
		funcs:  make(map[string]*FuncType),
	}
}

// DefineVar defines a variable in the current scope.
func (e *Env) DefineVar(name string, ty Type) {
	e.vars[name] = ty
}

// LookupVar looks up a variable in the current scope and parents.
func (e *Env) LookupVar(name string) (Type, bool) {
	if ty, ok := e.vars[name]; ok {
		return ty, true
	}
	if e.parent != nil {
		return e.parent.LookupVar(name)
	}
	return nil, false
}

// DefineType registers a type definition (struct/enum).
func (e *Env) DefineType(name string, ty Type) {
	e.types[name] = ty
}

// LookupType looks up a type by name.
func (e *Env) LookupType(name string) (Type, bool) {
	if ty, ok := e.types[name]; ok {
		return ty, true
	}
	if e.parent != nil {
		return e.parent.LookupType(name)
	}
	return nil, false
}

// DefineFunc registers a function signature.
func (e *Env) DefineFunc(name string, fn *FuncType) {
	e.funcs[name] = fn
}

// LookupFunc looks up a function by name.
func (e *Env) LookupFunc(name string) (*FuncType, bool) {
	if fn, ok := e.funcs[name]; ok {
		return fn, true
	}
	if e.parent != nil {
		return e.parent.LookupFunc(name)
	}
	return nil, false
}

func (e *Env) registerStdlib() {
	// io module functions
	e.funcs["io.println"] = &FuncType{Params: []Type{AnyType{}}, Return: VoidType{}}
	e.funcs["io.print"] = &FuncType{Params: []Type{AnyType{}}, Return: VoidType{}}

	// http module functions
	respTuple := AnyType{} // (Response, error) — simplified for now
	e.funcs["http.get"] = &FuncType{Params: []Type{StringType{}}, Return: respTuple}
	e.funcs["http.post"] = &FuncType{Params: []Type{StringType{}, AnyType{}}, Return: respTuple}

	// Built-in functions
	e.funcs["env"] = &FuncType{Params: []Type{StringType{}}, Return: StringType{}}
	e.funcs["len"] = &FuncType{Params: []Type{AnyType{}}, Return: IntType{}}
	e.funcs["keys"] = &FuncType{Params: []Type{AnyType{}}, Return: ListType{Elem: StringType{}}}
	e.funcs["join"] = &FuncType{Params: []Type{AnyType{}, StringType{}}, Return: StringType{}}
}
