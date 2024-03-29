package rf

import r "reflect"

// Flags constituting the return value of `rf.Filter`.
// Unknown bits will be ignored.
const (
	// Don't visit self or descendants. Being zero, this is the default.
	VisNone = 0b_0000_0000

	// Visit self.
	VisSelf = 0b_0000_0001

	// Visit descendants.
	VisDesc = 0b_0000_0010

	// Visit both self and descendants.
	VisBoth = VisSelf | VisDesc

	// Same effect as `rf.VisBoth`. Provided for arithmetic.
	VisAll = 0b_1111_1111
)

/*
Tool for implementing efficient reflect-based deep walking. Determines if a
particular node should be visited during a walk, and how. This package provides
several filter implementations, such as filtering by type, by struct tag, or
combining other filters.

The return value of `rf.Filter.Visit` is a combination of two optional flags:
`rf.VisSelf` and `rf.VisDesc`. Flags are combined with bitwise OR. The
following combinations are known:

	return VisNone // Zero value / default.
	return VisSelf
	return VisDesc
	return VisSelf | VisDesc
	return VisBoth // Shortcut for the above.

If the flag `rf.VisDesc` is set, we attempt to generate an inner walker that
visits the descendants of the current node, such as the elements of a slice,
the fields of a struct, the value behind a pointer, or the value referenced by
an interface. Otherwise, we don't attempt to generate an inner walker.

If the flag `rf.VisSelf` is set, we generate a walker that invokes
`Visitor.Visit` on the current node. Otherwise the resulting walker will not
visit the current node, and may possibly be nil.

For technical reasons, all implementations of this interface must be values
rather than references. For example, filters provided by this package must be
used as values rather than pointers. The following is the CORRECT way to
construct filters:

	var filter rf.Filter = rf.And{
		TypeFilter[string]{},
		rf.TagFilter{`json`, `fieldName`},
	}

The following is the INCORRECT way to construct filters. Due to internal
validation, this will cause panics at runtime:

	var filter rf.Filter = &rf.And{
		&rf.TypeFilter[string]{},
		&rf.TagFilter{`json`, `fieldName`},
	}

See also:

	rf.Walker
	rf.Visitor
	rf.GetWalker
	rf.Walk
*/
type Filter interface {
	Visit(r.Type, r.StructField) byte
}

/*
Tool for implementing efficient reflect-based deep walking. The function
`rf.GetWalker` generates a walker for a SPECIFIC combination of parent type and
`rf.Filter`. The resulting walker is specialized for that combination, and
walks its input precisely and efficiently.

A walker takes a visitor and walks that visitor through a structure, invoking
the visitor on the nodes approved by the filter. See the interface
`rf.Visitor`.

For simplicity and efficiency, walkers generated by this package don't
additionally assert that the `reflect.Value` provided at the top level has the
same type for which the walker was generated. When using `rf.Walk` or
`rf.WalkFunc`, it always matches. Otherwise, it's your responsibility to pass
the matching input first to `rf.GetWalker`, then to `.Walk`. For simplicity,
walkers also assume that the visitor is non-nil.

This package currently does NOT support walking into maps, for two reasons:
unclear semantics and inefficiency. It's unclear if we should walk keys,
values, or key-value pairs, and how that affects the rest of the walking API.
Currently in Go 1.17, reflect-based map walking has horrible inefficiencies
which can't be amortized by 3rd party code. It would be a massive performance
footgun.

This package does support walking into interface values included into other
structures, but at an efficiency loss. In general, our walking mechanism relies
on statically determining what we should and shouldn't visit, which is possible
only with static types. Using interfaces as dynamically-typed containers of
unknown values defeats this design by forcing us to always visit each of them,
and may produce significant slowdowns. However, while visiting each interface
value is an unfortunate inefficiency, walking the value REFERENCED by an
interface is as precise and efficient as with static types.
*/
type Walker interface {
	Walk(r.Value, Visitor)
}

/*
Used by `rf.Walker` and `rf.Walk` to visit certain nodes of the given value. A
visitor can be an arbitrary value or a function; see `rf.VisitorFunc`.
*/
type Visitor interface {
	Visit(r.Value, r.StructField)
}

/*
Function type that implements `rf.Visitor`. Used by `rf.WalkFunc`. Converting a
func to an interface value is alloc-free.
*/
type VisitorFunc func(r.Value, r.StructField)

// Implement `rf.Visitor` by calling itself.
func (self VisitorFunc) Visit(val r.Value, field r.StructField) {
	if self == nil {
		return
	}
	self(val, field)
}

// Shortcut for calling `rf.Walk` with a visitor func.
func WalkFunc(val r.Value, fil Filter, vis VisitorFunc) {
	// `Walk` can't detect this case. We have to check it here.
	if vis == nil {
		return
	}
	Walk(val, fil, vis)
}

/*
Takes an arbitrary value and performs deep traversal, invoking the visitor for
each node allowed by the filter. Internally, uses `rf.GetWalker` to get or
create a walker specialized for this combination of type and filter. For each
type+filter combination, `rf.GetWalker` generates a specialized walker, caching
it for future calls. This approach allows MUCH more efficient walking.

If the input is zero/invalid/nil or the visitor is nil, this is a nop. For
slightly better performance, pass a pointer to reduce copying.

See also:

	rf.Walker
	rf.Filter
	rf.Visitor
	rf.GetWalker
*/
func Walk(val r.Value, fil Filter, vis Visitor) {
	if vis == nil {
		return
	}

	wal := GetWalker(ValueType(val), fil)
	if wal == nil {
		return
	}

	wal.Walk(val, vis)
}

/*
Shortcut for `rf.Walk` on the given value, which must be either a valid pointer
or nil. If the value is nil, this is a nop. Requiring a pointer is useful for
both efficiency and correctness. Even if the walker doesn't modify anything,
passing a pointer reduces copying. If the walker does modify walked values, and
you try to walk a non-pointer, you will get uninformative panics from
the "reflect" package. This function validates the inputs early, making it
easier to catch such bugs.
*/
func WalkPtr(val any, fil Filter, vis Visitor) {
	if val == nil {
		return
	}
	Walk(ValueDeref(ValidateValueKind(r.ValueOf(val), r.Ptr)), fil, vis)
}

// Shortcut for calling `rf.WalkPtr` with a visitor func.
func WalkPtrFunc(val any, fil Filter, vis VisitorFunc) {
	if val == nil {
		return
	}

	// Validate before early return.
	tar := ValueDeref(ValidateValueKind(r.ValueOf(val), r.Ptr))

	// `Walk` can't detect this case. We have to check it here.
	if vis == nil {
		return
	}

	Walk(tar, fil, vis)
}

/*
Returns an `rf.Walker` for the given type with the given filter. Uses caching to
avoid generating a walker more than once. Future calls with the same inputs
will return the same walker instance. Returns nil if for this combination of
type and filter, nothing will be visited. A nil filter is equivalent to a
filter that always returns false, resulting in a nil walker.
*/
func GetWalker(typ r.Type, fil Filter) Walker {
	if typ == nil || fil == nil {
		return nil
	}
	return walkerCacheStatic.getOrMake(typ, fil)
}

/*
Shortcut for `rf.TrawlWith` without an additional filter. Takes an arbitrary
source value and a pointer to an output slice. Walks the source value,
appending all non-zero values of the matching type to the given slice.
*/
func Trawl[Src any, Out ~[]Elem, Elem any](src *Src, out *Out) {
	TrawlWith(src, out, nil)
}

/*
Shortcut for using `rf.Appender` and `rf.Walk` to trawl the provided "source"
value to collect all non-zero values of a specific type into an "output" slice.
The source value may be of arbitrary type. The output must be a non-nil pointer
to a slice. The additional filter is optional.
*/
func TrawlWith[Src any, Out ~[]Elem, Elem any](src *Src, out *Out, fil Filter) {
	if src == nil || out == nil {
		return
	}

	/**
	The unsafe cast is correct and safe. Workaround for Go limitations.
	The following is equivalent and should work, but does not compile:

		appender := (*Appender[Elem])(out)
	*/
	appender := cast[*Appender[Elem]](out)

	filter := MaybeAnd(appender.Filter(), fil)
	Walk(r.ValueOf(src), filter, appender)
}

// Implementation of `rf.Filter` that always returns `rf.VisSelf`.
type Self struct{}

// Implement `rf.Filter`.
func (Self) Visit(r.Type, r.StructField) byte { return VisSelf }

// Implementation of `rf.Filter` that always returns `rf.VisDesc`.
type Desc struct{}

// Implement `rf.Filter`.
func (Desc) Visit(r.Type, r.StructField) byte { return VisDesc }

// Implementation of `rf.Filter` that always returns `rf.VisBoth`.
type Both struct{}

// Implement `rf.Filter`.
func (Both) Visit(r.Type, r.StructField) byte { return VisBoth }

// Implementation of `rf.Filter` that always returns `rf.VisAll`.
type All struct{}

// Implement `rf.Filter`.
func (All) Visit(r.Type, r.StructField) byte { return VisAll }

/*
Implementation of `rf.Filter` that allows to visit values of this specific type.
If the type is nil, this won't visit anything. The type may be either concrete
or an interface. It also allows to visit descendants.
*/
type TypeFilter[_ any] struct{}

// Implement `rf.Filter`.
func (TypeFilter[A]) Visit(typ r.Type, _ r.StructField) byte {
	if typ == Type[A]() {
		return VisBoth
	}
	return VisDesc
}

/*
Implementation of `rf.Filter` that allows to visit values of the given
`reflect.Kind`. If the kind is `reflect.Invalid`, this won't visit anything.

Untested.
*/
type KindFilter r.Kind

// Implement `rf.Filter`.
func (self KindFilter) Visit(typ r.Type, _ r.StructField) byte {
	if r.Kind(self) == typ.Kind() {
		return VisBoth
	}
	return VisDesc
}

/*
Implementation of `rf.Filter` that allows to visit values whose types implement
the given interface BY POINTER. If the type is nil, this won't visit anything.
The type must represent an interface, otherwise this will panic. The visitor
must explicitly take value address:

	func visit(val r.Value, _ r.StructField) {
		val.Addr().Interface().(SomeInterface).SomeMethod()
	}
*/
type IfaceFilter[_ any] struct{}

// Implement `rf.Filter`.
func (IfaceFilter[A]) Visit(typ r.Type, _ r.StructField) byte {
	return ifaceVisit(typ, Type[A](), VisBoth)
}

/*
Like `rf.IfaceFilter`, but visits either self or descendants, not both. In other
words, once it finds a node that implements the given interface (by pointer),
it allows to visit that node and stops there, without walking its descendants.
*/
type ShallowIfaceFilter[_ any] struct{}

// Implement `rf.Filter`.
func (ShallowIfaceFilter[A]) Visit(typ r.Type, _ r.StructField) byte {
	return ifaceVisit(typ, Type[A](), VisSelf)
}

/*
Implementation of `rf.Filter` that allows to visit values whose struct tag has a
specific tag with a specific value, such as tag "json" with value "-". It also
allows to visit descendants.

Known limitation: can't differentiate empty tag from missing tag.
*/
type TagFilter [2]string

// Implement `rf.Filter`.
func (self TagFilter) Visit(_ r.Type, field r.StructField) byte {
	key, val := self[0], self[1]
	if key != `` && field.Tag.Get(key) == val {
		return VisBoth
	}
	return VisDesc
}

/*
Implementation of `rf.Filter` that inverts the "self" bit of the inner filter,
without changing the other flags. If the inner filter is nil, this always
returns `rf.VisNone`.
*/
type InvertSelf [1]Filter

// Implement `rf.Filter`.
func (self InvertSelf) Visit(typ r.Type, field r.StructField) byte {
	if self[0] == nil {
		return VisNone
	}
	return self[0].Visit(typ, field) ^ VisSelf
}

/*
Micro-optimization for `rf.And`. If the input has NO non-nil filters, returns
nil. If the input has ONE non-nil filter, returns that filter, avoiding an
allocation of `rf.And{}`. Otherwise combines the filters via `rf.And`.
*/
func MaybeAnd(vals ...Filter) Filter {
	var out And
	slice := maybeCombineFilters(vals, out[:0])

	switch len(slice) {
	case 0:
		return nil
	case 1:
		return slice[0]
	default:
		return out
	}
}

/*
Implementation of `rf.Filter` that combines other filters, AND-ing their outputs
via `&`. Nil elements are ignored. If all elements are nil, the output is
automatically `VisNone`.
*/
type And [8]Filter

// Implement `rf.Filter`.
func (self And) Visit(typ r.Type, field r.StructField) (vis byte) {
	var found bool

	for _, val := range self {
		if val != nil {
			if !found {
				found = true
				vis = val.Visit(typ, field)
			} else {
				vis &= val.Visit(typ, field)
			}
		}
	}

	return
}

/*
Micro-optimization for `rf.Or`. If the input has NO non-nil filters, returns
nil. If the input has ONE non-nil filter, returns that filter, avoiding an
allocation of `rf.Or{}`. Otherwise combines the filters via `rf.Or`.
*/
func MaybeOr(vals ...Filter) Filter {
	var out Or
	slice := maybeCombineFilters(vals, out[:0])

	switch len(slice) {
	case 0:
		return nil
	case 1:
		return slice[0]
	default:
		return out
	}
}

/*
Implementation of `rf.Filter` that combines other filters, OR-ing their outputs
via `|`. Nil elements are ignored. If all elements are nil, the output is
automatically `VisNone`.
*/
type Or [8]Filter

// Implement `rf.Filter`.
func (self Or) Visit(typ r.Type, field r.StructField) (vis byte) {
	for _, val := range self {
		if val != nil {
			vis |= val.Visit(typ, field)
		}
	}
	return
}

// No-op implementation of both `rf.Visitor` that does nothing upon visit.
type Nop struct{}

// Implement `rf.Visitor`.
func (Nop) Visit(r.Value, r.StructField) {}

// Implements `rf.Visitor` by appending visited non-zero elements.
type Appender[A any] []A

/*
Implement `rf.Visitor` by appending the input value to the inner slice, if the
value is non-zero.
*/
func (self *Appender[A]) Visit(val r.Value, _ r.StructField) {
	if self != nil && !val.IsZero() {
		if val.CanAddr() {
			*self = append(*self, *val.Addr().Interface().(*A))
		} else {
			*self = append(*self, val.Interface().(A))
		}
	}
}

/*
Returns a filter that allows to visit only values suitable to be elements of the
slice held by the appender.
*/
func (self Appender[A]) Filter() Filter { return TypeFilter[A]{} }
