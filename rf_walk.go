package rf

import r "reflect"

/*
Tool for implementing efficient reflect-based deep walking. Determines if a
particular node should be visited during a walk. This package provides several
filter implementations, such as filtering by type, by struct tag, or combining
other filters.

For technical reasons, all implementations of this interface must be values
rather than references. For example, filters provided by this package must be
used as values rather than pointers. The following is the CORRECT way to
construct filters:

	var filter rf.Filter = rf.And{
		TypeFilterFor((*string)(nil)),
		rf.TagFilter{`json`, `fieldName`},
	}

The following is the INCORRECT way to construct filters. Due to internal
validation, this will cause panics at runtime:

	var filter rf.Filter = &rf.And{
		&rf.TypeFilter{rf.DerefType((*string)(nil))},
		&rf.TagFilter{`json`, `fieldName`},
	}

See also:

	rf.Walker
	rf.Visitor
	rf.GetWalker
	rf.Walk
*/
type Filter interface {
	ShouldVisit(r.Type, r.StructField) bool
}

/*
Tool for implementing efficient reflect-based deep walking. The function
`rf.GetWalker` generates a walker for a SPECIFIC combination of parent type and
`rf.Filter`. The resulting walker is specialized for that combination, and
walks its input precisely and efficiently.

For simplicity and efficiency reasons, walkers generated by this package don't
additionally assert that the provided `reflect.Value` has the same type for
which the walker is generated. When using `rf.Walk` or `rf.WalkFunc`, this is
handled for you. Otherwise, it's your responsibility to pass a value of the
same type. Walkers also assume that the visitor is non-nil.

This package currently does NOT support walking into maps, for two reasons:
unclear semantics and inefficiency. It's unclear if we should walk keys,
values, or key-value pairs, and how that affects the rest of the walking API.
Currently in Go 1.17, reflect-based map walking has horrible inefficiencies
which can't be amortized by 3rd party code. It would be a massive performance
footgun.

This package does support walking into `interface{}` values included into other
structures, but at an efficiency loss. In general, our walking mechanism relies
on statically determining what we should and shouldn't visit, which is possible
only with static types. The dynamic typing of `interface{}` defeats this design
by forcing us to always visit every `interface{}`, and may produce significant
slowdowns. However, while visiting each `interface{}` is an unfortunate
inefficiency, walking the value referenced by an `interface{}` is as precise
and efficient as with static types.
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
	if self != nil {
		self(val, field)
	}
}

// Shortcut for calling `rf.Walk` with a visitor func.
func WalkFunc(val r.Value, fil Filter, vis VisitorFunc) {
	if vis != nil {
		Walk(val, fil, vis)
	}
}

/*
Takes an arbitrary value and performs deep traversal, invoking the visitor for
each node allowed by the filter. Internally, uses `rf.GetWalker` to get or
create a walker specialized for this combination of type and filter. For each
type+filter combination, `rf.GetWalker` generates a specialized walker, caching
it for future calls. This approach allows MUCH more efficient walking.
*/
func Walk(val r.Value, fil Filter, vis Visitor) {
	if vis == nil {
		return
	}

	wal := GetWalker(ValueType(val), fil)
	if wal != nil {
		wal.Walk(val, vis)
	}
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
	validateFilter(fil)
	return walkerCacheStatic.get(typ, nil, r.StructField{}, fil)
}

/*
Shortcut for `rf.TrawlWith` without an additional filter. Takes an arbitrary
source value and a pointer to an output slice. Walks the source value,
appending all non-zero values of the matching type to the given slice.
*/
func Trawl(src, out interface{}) {
	TrawlWith(src, out, nil)
}

/*
Shortcut for using `rf.Appender` and `rf.Walk` to trawl the provided "source"
value to collect all non-zero values of a specific type into an "output" slice.
The source value may be of arbitrary type. The output must be a non-nil pointer
to a slice. The additional filter is optional.
*/
func TrawlWith(src, out interface{}, fil Filter) {
	appender := Appender{ValidPtrToKind(out, r.Slice).Elem()}
	filter := MaybeAnd(appender.Filter(), fil)
	Walk(r.ValueOf(src), filter, appender)
}

// Implementation of `rf.Filter` that always returns true.
type True struct{}

// Implement `rf.Filter`.
func (True) ShouldVisit(r.Type, r.StructField) bool { return true }

// Implementation of `rf.Filter` that always returns false. Useless because this
// is the default output for a nil filter. Provided only for symmetry with
// `rf.True`.
type False struct{}

// Implement `rf.Filter`.
func (False) ShouldVisit(r.Type, r.StructField) bool { return false }

/*
Returns a cached `rf.Filter` for the given type, avoiding an allocation caused
by converting `rf.TypeFilter` to an interface. This may actually cost slightly
more CPU cycles due to intermediary `sync.Map` usage. May not be worth it.
*/
func GetTypeFilter(typ r.Type) Filter {
	return typeFilterCache.Get(typ).(Filter)
}

// Shortcut, same as `rf.GetTypeFilter(rf.DerefType(typ))`.
func TypeFilterFor(typ interface{}) Filter {
	return GetTypeFilter(DerefType(typ))
}

var typeFilterCache = Cache{Func: func(typ r.Type) interface{} { return TypeFilter{typ} }}

// Implementation of `rf.Filter` that allows to visit only values of this
// specific type.
type TypeFilter [1]r.Type

// Implement `rf.Filter`.
func (self TypeFilter) ShouldVisit(typ r.Type, _ r.StructField) bool {
	return self[0] == typ
}

// Implementation of `rf.Filter` that allows to visit only values with a
// specific value of a specific struct field tag.
type TagFilter [2]string

// Implement `rf.Filter`.
func (self TagFilter) ShouldVisit(_ r.Type, field r.StructField) bool {
	key, val := self[0], self[1]
	return key != `` && field.Tag.Get(key) == val
}

// Implementation of `rf.Filter` that inverts the filter provided to it.
type Not [1]Filter

// Implement `rf.Filter`.
func (self Not) ShouldVisit(typ r.Type, field r.StructField) bool {
	if self[0] != nil {
		return !self[0].ShouldVisit(typ, field)
	}
	return false
}

/*
Optimization for `rf.Or`. If the input has NO non-nil filters, this returns
nil, avoiding an allocation of `rf.Or`. If the input has ONE non-nil filter,
this returns that filter, avoiding an allocation of `rf.Or{}`. Otherwise it
combines the filters via `rf.Or`.
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
Implementation of `rf.Filter` that combines other filters, allowing to visit
nodes for which at least one of the filters returns true.
*/
type Or [8]Filter

// Implement `rf.Filter`.
func (self Or) ShouldVisit(typ r.Type, field r.StructField) bool {
	for _, val := range self {
		if val != nil && val.ShouldVisit(typ, field) {
			return true
		}
	}
	return false
}

/*
Optimization for `rf.And`. If the input has NO non-nil filters, this returns
nil, avoiding an allocation of `rf.And`. If the input has ONE non-nil filter,
this returns that filter, avoiding an allocation of `rf.And{}`. Otherwise it
combines the filters via `rf.And`.
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
Implementation of `rf.Filter` that combines other filters, allowing to visit
nodes for which all non-nil filters return true, and at least one filter is
non-nil. If all filters are nil, this returns false.
*/
type And [8]Filter

// Implement `rf.Filter`.
func (self And) ShouldVisit(typ r.Type, field r.StructField) bool {
	var found bool

	for _, val := range self {
		if val == nil {
			continue
		}
		if !val.ShouldVisit(typ, field) {
			return false
		}
		found = true
	}

	return found
}

// No-op implementation of both `rf.Filter` and `rf.Visitor` that doesn't visit
// anything and does nothing upon visit.
type Nop struct{}

// Implement `rf.Filter`.
func (Nop) ShouldVisit(r.Type, r.StructField) bool { return false }

// Implement `rf.Visitor`.
func (Nop) Visit(r.Value, r.StructField) {}

// Shortcut for making `rf.Appender`. The input must be a carrier of the element
// type, not the slice type.
func AppenderFor(typ interface{}) Appender {
	return Appender{r.New(SliceType(typ)).Elem()}
}

/*
Implementation of `rf.Visitor` for collecting non-zero values of a single type
into a slice. The inner value must be `reflect.Value` holding a slice. The
value must be settable. Use `rf.AppenderFor` to instantiate this correctly.
*/
type Appender [1]r.Value

// Implement `rf.Visitor` by appending the input value to the inner slice, if
// the value is non-zero.
func (self Appender) Visit(val r.Value, _ r.StructField) {
	if !val.IsZero() {
		self[0].Set(r.Append(self[0], val))
	}
}

// Returns a filter that allows to visit only values suitable to be elements of
// the slice held by the appender.
func (self Appender) Filter() Filter {
	return GetTypeFilter(self[0].Type().Elem())
}

// Shortcut for `self[0].Interface()` insulating the caller from implementation
// details.
func (self Appender) Interface() interface{} { return self[0].Interface() }