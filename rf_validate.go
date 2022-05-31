package rf

import (
	"fmt"
	r "reflect"
)

/*
Ensures that the type has the required kind or panics with a descriptive error.
Returns the same type, allowing shorter code.
*/
func ValidateTypeKind(typ r.Type, exp r.Kind) r.Type {
	act := TypeKind(typ)

	if exp != act {
		panic(Err{
			`validating type kind`,
			fmt.Errorf(`expected kind %v, got type %v of kind %v`, exp, typ, act),
		})
	}

	return typ
}

/*
Ensures that the value has the required kind or panics with a descriptive error.
Returns the same value, allowing shorter code.
*/
func ValidateValueKind(val r.Value, exp r.Kind) r.Value {
	act := val.Kind()

	if exp != act {
		panic(Err{
			`validating value kind`,
			fmt.Errorf(`expected kind %v, got value %v of kind %v`, exp, val, act),
		})
	}

	return val
}

// Shortcut for `rf.ValidateTypeKind(typ, reflect.Func)`.
func ValidateTypeFunc(typ r.Type) r.Type { return ValidateTypeKind(typ, r.Func) }

// Shortcut for `rf.ValidateTypeKind(typ, reflect.Map)`.
func ValidateTypeMap(typ r.Type) r.Type { return ValidateTypeKind(typ, r.Map) }

// Shortcut for `rf.ValidateTypeKind(typ, reflect.Ptr)`.
func ValidateTypePtr(typ r.Type) r.Type { return ValidateTypeKind(typ, r.Ptr) }

// Shortcut for `rf.ValidateTypeKind(typ, reflect.Slice)`.
func ValidateTypeSlice(typ r.Type) r.Type { return ValidateTypeKind(typ, r.Slice) }

// Shortcut for `rf.ValidateTypeKind(typ, reflect.Struct)`.
func ValidateTypeStruct(typ r.Type) r.Type { return ValidateTypeKind(typ, r.Struct) }

// Shortcut for `rf.ValidateValueKind(val, reflect.Func)`.
func ValidateFunc(val r.Value) r.Value { return ValidateValueKind(val, r.Func) }

// Shortcut for `rf.ValidateValueKind(val, reflect.Map)`.
func ValidateMap(val r.Value) r.Value { return ValidateValueKind(val, r.Map) }

// Shortcut for `rf.ValidateValueKind(val, reflect.Slice)`.
func ValidateSlice(val r.Value) r.Value { return ValidateValueKind(val, r.Slice) }

// Shortcut for `rf.ValidateValueKind(val, reflect.Struct)`.
func ValidateStruct(val r.Value) r.Value { return ValidateValueKind(val, r.Struct) }

// Similar to `rf.ValidateValueKind(val, reflect.Ptr)`, but also ensures that
// the pointer is non-nil.
func ValidatePtr(val r.Value) r.Value {
	ValidateValueKind(val, r.Ptr)

	if val.IsNil() {
		panic(Err{
			`validating pointer`,
			fmt.Errorf(`expected non-nil pointer %T, got nil`, val),
		})
	}

	return val
}

// Shortcut for `rf.ValidateFunc(reflect.ValueOf(val))`.
func ValidFunc(val any) r.Value { return ValidateFunc(r.ValueOf(val)) }

// Shortcut for `rf.ValidateMap(reflect.ValueOf(val))`.
func ValidMap(val any) r.Value { return ValidateMap(r.ValueOf(val)) }

// Shortcut for `rf.ValidateSlice(reflect.ValueOf(val))`.
func ValidSlice(val any) r.Value { return ValidateSlice(r.ValueOf(val)) }

// Shortcut for `rf.ValidateStruct(reflect.ValueOf(val))`.
func ValidStruct(val any) r.Value { return ValidateStruct(r.ValueOf(val)) }

// Shortcut for `rf.ValidatePtr(reflect.ValueOf(val))`.
func ValidPtr(val any) r.Value { return ValidatePtr(r.ValueOf(val)) }

// Shortcut for `rf.ValidateTypeFunc(reflect.TypeOf(val))`.
func ValidTypeFunc(val any) r.Type { return ValidateTypeFunc(r.TypeOf(val)) }

// Shortcut for `rf.ValidateTypeMap(reflect.TypeOf(val))`.
func ValidTypeMap(val any) r.Type { return ValidateTypeMap(r.TypeOf(val)) }

// Shortcut for `rf.ValidateTypeSlice(reflect.TypeOf(val))`.
func ValidTypeSlice(val any) r.Type { return ValidateTypeSlice(r.TypeOf(val)) }

// Shortcut for `rf.ValidateTypeStruct(reflect.TypeOf(val))`.
func ValidTypeStruct(val any) r.Type { return ValidateTypeStruct(r.TypeOf(val)) }

// Shortcut for `rf.ValidateTypePtr(reflect.TypeOf(val))`.
func ValidTypePtr(val any) r.Type { return ValidateTypePtr(r.TypeOf(val)) }

/*
Ensures that the value is a non-nil pointer where the underlying type has the
required kind, or panics with a descriptive error. Returns the same value,
allowing shorter code. Supports pointers of any depth: `*T`, `**T`, etc. Known
limitation: only the outermost pointer is required to be non-nil. Inner
pointers may be nil.
*/
func ValidatePtrToKind(val r.Value, exp r.Kind) r.Value {
	ValidatePtr(val)

	// Deep deref allows pointers of any depth: `*T`, `**T`, etc.
	act := TypeKind(TypeDeref(val.Type()))

	if exp != act {
		panic(Err{
			`validating pointer`,
			fmt.Errorf(`expected pointer to kind %v, got %v`, exp, val.Type()),
		})
	}

	return val
}

/*
Shortcut for `rf.ValidatePtrToKind(reflect.ValueOf(val))`. Converts the input to
`reflect.Value`, ensures that it's a non-nil pointer where the inner type has
the required kind, and returns the resulting `reflect.Value`.
*/
func ValidPtrToKind(val any, exp r.Kind) r.Value {
	return ValidatePtrToKind(r.ValueOf(val), exp)
}

/*
Ensures that the value is a slice where the element type has the required kind,
or panics with a descriptive error. Returns the same value, allowing shorter
code. Doesn't automatically dereference the input.
*/
func ValidateSliceOfKind(val r.Value, exp r.Kind) r.Value {
	ValidateSlice(val)

	act := TypeKind(val.Type().Elem())

	if exp != act {
		panic(Err{
			`validating slice`,
			fmt.Errorf(`expected slice of kind %v, got %v`, exp, val.Type()),
		})
	}

	return val
}

/*
Shortcut for `rf.ValidateSliceOfKind(reflect.ValueOf(val))`. Converts the input
to `reflect.Value`, ensures that it's a slice where the element type has the
required kind, and returns the resulting `reflect.Value`.
*/
func ValidSliceOfKind(val any, exp r.Kind) r.Value {
	return ValidateSliceOfKind(r.ValueOf(val), exp)
}

/*
Ensures that the value is a slice where the element type has the required type,
or panics with a descriptive error. Returns the same value, allowing shorter
code. Doesn't automatically dereference the input.
*/
func ValidateSliceOf(val r.Value, exp r.Type) r.Value {
	ValidateSlice(val)

	act := val.Type().Elem()

	if exp != act {
		panic(Err{
			`validating slice`,
			fmt.Errorf(`expected slice of type %v, got slice of type %v`, exp, act),
		})
	}

	return val
}

/*
Shortcut for `rf.ValidateSliceOf(reflect.ValueOf(val))`. Converts the input to
`reflect.Value`, ensures that it's a slice with the required element type, and
returns the resulting `reflect.Value`.
*/
func ValidSliceOf(val any, exp r.Type) r.Value {
	return ValidateSliceOf(r.ValueOf(val), exp)
}

/*
Takes a func type and ensures that it has the required count of input
parameters, or panics with a descriptive error.
*/
func ValidateFuncNumIn(typ r.Type, exp int) {
	if exp != typ.NumIn() {
		panic(Err{
			`validating func type`,
			fmt.Errorf(`expected func type with %v input parameters, found type %v`, exp, typ),
		})
	}
}

/*
Takes a func type and ensures that it has the required count of output
parameters, or panics with a descriptive error.
*/
func ValidateFuncNumOut(typ r.Type, exp int) {
	if exp != typ.NumOut() {
		panic(Err{
			`validating func type`,
			fmt.Errorf(`expected func type with %v output parameters, found type %v`, exp, typ),
		})
	}
}

/*
Takes a func type and ensures that its input parameters exactly match the count
and types provided to this function, or panics with a descriptive error. Among
the provided parameter types, nil serves as a wildcard that matches any type.
*/
func ValidateFuncIn(typ r.Type, params ...r.Type) {
	ValidateTypeFunc(typ)
	ValidateFuncNumIn(typ, len(params))

	for i, param := range params {
		if param != nil && param != typ.In(i) {
			panic(Err{
				`validating func type`,
				fmt.Errorf(`expected func type with input types %v, found type %v`, params, typ),
			})
		}
	}
}

/*
Takes a func type and ensures that its return parameters exactly match the count
and types provided to this function, or panics with a descriptive error. Among
the provided parameter types, nil serves as a wildcard that matches any type.
*/
func ValidateFuncOut(typ r.Type, params ...r.Type) {
	ValidateTypeFunc(typ)
	ValidateFuncNumOut(typ, len(params))

	for i, param := range params {
		if param != nil && param != typ.Out(i) {
			panic(Err{
				`validating func type`,
				fmt.Errorf(`expected func type with output types %v, found type %v`, params, typ),
			})
		}
	}
}

/*
Ensures that the given value either directly or indirectly (through any number
of arbitrarily-nested pointer types) contains a type of the provided kind, and
returns its dereferenced value. If any intermediary pointer is nil, the
returned value is invalid.
*/
func DerefWithKind(src any, kind r.Kind) r.Value {
	val := r.ValueOf(src)
	ValidateTypeKind(TypeDeref(ValueType(val)), kind)
	return ValueDeref(val)
}

// Shortcut for `rf.DerefWithKind(val, reflect.Func)`.
func DerefFunc(val any) r.Value { return DerefWithKind(val, r.Func) }

// Shortcut for `rf.DerefWithKind(val, reflect.Map)`.
func DerefMap(val any) r.Value { return DerefWithKind(val, r.Map) }

// Shortcut for `rf.DerefWithKind(val, reflect.Slice)`.
func DerefSlice(val any) r.Value { return DerefWithKind(val, r.Slice) }

// Shortcut for `rf.DerefWithKind(val, reflect.Struct)`.
func DerefStruct(val any) r.Value { return DerefWithKind(val, r.Struct) }
