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
	any := AnyType{}
	str := StringType{}
	integer := IntType{}
	float := FloatType{}
	boolean := BoolType{}
	void := VoidType{}
	strList := ListType{Elem: str}
	anyList := ListType{Elem: any}

	// io module functions
	e.funcs["io.println"] = &FuncType{Params: []Type{any}, Return: void}
	e.funcs["io.print"] = &FuncType{Params: []Type{any}, Return: void}
	e.funcs["io.printf"] = &FuncType{Params: []Type{str, any}, Return: void}
	e.funcs["io.readln"] = &FuncType{Params: nil, Return: str}
	e.funcs["io.eprintln"] = &FuncType{Params: []Type{any}, Return: void}
	e.funcs["io.eprintf"] = &FuncType{Params: []Type{str, any}, Return: void}

	// http module functions
	respTuple := any // (Response, error) — simplified for now
	e.funcs["http.get"] = &FuncType{Params: []Type{str}, Return: respTuple}
	e.funcs["http.get_with_headers"] = &FuncType{Params: []Type{str, any}, Return: respTuple}
	e.funcs["http.post"] = &FuncType{Params: []Type{str, any}, Return: respTuple}
	e.funcs["http.post_with_headers"] = &FuncType{Params: []Type{str, any, any}, Return: respTuple}
	e.funcs["http.put"] = &FuncType{Params: []Type{str, any}, Return: respTuple}
	e.funcs["http.put_with_headers"] = &FuncType{Params: []Type{str, any, any}, Return: respTuple}
	e.funcs["http.delete"] = &FuncType{Params: []Type{str}, Return: respTuple}
	e.funcs["http.delete_with_headers"] = &FuncType{Params: []Type{str, any}, Return: respTuple}

	// json module functions
	e.funcs["json.marshal"] = &FuncType{Params: []Type{any}, Return: any}
	e.funcs["json.unmarshal"] = &FuncType{Params: []Type{str}, Return: any}
	e.funcs["json.marshal_pretty"] = &FuncType{Params: []Type{any}, Return: any}
	e.funcs["json.parse"] = &FuncType{Params: []Type{str}, Return: any}

	// string module functions
	e.funcs["string.len"] = &FuncType{Params: []Type{str}, Return: integer}
	e.funcs["string.is_empty"] = &FuncType{Params: []Type{str}, Return: boolean}
	e.funcs["string.trim"] = &FuncType{Params: []Type{str}, Return: str}
	e.funcs["string.trim_left"] = &FuncType{Params: []Type{str}, Return: str}
	e.funcs["string.trim_right"] = &FuncType{Params: []Type{str}, Return: str}
	e.funcs["string.to_upper"] = &FuncType{Params: []Type{str}, Return: str}
	e.funcs["string.to_lower"] = &FuncType{Params: []Type{str}, Return: str}
	e.funcs["string.contains"] = &FuncType{Params: []Type{str, str}, Return: boolean}
	e.funcs["string.starts_with"] = &FuncType{Params: []Type{str, str}, Return: boolean}
	e.funcs["string.ends_with"] = &FuncType{Params: []Type{str, str}, Return: boolean}
	e.funcs["string.index_of"] = &FuncType{Params: []Type{str, str}, Return: integer}
	e.funcs["string.last_index_of"] = &FuncType{Params: []Type{str, str}, Return: integer}
	e.funcs["string.split"] = &FuncType{Params: []Type{str, str}, Return: strList}
	e.funcs["string.join"] = &FuncType{Params: []Type{any, str}, Return: str}
	e.funcs["string.substring"] = &FuncType{Params: []Type{str, integer, integer}, Return: str}
	e.funcs["string.char_at"] = &FuncType{Params: []Type{str, integer}, Return: str}
	e.funcs["string.replace"] = &FuncType{Params: []Type{str, str, str}, Return: str}
	e.funcs["string.replace_all"] = &FuncType{Params: []Type{str, str, str}, Return: str}
	e.funcs["string.repeat"] = &FuncType{Params: []Type{str, integer}, Return: str}

	// regex module functions
	e.funcs["regex.is_match"] = &FuncType{Params: []Type{str, str}, Return: boolean}
	e.funcs["regex.find"] = &FuncType{Params: []Type{str, str}, Return: str}
	e.funcs["regex.find_all"] = &FuncType{Params: []Type{str, str}, Return: anyList}
	e.funcs["regex.replace"] = &FuncType{Params: []Type{str, str, str}, Return: str}
	e.funcs["regex.replace_all"] = &FuncType{Params: []Type{str, str, str}, Return: str}
	e.funcs["regex.captures"] = &FuncType{Params: []Type{str, str}, Return: anyList}
	e.funcs["regex.split"] = &FuncType{Params: []Type{str, str}, Return: anyList}

	// math module functions
	e.funcs["math.abs"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.min"] = &FuncType{Params: []Type{any, any}, Return: float}
	e.funcs["math.max"] = &FuncType{Params: []Type{any, any}, Return: float}
	e.funcs["math.clamp"] = &FuncType{Params: []Type{any, any, any}, Return: float}
	e.funcs["math.floor"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.ceil"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.round"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.trunc"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.pow"] = &FuncType{Params: []Type{any, any}, Return: float}
	e.funcs["math.sqrt"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.cbrt"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.exp"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.log"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.log10"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.log2"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.sin"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.cos"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.tan"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.asin"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.acos"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.atan"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["math.atan2"] = &FuncType{Params: []Type{any, any}, Return: float}
	e.funcs["math.random"] = &FuncType{Params: nil, Return: float}
	e.funcs["math.random_int"] = &FuncType{Params: []Type{any, any}, Return: integer}
	e.funcs["math.pi"] = &FuncType{Params: nil, Return: float}
	e.funcs["math.e"] = &FuncType{Params: nil, Return: float}

	// conv module functions
	e.funcs["conv.int_to_string"] = &FuncType{Params: []Type{any}, Return: str}
	e.funcs["conv.float_to_string"] = &FuncType{Params: []Type{any}, Return: str}
	e.funcs["conv.bool_to_string"] = &FuncType{Params: []Type{any}, Return: str}
	e.funcs["conv.string_to_int"] = &FuncType{Params: []Type{str}, Return: any}
	e.funcs["conv.string_to_float"] = &FuncType{Params: []Type{str}, Return: any}
	e.funcs["conv.string_to_bool"] = &FuncType{Params: []Type{str}, Return: any}
	e.funcs["conv.int_to_float"] = &FuncType{Params: []Type{any}, Return: float}
	e.funcs["conv.float_to_int"] = &FuncType{Params: []Type{any}, Return: integer}
	e.funcs["conv.int_to_hex"] = &FuncType{Params: []Type{any}, Return: str}
	e.funcs["conv.int_to_binary"] = &FuncType{Params: []Type{any}, Return: str}
	e.funcs["conv.int_to_octal"] = &FuncType{Params: []Type{any}, Return: str}
	e.funcs["conv.hex_to_int"] = &FuncType{Params: []Type{str}, Return: any}

	// array module functions
	e.funcs["array.len"] = &FuncType{Params: []Type{any}, Return: integer}
	e.funcs["array.is_empty"] = &FuncType{Params: []Type{any}, Return: boolean}
	e.funcs["array.first"] = &FuncType{Params: []Type{any}, Return: any}
	e.funcs["array.last"] = &FuncType{Params: []Type{any}, Return: any}
	e.funcs["array.get"] = &FuncType{Params: []Type{any, integer}, Return: any}
	e.funcs["array.push"] = &FuncType{Params: []Type{any, any}, Return: anyList}
	e.funcs["array.pop"] = &FuncType{Params: []Type{any}, Return: any}
	e.funcs["array.insert"] = &FuncType{Params: []Type{any, integer, any}, Return: anyList}
	e.funcs["array.remove"] = &FuncType{Params: []Type{any, integer}, Return: anyList}
	e.funcs["array.slice"] = &FuncType{Params: []Type{any, integer, integer}, Return: anyList}
	e.funcs["array.take"] = &FuncType{Params: []Type{any, integer}, Return: anyList}
	e.funcs["array.drop"] = &FuncType{Params: []Type{any, integer}, Return: anyList}
	e.funcs["array.contains"] = &FuncType{Params: []Type{any, any}, Return: boolean}
	e.funcs["array.index_of"] = &FuncType{Params: []Type{any, any}, Return: integer}
	e.funcs["array.reverse"] = &FuncType{Params: []Type{any}, Return: anyList}
	e.funcs["array.concat"] = &FuncType{Params: []Type{any, any}, Return: anyList}
	e.funcs["array.flatten"] = &FuncType{Params: []Type{any}, Return: anyList}
	e.funcs["array.unique"] = &FuncType{Params: []Type{any}, Return: anyList}
	e.funcs["array.sort"] = &FuncType{Params: []Type{any}, Return: anyList}
	e.funcs["array.join"] = &FuncType{Params: []Type{any, str}, Return: str}

	// map module functions
	e.funcs["map.len"] = &FuncType{Params: []Type{any}, Return: integer}
	e.funcs["map.is_empty"] = &FuncType{Params: []Type{any}, Return: boolean}
	e.funcs["map.get"] = &FuncType{Params: []Type{any, str}, Return: any}
	e.funcs["map.has"] = &FuncType{Params: []Type{any, str}, Return: boolean}
	e.funcs["map.set"] = &FuncType{Params: []Type{any, str, any}, Return: any}
	e.funcs["map.remove"] = &FuncType{Params: []Type{any, str}, Return: any}
	e.funcs["map.keys"] = &FuncType{Params: []Type{any}, Return: anyList}
	e.funcs["map.values"] = &FuncType{Params: []Type{any}, Return: anyList}
	e.funcs["map.entries"] = &FuncType{Params: []Type{any}, Return: anyList}
	e.funcs["map.merge"] = &FuncType{Params: []Type{any, any}, Return: any}
	e.funcs["map.contains_value"] = &FuncType{Params: []Type{any, any}, Return: boolean}

	// postgres module
	e.funcs["postgres.connect"] = &FuncType{Params: []Type{str}, Return: any}

	// slack module
	e.funcs["slack.send"] = &FuncType{Params: []Type{str, str, str}, Return: any}

	// excel module
	e.funcs["excel.open"] = &FuncType{Params: []Type{str}, Return: any}

	// time module
	e.funcs["time.sleep"] = &FuncType{Params: []Type{any}, Return: void}
	e.funcs["time.now"] = &FuncType{Params: nil, Return: str}
	e.funcs["time.slug"] = &FuncType{Params: nil, Return: str}

	// Built-in functions
	e.funcs["env"] = &FuncType{Params: []Type{str}, Return: str}
	e.funcs["len"] = &FuncType{Params: []Type{any}, Return: integer}
	e.funcs["keys"] = &FuncType{Params: []Type{any}, Return: strList}
	e.funcs["join"] = &FuncType{Params: []Type{any, str}, Return: str}
}
