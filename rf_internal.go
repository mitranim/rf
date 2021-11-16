package rf

import (
	"fmt"
	r "reflect"
	"sync"
)

const (
	expectedStructNesting = 8
)

var (
	walkerCacheStatic walkerCache

	typeFilter = r.TypeOf((*Filter)(nil)).Elem()
	typeType   = r.TypeOf((*r.Type)(nil)).Elem()
)

/*
Return value of `Filter.Visit`. Private, to avoid putting a concrete library
type into an interface.
*/
type vis byte

func (self vis) self() bool { return (self & VisSelf) != 0 }
func (self vis) desc() bool { return (self & VisDesc) != 0 }

type walkey struct {
	Type   r.Type
	Parent r.Type
	Index  int
	Filter Filter
}

func (self walkey) GoString() string {
	return fmt.Sprintf(`rf.walkey{%v, %v, %v, %#v}`, self.Type, self.Parent, self.Index, self.Filter)
}

type walkerCache struct {
	sync.RWMutex
	Map map[walkey]Walker
}

func (self *walkerCache) get(typ, parent r.Type, field r.StructField, fil Filter) Walker {
	key := walkey{typ, parent, fieldIndex(field), fil}
	val, ok := self.got(key)
	if ok {
		return val
	}
	return self.set(key, makeWalker(typ, parent, field, fil))
}

func (self *walkerCache) got(key walkey) (Walker, bool) {
	self.RLock()
	defer self.RUnlock()

	val, ok := self.Map[key]
	return val, ok
}

func (self *walkerCache) set(key walkey, next Walker) Walker {
	self.Lock()
	defer self.Unlock()

	prev, ok := self.Map[key]
	if ok {
		return prev
	}

	if self.Map == nil {
		self.Map = map[walkey]Walker{}
	}
	self.Map[key] = next

	return next
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
	for i := range Iter(val.Len()) {
		self[0].Walk(val.Index(i), vis)
	}
}

type structWalker []indexWalker

func (self structWalker) Walk(val r.Value, vis Visitor) {
	for _, walker := range self {
		walker.Walk(val, vis)
	}
}

type indexWalker struct {
	Index int
	Inner Walker
}

func (self indexWalker) Walk(val r.Value, vis Visitor) {
	self.Inner.Walk(val.Field(self.Index), vis)
}

type ifaceWalker struct {
	Parent r.Type
	Field  r.StructField
	Filter Filter
}

func (self ifaceWalker) Walk(val r.Value, vis Visitor) {
	if val.IsNil() {
		return
	}

	val = val.Elem()
	walker := walkerCacheStatic.get(val.Type(), self.Parent, self.Field, self.Filter)
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

func makeWalker(typ, parent r.Type, field r.StructField, fil Filter) Walker {
	if typ == nil || fil == nil {
		return nil
	}

	switch typ.Kind() {
	case r.Ptr:
		return makePtrWalker(typ, parent, field, fil)

	case r.Array, r.Slice:
		return makeListWalker(typ, parent, field, fil)

	case r.Struct:
		return makeStructWalker(typ, parent, field, fil)

	case r.Interface:
		return makeIfaceWalker(typ, parent, field, fil)

	default:
		return makeLeafWalker(typ, parent, field, vis(fil.Visit(typ, field)))
	}
}

func makePtrWalker(typ, parent r.Type, field r.StructField, fil Filter) Walker {
	vis := vis(fil.Visit(typ, field))

	if vis.desc() {
		inner := walkerCacheStatic.get(typ.Elem(), parent, field, fil)
		if inner != nil {
			return makeNodeWalker(typ, parent, field, vis, ptrWalker{inner})
		}
	}

	return makeLeafWalker(typ, parent, field, vis)
}

func makeListWalker(typ, parent r.Type, field r.StructField, fil Filter) Walker {
	vis := vis(fil.Visit(typ, field))

	if vis.desc() {
		inner := walkerCacheStatic.get(typ.Elem(), parent, field, fil)
		if inner != nil {
			return makeNodeWalker(typ, parent, field, vis, listWalker{inner})
		}
	}

	return makeLeafWalker(typ, parent, field, vis)
}

func makeStructWalker(typ, parent r.Type, field r.StructField, fil Filter) Walker {
	vis := vis(fil.Visit(typ, field))

	if vis.desc() {
		var walkers structWalker
		for i := range Iter(typ.NumField()) {
			walkers = maybeAppendIndexWalker(walkers, typ, i, fil)
		}

		if len(walkers) > 0 {
			return makeNodeWalker(typ, parent, field, vis, walkers)
		}
	}

	return makeLeafWalker(typ, parent, field, vis)
}

func maybeAppendIndexWalker(out structWalker, typ r.Type, index int, fil Filter) structWalker {
	field := typ.Field(index)
	inner := walkerCacheStatic.get(field.Type, typ, field, fil)
	if inner != nil {
		out = append(out, indexWalker{index, inner})
	}
	return out
}

func makeIfaceWalker(typ, parent r.Type, field r.StructField, fil Filter) Walker {
	vis := vis(fil.Visit(typ, field))
	if vis.desc() {
		return makeNodeWalker(typ, parent, field, vis, ifaceWalker{parent, field, fil})
	}
	return makeLeafWalker(typ, parent, field, vis)
}

func makeNodeWalker(typ, parent r.Type, field r.StructField, vis vis, inner Walker) Walker {
	if inner == nil {
		panic(Err{
			`making node walker`,
			fmt.Errorf(
				`internal violation: attempted to construct useless node walker without inner walker`,
			),
		})
	}

	if vis.self() {
		if isFieldValid(field) {
			return selfFieldWalker{field, inner}
		}
		return selfWalker{inner}
	}
	return inner
}

func makeLeafWalker(typ, parent r.Type, field r.StructField, vis vis) Walker {
	if vis.self() {
		if isFieldValid(field) {
			return leafFieldWalker(field)
		}
		return leafWalker{}
	}
	return nil
}

func isFieldValid(val r.StructField) bool {
	return val.Type != nil && len(val.Index) > 0
}

func fieldIndex(val r.StructField) int {
	if len(val.Index) > 0 {
		return val.Index[0]
	}
	return 0
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
	for i := range Iter(val.Len()) {
		validateFilterValue(src, val.Index(i))
	}
}

func validateFilterStruct(src Filter, val r.Value) {
	for i := range Iter(val.NumField()) {
		validateFilterValue(src, val.Field(i))
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

var typeFieldsCache = Cache{Func: func(typ r.Type) interface{} {
	typ = ValidateTypeStruct(typ)
	out := make([]r.StructField, 0, typ.NumField())
	for i := range Iter(typ.NumField()) {
		out = append(out, typ.Field(i))
	}
	return out
}}

var typeDeepFieldsCache = Cache{Func: func(typ r.Type) interface{} {
	typ = ValidateTypeStruct(typ)
	buf := make([]r.StructField, 0, typ.NumField())
	path := make(Path, 0, expectedStructNesting)
	appendStructDeepFields(&buf, &path, r.StructField{Type: typ, Anonymous: true})
	return buf
}}

func appendStructDeepFields(
	buf *[]r.StructField, path *Path, field r.StructField,
) {
	defer path.Add(field.Index).Reset()

	typ := TypeDeref(field.Type)
	if IsEmbed(field) {
		for _, inner := range TypeFields(typ) {
			inner.Offset += field.Offset
			appendStructDeepFields(buf, path, inner)
		}
	} else {
		field.Index = path.Copy()
		*buf = append(*buf, field)
	}
}

var typeOffsetFieldsCache = Cache{Func: func(typ r.Type) interface{} {
	typ = ValidateTypeStruct(typ)
	if typ == nil {
		return map[uintptr][]r.StructField(nil)
	}

	fields := TypeDeepFields(typ)
	out := make(map[uintptr][]r.StructField, len(fields))
	for _, field := range fields {
		out[field.Offset] = append(out[field.Offset], field)
	}
	return out
}}
