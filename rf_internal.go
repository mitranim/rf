package rf

import (
	"fmt"
	r "reflect"
	"sync"
)

var (
	walkerCacheStatic walkerCache

	typeEmptyIface = r.TypeOf((*interface{})(nil)).Elem()
	typeFilter     = r.TypeOf((*Filter)(nil)).Elem()
	typeType       = r.TypeOf((*r.Type)(nil)).Elem()
)

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
		if typ == typeEmptyIface {
			return makeIfaceWalker(typ, parent, field, fil)
		}
		return makeLeafWalker(typ, parent, field, fil)

	default:
		return makeLeafWalker(typ, parent, field, fil)
	}
}

func makePtrWalker(typ, parent r.Type, field r.StructField, fil Filter) Walker {
	inner := walkerCacheStatic.get(typ.Elem(), parent, field, fil)
	if inner != nil {
		return makeNodeWalker(typ, parent, field, fil, ptrWalker{inner})
	}
	return makeLeafWalker(typ, parent, field, fil)
}

func makeListWalker(typ, parent r.Type, field r.StructField, fil Filter) Walker {
	inner := walkerCacheStatic.get(typ.Elem(), parent, field, fil)
	if inner != nil {
		return makeNodeWalker(typ, parent, field, fil, listWalker{inner})
	}
	return makeLeafWalker(typ, parent, field, fil)
}

func makeStructWalker(typ, parent r.Type, field r.StructField, fil Filter) Walker {
	var walkers structWalker
	for i := range Iter(typ.NumField()) {
		walkers = maybeAppendIndexWalker(walkers, typ, i, fil)
	}

	if len(walkers) > 0 {
		return makeNodeWalker(typ, parent, field, fil, walkers)
	}
	return makeLeafWalker(typ, parent, field, fil)
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
	return makeNodeWalker(typ, parent, field, fil, ifaceWalker{parent, field, fil})
}

func makeNodeWalker(typ, parent r.Type, field r.StructField, fil Filter, inner Walker) Walker {
	if fil.ShouldVisit(typ, field) {
		if isFieldValid(field) {
			return selfFieldWalker{field, inner}
		}
		return selfWalker{inner}
	}
	return inner
}

func makeLeafWalker(typ, parent r.Type, field r.StructField, fil Filter) Walker {
	if fil.ShouldVisit(typ, field) {
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
