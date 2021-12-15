package rf

import (
	r "reflect"
	"sync"
	"testing"
	"unsafe"
)

// Causes the input to escape, allowing us to measure allocs.
var (
	pathNop    = func(Path) {}
	filterNop  = func(Filter) {}
	stringsNop = func(string, string) {}
	bytesNop   = func([]byte) {}
)

func BenchmarkGetWalker(b *testing.B) {
	for range Iter(b.N) {
		benchGetWalker()
	}
}

func benchGetWalker() {
	GetWalker(r.TypeOf(&testOuter), All{})
}

func BenchmarkWalk(b *testing.B) {
	for range Iter(b.N) {
		benchWalk()
	}
}

func benchWalk() {
	Walk(r.ValueOf(&testOuter), All{}, Nop{})
}

func Benchmark_range_slice_values(b *testing.B) {
	for range Iter(b.N) {
		benchRangeSliceValues(testSlice)
	}
}

func benchRangeSliceValues(src interface{}) {
	val := r.ValueOf(src)
	for i := range Iter(val.Len()) {
		val.Index(i).Interface()
	}
}

func Benchmark_range_slice_pointers(b *testing.B) {
	for range Iter(b.N) {
		benchRangeSlicePointers(&testSlice)
	}
}

func benchRangeSlicePointers(src interface{}) {
	val := r.ValueOf(src).Elem()
	for i := range Iter(val.Len()) {
		val.Index(i).Addr().Interface()
	}
}

func Benchmark_append_values(b *testing.B) {
	slice := make([]string, 0, 32)
	val := r.ValueOf(&slice).Elem()
	app := Appender{val}
	src := r.ValueOf(&testOuter)
	filter := app.Filter()
	visitor := Visitor(app)

	b.ResetTimer()

	for range Iter(b.N) {
		val.SetLen(0)
		Walk(src, filter, visitor)
	}
}

func BenchmarkTypeFilter(b *testing.B) {
	for range Iter(b.N) {
		filterNop(TypeFilter{r.TypeOf((*string)(nil))})
	}
}

func BenchmarkTypeFilterFor(b *testing.B) {
	for range Iter(b.N) {
		filterNop(TypeFilterFor((*string)(nil)))
	}
}

func BenchmarkMaybeAndEmpty(b *testing.B) {
	for range Iter(b.N) {
		filterNop(MaybeAnd(nil, nil))
	}
}

func BenchmarkMaybeAndUnary(b *testing.B) {
	for range Iter(b.N) {
		filterNop(MaybeAnd(nil, All{}))
	}
}

func BenchmarkMaybeAndMulti(b *testing.B) {
	for range Iter(b.N) {
		filterNop(MaybeAnd(Self{}, Desc{}))
	}
}

func BenchmarkAndEmpty(b *testing.B) {
	for range Iter(b.N) {
		filterNop(And{})
	}
}

func BenchmarkAndMulti(b *testing.B) {
	for range Iter(b.N) {
		filterNop(And{Self{}, Desc{}})
	}
}

func Benchmark_map_iter_alloc(b *testing.B) {
	for range Iter(b.N) {
		testDictVal.MapRange()
	}
}

func Benchmark_map_iter_pool(b *testing.B) {
	for range Iter(b.N) {
		benchMapIterPool()
	}
}

func benchMapIterPool() {
	iter := getMapIter(testDictVal)
	defer putMapIter(iter)
}

var mapIterPool = sync.Pool{New: func() interface{} { return new(r.MapIter) }}

func getMapIter(val r.Value) *r.MapIter {
	iter := mapIterPool.Get().(*r.MapIter)
	mapIterReset(iter, val)
	return iter
}

func putMapIter(iter *r.MapIter) {
	mapIterClear(iter)
	mapIterPool.Put(iter)
}

// Workaround for the missing `(*r.MapIter).Reset` in Go 1.17.
func mapIterClear(iter *r.MapIter) {
	*iter = r.MapIter{}
}

// Workaround for the missing `(*r.MapIter).Reset` in Go 1.17.
func mapIterReset(iter *r.MapIter, val r.Value) {
	mapIterClear(iter)
	*(*r.Value)(unsafe.Pointer(iter)) = val
}

// Kinda slow, but tolerable.
func Benchmark_map_iter_range(b *testing.B) {
	for range Iter(b.N) {
		benchMapIterRange()
	}
}

func benchMapIterRange() {
	iter := getMapIter(testDictVal)
	defer putMapIter(iter)

	for iter.Next() {
	}
}

// Too slow, non-viable.
func Benchmark_map_iter_range_with_access(b *testing.B) {
	for range Iter(b.N) {
		benchMapIterRangeWithAccess()
	}
}

func benchMapIterRangeWithAccess() {
	iter := getMapIter(testDictVal)
	defer putMapIter(iter)

	for iter.Next() {
		iter.Key()
		iter.Value()
	}
}

func Benchmark_map_range(b *testing.B) {
	for range Iter(b.N) {
		for key, val := range testDict {
			stringsNop(key, val)
		}
	}
}

func Benchmark_Path_Add_Reset(b *testing.B) {
	path := make(Path, 0, expectedStructNesting)
	b.ResetTimer()

	for i := range Iter(b.N) {
		benchPathAddReset(&path, i)
	}
}

func benchPathAddReset(path *Path, index int) {
	defer path.Add([]int{index}).Reset()
	pathNop(*path)
}

func Benchmark_reflect_Value_Interface(b *testing.B) {
	val := r.ValueOf(struct{ Inner []byte }{}).Field(0)
	b.ResetTimer()

	for range Iter(b.N) {
		bytesNop(val.Interface().([]byte))
	}
}

func Benchmark_reflect_Value_Addr_Interface(b *testing.B) {
	val := r.ValueOf(&struct{ Inner []byte }{}).Elem().Field(0)
	b.ResetTimer()

	for range Iter(b.N) {
		bytesNop(*val.Addr().Interface().(*[]byte))
	}
}
