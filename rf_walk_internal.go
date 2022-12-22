package rf

import (
	"fmt"
	r "reflect"
	"sync"
)

var (
	walkerCacheStatic walkerCache
	typeFilter        = r.TypeOf((*Filter)(nil)).Elem()
	typeType          = r.TypeOf((*r.Type)(nil)).Elem()
)

/**
Represents the return value of `Filter.Visit`. We use `byte` in the method's
signature and keep this type private due to the general principle that
interfaces should avoid concrete library types, and contain only built-in
types, standard library types, and other interfaces.
*/
type vis byte

func (self vis) self() bool { return (self & VisSelf) != 0 }
func (self vis) desc() bool { return (self & VisDesc) != 0 }

type walkerCache struct {
	sync.RWMutex
	Map map[walkRef]Walker
}

func (self *walkerCache) getOrMake(typ r.Type, fil Filter) Walker {
	var ref walkRef
	ref.Filter = fil
	ref.Type = typ
	return self.getOrMakeFor(ref)
}

func (self *walkerCache) getOrMakeFor(ref walkRef) Walker {
	val, ok := self.got(ref)
	if ok {
		return val
	}
	ref.validate()
	return self.set(ref, ref.makeWalker())
}

// The boolean is not redundant: we generate nil walkers for some keys.
func (self *walkerCache) got(ref walkRef) (Walker, bool) {
	self.RLock()
	defer self.RUnlock()
	val, ok := self.Map[ref]
	return val, ok
}

func (self *walkerCache) set(ref walkRef, val Walker) Walker {
	self.Lock()
	defer self.Unlock()

	if self.Map == nil {
		self.Map = map[walkRef]Walker{ref: val}
	} else {
		self.Map[ref] = val
	}
	return val
}

/**
Very similar to `reflect.StructField`, but provides more information, uses less
memory, and is usable in map keys such as `walkRef`. The tradeoff is that
generating `reflect.StructField` from this type has a measurable performance
cost, and should be done only during walker construction, not during walking.

We would prefer to simply use `reflect.StructField` here, but that type is not
suitable for map keys. The combination of `.Parent` and `.Index` is equivalent
and suitable for map keys.

When `.Parent` is nil, `.Index` must be zero. The complete zero value represents
a non-field.
*/
type fieldRef struct {
	Parent r.Type
	Index  int
}

/**
Caution: this should be called only during walker building, and never during
actual walking. Walkers such as `ifaceWalker` that use this type should avoid
calling this.
*/
func (self fieldRef) StructField() (_ r.StructField) {
	if self.Parent == nil {
		return
	}
	return self.Parent.Field(self.Index)
}

type filterRef struct {
	fieldRef
	Filter Filter
}

func (self filterRef) walkRef(typ r.Type) (out walkRef) {
	out.filterRef = self
	out.Type = typ
	return
}

/**
When `.fieldRef` is zero, this represents a top-level type, i.e. a type for
which some caller has explicitly requested a walker via some public API. In all
other cases, `.fieldRef` must be non-zero.
*/
type walkRef struct {
	filterRef
	Type r.Type
}

func (self walkRef) validate() { validateFilter(self.Filter) }

func (self walkRef) vis() (_ vis) {
	if self.Type != nil && self.Filter != nil {
		return vis(self.Filter.Visit(self.Type, self.StructField()))
	}
	return
}

func (self walkRef) makeWalker() Walker {
	var bui walkBui
	bui.walkRef = self
	return bui.makeWalker()
}

/**
The term "bui" is short for "builder". This wrapper allows us to detect cyclic
types and avoid infinite recursion / stack overflow when attempting to build a
walker for inner occurrences of the same outer type. The parent pointer refers
to an earlier stack frame, using the language's own stack for book keeping, and
avoiding the need for another data structure.

Currently, we simply skip inner occurrences of an outer type, generating a nil
walker. The outermost occurrence of any given type is walked as usual, but its
inner occurrences are not walked. This limitation is due to technical
difficulties, and we would like to lift it in the future.
*/
type walkBui struct {
	walkRef
	parent *walkBui
}

func (self *walkBui) isCyclic() bool {
	return self.parent != nil &&
		self.Type != nil &&
		self.parent.isTypePending(self.Type)
}

func (self *walkBui) isTypePending(typ r.Type) bool {
	return typ == self.Type ||
		(self.parent != nil && self.parent.isTypePending(typ))
}

func (self *walkBui) elem() (out walkBui) {
	out.walkRef = self.walkRef
	out.Type = self.Type.Elem()
	out.parent = self
	return
}

func (self *walkBui) field(index int) (out walkBui) {
	field := self.Type.Field(index)
	out.Type = field.Type
	out.Parent = self.Type
	out.Index = index
	out.Filter = self.Filter
	out.parent = self
	return
}

func (self *walkBui) makeWalker() Walker {
	if self.Type == nil || self.Filter == nil || self.isCyclic() {
		return nil
	}

	switch self.Type.Kind() {
	case r.Ptr:
		return self.makePtrWalker()
	case r.Array, r.Slice:
		return self.makeListWalker()
	case r.Struct:
		return self.makeStructWalker()
	case r.Interface:
		return self.makeIfaceWalker()
	default:
		return self.makeLeafWalker()
	}
}

func (self *walkBui) makePtrWalker() Walker {
	if self.vis().desc() {
		sub := self.elem()
		inner := sub.makeWalker()

		if inner != nil {
			return self.makeNodeWalker(ptrWalker{inner})
		}
	}

	return self.makeLeafWalker()
}

func (self *walkBui) makeListWalker() Walker {
	if self.vis().desc() {
		sub := self.elem()
		inner := sub.makeWalker()

		if inner != nil {
			return self.makeNodeWalker(listWalker{inner})
		}
	}

	return self.makeLeafWalker()
}

func (self *walkBui) makeStructWalker() Walker {
	if self.vis().desc() {
		var tar structWalker
		for ind := range Iter(self.Type.NumField()) {
			tar.maybeAppend(self.makeFieldIndexWalker(ind))
		}

		if len(tar) > 0 {
			return self.makeNodeWalker(tar)
		}
	}

	return self.makeLeafWalker()
}

// Note: returned walker may be invalid.
func (self *walkBui) makeFieldIndexWalker(index int) (out fieldIndexWalker) {
	field := self.Type.Field(index)
	if !IsFieldPublic(field) {
		return
	}

	sub := self.field(index)
	out.Index = index
	out.Inner = sub.makeWalker()
	return
}

func (self *walkBui) makeIfaceWalker() Walker {
	if self.vis().desc() {
		return self.makeNodeWalker(ifaceWalker(self.filterRef))
	}
	return self.makeLeafWalker()
}

func (self *walkBui) makeNodeWalker(inner Walker) Walker {
	if inner == nil {
		panic(errUselessNodeWalker)
	}

	if self.vis().self() {
		field := self.StructField()
		if isFieldValid(field) {
			return selfFieldWalker{field, inner}
		}
		return selfWalker{inner}
	}
	return inner
}

var errUselessNodeWalker = Err{
	`making node walker`,
	ErrStr(`internal violation: attempted to construct useless node walker without inner walker`),
}

func (self *walkBui) makeLeafWalker() Walker {
	if self.vis().self() {
		field := self.StructField()
		if isFieldValid(field) {
			return leafFieldWalker(field)
		}
		return leafWalker{}
	}
	return nil
}

type selfWalker [1]Walker

func (self selfWalker) Walk(val r.Value, vis Visitor) {
	vis.Visit(val, r.StructField{})
	self[0].Walk(val, vis)
}

type selfFieldWalker struct {
	Field r.StructField
	Inner Walker
}

func (self selfFieldWalker) Walk(val r.Value, vis Visitor) {
	vis.Visit(val, self.Field)
	self.Inner.Walk(val, vis)
}

type ptrWalker [1]Walker

func (self ptrWalker) Walk(val r.Value, vis Visitor) {
	if !val.IsNil() {
		self[0].Walk(val.Elem(), vis)
	}
}

type listWalker [1]Walker

func (self listWalker) Walk(val r.Value, vis Visitor) {
	for ind := range Iter(val.Len()) {
		self[0].Walk(val.Index(ind), vis)
	}
}

type structWalker []fieldIndexWalker

func (self structWalker) Walk(val r.Value, vis Visitor) {
	for _, walker := range self {
		walker.Walk(val, vis)
	}
}

func (self *structWalker) maybeAppend(val fieldIndexWalker) {
	if val.isValid() {
		*self = append(*self, val)
	}
}

/**
Implementation note: it may seem unintuitive that this walker does not store
`reflect.StructField`, while some other walkers do store it. That's due to the
signatures of our `Walker` and `Visitor` interfaces. The outer walker invokes
the inner walker without providing the field (see the `Walker` interface). The
innermost walker stores the field and provides it to the visitor.
*/
type fieldIndexWalker struct {
	Index int
	Inner Walker
}

/**
Does not nil-check `.Inner` because when `.Inner` is nil, this should be
excluded from `structWalker`. Having a nil inner walker would be a bug.
*/
func (self fieldIndexWalker) Walk(val r.Value, vis Visitor) {
	self.Inner.Walk(val.Field(self.Index), vis)
}

func (self fieldIndexWalker) isValid() bool { return self.Inner != nil }

type ifaceWalker filterRef

func (self ifaceWalker) Walk(val r.Value, vis Visitor) {
	if val.IsNil() {
		return
	}

	val = val.Elem()
	walker := walkerCacheStatic.getOrMakeFor(filterRef(self).walkRef(val.Type()))
	if walker != nil {
		walker.Walk(val, vis)
	}
}

type leafFieldWalker r.StructField

func (self leafFieldWalker) Walk(val r.Value, vis Visitor) {
	vis.Visit(val, r.StructField(self))
}

type leafWalker struct{}

func (leafWalker) Walk(val r.Value, vis Visitor) {
	vis.Visit(val, r.StructField{})
}

func validateFilter(src Filter) {
	validateFilterValue(src, r.ValueOf(src))
}

func validateFilterValue(src Filter, val r.Value) {
	switch val.Kind() {
	case r.Chan, r.Func, r.Map, r.Ptr, r.Slice, r.UnsafePointer:
		panic(errInvalidFilter(src, val))
	case r.Array:
		validateFilterArray(src, val)
	case r.Struct:
		validateFilterStruct(src, val)
	case r.Interface:
		switch val.Type() {
		case typeType:
		case typeFilter:
			if !val.IsNil() {
				validateFilterValue(src, val.Elem())
			}
		default:
			panic(errInvalidFilter(src, val))
		}
	}
}

func validateFilterArray(src Filter, val r.Value) {
	for ind := range Iter(val.Len()) {
		validateFilterValue(src, val.Index(ind))
	}
}

func validateFilterStruct(src Filter, val r.Value) {
	for ind := range Iter(val.NumField()) {
		validateFilterValue(src, val.Field(ind))
	}
}

func errInvalidFilter(src Filter, val r.Value) Err {
	return Err{
		`validating walk filter`,
		fmt.Errorf(`invalid filter %#v: contains %v of kind %v`, src, val, val.Kind()),
	}
}

func maybeCombineFilters(src, out []Filter) []Filter {
	for _, val := range src {
		if val == nil {
			continue
		}

		if len(out) >= cap(out) {
			panic(Err{
				`building a combined filter`,
				fmt.Errorf(`exceeding filter capacity %v`, cap(out)),
			})
		}
		out = append(out, val)
	}
	return out
}

func ifaceVisit(visTyp, ifaceTyp r.Type, hit byte) byte {
	if visTyp == nil || ifaceTyp == nil {
		return VisNone
	}
	if r.PtrTo(visTyp).Implements(ifaceTyp) {
		return hit
	}
	return VisDesc
}
