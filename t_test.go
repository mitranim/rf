package rf

import (
	"fmt"
	r "reflect"
	"testing"
	"time"
	u "unsafe"
)

func TestKind(t *testing.T) {
	eq(t, r.Invalid, Kind(nil))
	eq(t, r.String, Kind(``))
	eq(t, r.Ptr, Kind(stringPtr(``)))
	eq(t, r.Ptr, Kind((*string)(nil)))
	eq(t, r.Slice, Kind([]string(nil)))
	eq(t, r.Ptr, Kind((*[]string)(nil)))
	eq(t, r.Struct, Kind(time.Time{}))
	eq(t, r.Ptr, Kind((*time.Time)(nil)))
}

func TestDerefKind(t *testing.T) {
	eq(t, r.Invalid, DerefKind(nil))
	eq(t, r.String, DerefKind(``))
	eq(t, r.String, DerefKind(stringPtr(``)))
	eq(t, r.String, DerefKind((*string)(nil)))
	eq(t, r.String, DerefKind((**string)(nil)))
	eq(t, r.Slice, DerefKind([]string(nil)))
	eq(t, r.Slice, DerefKind((*[]string)(nil)))
	eq(t, r.Slice, DerefKind((**[]string)(nil)))
	eq(t, r.Struct, DerefKind(time.Time{}))
	eq(t, r.Struct, DerefKind((*time.Time)(nil)))
	eq(t, r.Struct, DerefKind((**time.Time)(nil)))
}

func TestDerefType(t *testing.T) {
	isNil(t, DerefType(nil))
	eq(t, r.TypeOf(``), DerefType(``))
	eq(t, r.TypeOf(``), DerefType((*string)(nil)))
	eq(t, r.TypeOf(``), DerefType((**string)(nil)))
	eq(t, r.TypeOf([]string(nil)), DerefType([]string(nil)))
	eq(t, r.TypeOf([]string(nil)), DerefType([]string{}))
	eq(t, r.TypeOf([]string(nil)), DerefType((*[]string)(nil)))
	eq(t, r.TypeOf([]string(nil)), DerefType((**[]string)(nil)))
}

func TestTypeDeref(t *testing.T) {
	isNil(t, TypeDeref(r.TypeOf(nil)))
	eq(t, r.TypeOf(``), TypeDeref(r.TypeOf((**string)(nil))))
	eq(t, r.TypeOf([]string(nil)), TypeDeref(r.TypeOf(([]string)(nil))))
	eq(t, r.TypeOf([]string(nil)), TypeDeref(r.TypeOf((**[]string)(nil))))
	eq(t, r.TypeOf([]*string(nil)), TypeDeref(r.TypeOf((**[]*string)(nil))))
}

func TestDeref(t *testing.T) {
	eq(t, r.Value{}, Deref(nil))
	eq(t, r.Value{}, Deref((*string)(nil)))
	eq(t, r.Value{}, Deref((*[]string)(nil)))

	test := func(exp, src any) {
		t.Helper()
		eq(t, r.ValueOf(exp).Interface(), Deref(src).Interface())
	}

	test(``, ``)
	test(`one`, `one`)
	test(``, stringPtr(``))
	test(`one`, stringPtr(`one`))
	test(0, 0)
	test(10, 10)
	test(0, intPtr(0))
	test(10, intPtr(10))
	test([]string(nil), []string(nil))
	test([]string{}, []string{})
	test([]string{`one`}, []string{`one`})
	test([]string{}, &[]string{})
	test([]string{`one`}, &[]string{`one`})
}

func TestValueDeref(t *testing.T) {
	eq(t, r.Value{}, ValueDeref(r.Value{}))
	eq(t, r.Value{}, ValueDeref(r.ValueOf((*string)(nil))))
	eq(t, r.ValueOf(``).Interface(), ValueDeref(r.ValueOf(stringPtr(``))).Interface())
	eq(t, r.ValueOf(`one`).Interface(), ValueDeref(r.ValueOf(stringPtr(`one`))).Interface())
}

func TestElemType(t *testing.T) {
	isNil(t, ElemType(nil))
	eq(t, r.TypeOf(nil), ElemType(nil))
	eq(t, r.TypeOf((*any)(nil)).Elem(), ElemType((*any)(nil)))
	eq(t, r.TypeOf((*any)(nil)).Elem(), ElemType((**any)(nil)))
	eq(t, r.TypeOf(``), ElemType(``))
	eq(t, r.TypeOf(``), ElemType(stringPtr(``)))
	eq(t, r.TypeOf(``), ElemType((*string)(nil)))
	eq(t, r.TypeOf(``), ElemType((**string)(nil)))
	eq(t, r.TypeOf(``), ElemType([]string(nil)))
	eq(t, r.TypeOf(``), ElemType((*[]string)(nil)))
	eq(t, r.TypeOf(``), ElemType([]*string(nil)))
	eq(t, r.TypeOf(``), ElemType((**[]string)(nil)))
	eq(t, r.TypeOf(``), ElemType([]**string(nil)))
	eq(t, r.TypeOf(``), ElemType((*[]*string)(nil)))
	eq(t, r.TypeOf(``), ElemType((*[]**string)(nil)))
	eq(t, r.TypeOf(``), ElemType((**[]*string)(nil)))
	eq(t, r.TypeOf(``), ElemType((**[]**string)(nil)))
	eq(t, r.TypeOf(``), ElemType((**[][]**string)(nil)))
	eq(t, r.TypeOf(``), ElemType((**[]**[]**string)(nil)))
	eq(t, r.TypeOf(0), ElemType(([]int)(nil)))
	eq(t, r.TypeOf(0), ElemType(([]**int)(nil)))
	eq(t, r.TypeOf(0), ElemType((**[]int)(nil)))
	eq(t, r.TypeOf(0), ElemType((**[]**int)(nil)))
	eq(t, r.TypeOf(0), ElemType((**[]**[]**int)(nil)))
	eq(t, r.TypeOf([0]string{}), ElemType((**[]**[0]string)(nil)))
	eq(t, r.TypeOf(chan string(nil)), ElemType((******chan string)(nil)))
}

func TestTypeElem(t *testing.T) {
	isNil(t, TypeElem(r.TypeOf(nil)))
	eq(t, r.TypeOf(``), TypeElem(r.TypeOf((**[]**[]**string)(nil))))
	eq(t, r.TypeOf([0]string{}), TypeElem(r.TypeOf((**[]**[0]string)(nil))))
}

func TestValueType(t *testing.T) {
	isNil(t, ValueType(r.Value{}))
	eq(t, r.TypeOf(``), ValueType(r.ValueOf(``)))
	eq(t, r.TypeOf((*string)(nil)), ValueType(r.ValueOf((*string)(nil))))
	eq(t, r.TypeOf((*any)(nil)), ValueType(r.ValueOf((*any)(nil))))
}

func TestTypeKind(t *testing.T) {
	eq(t, r.Invalid, Kind(nil))
	eq(t, r.String, TypeKind(r.TypeOf(``)))
	eq(t, r.Ptr, TypeKind(r.TypeOf(stringPtr(``))))
	eq(t, r.Ptr, TypeKind(r.TypeOf((*string)(nil))))
	eq(t, r.Slice, TypeKind(r.TypeOf([]string(nil))))
	eq(t, r.Ptr, TypeKind(r.TypeOf((*[]string)(nil))))
	eq(t, r.Struct, TypeKind(r.TypeOf(time.Time{})))
	eq(t, r.Ptr, TypeKind(r.TypeOf((*time.Time)(nil))))
}

func TestFuncName(t *testing.T) {
	eq(t, ``, FuncName(nil))
	eq(t, `github.com/mitranim/rf.TestFuncName`, FuncName(TestFuncName))
	eq(t, `github.com/mitranim/rf.FuncName`, FuncName(FuncName))
}

func TestIsNil(t *testing.T) {
	testIsNil(func(exp bool, val any) {
		t.Helper()
		eq(t, exp, IsNil(val))
	})
}

func TestValueIsNil(t *testing.T) {
	testIsNil(func(exp bool, val any) {
		t.Helper()
		eq(t, exp, IsValueNil(r.ValueOf(val)))
	})
}

func testIsNil(test func(bool, any)) {
	test(true, nil)
	test(true, (*string)(nil))
	test(true, []string(nil))
	test(true, (*[]string)(nil))
	test(true, new(func()))
	test(true, new(*string))

	test(false, ``)
	test(false, 0)
	test(false, []string{})
}

func TestIsPublic(t *testing.T) {
	testIsPublic(t, func(val r.StructField) bool {
		return IsPublic(val.PkgPath)
	})
}

func TestIsFieldPublic(t *testing.T) {
	testIsPublic(t, IsFieldPublic)
}

func testIsPublic(t *testing.T, fun func(r.StructField) bool) {
	typ := r.TypeOf(struct {
		Public  string
		private string
	}{})

	fieldPublic, _ := typ.FieldByName(`Public`)
	fieldPrivate, _ := typ.FieldByName(`private`)

	eq(t, true, fun(fieldPublic))
	eq(t, false, fun(fieldPrivate))
}

func TestTagIdent(t *testing.T) {
	test := func(exp, val string) {
		t.Helper()
		eq(t, exp, TagIdent(val))
	}

	test(``, ``)
	test(``, `-`)
	test(``, `-,`)
	test(``, `-,blah`)
	test(``, `-,blah,`)
	test(``, `-,,blah`)
	test(``, `-,blah,blah`)
	test(`ident`, `ident`)
	test(`ident`, `ident,`)
	test(`ident`, `ident,blah`)
	test(`ident`, `ident,blah,`)
	test(`ident`, `ident,,`)
	test(`ident`, `ident,,blah`)
	test(`ident`, `ident,blah,blah`)
}

func TestZero(t *testing.T) {
	Zero(nil)
	Zero((*string)(nil))

	ptr := stringPtr(`one`)
	eq(t, `one`, *ptr)

	Zero(ptr)
	eq(t, ``, *ptr)
}

func TestIsZero(t *testing.T) {
	testIsZero(func(exp bool, val any) {
		t.Helper()
		eq(t, exp, IsZero(val))
	})
}

func TestValueIsZero(t *testing.T) {
	testIsZero(func(exp bool, val any) {
		t.Helper()
		eq(t, exp, IsValueZero(r.ValueOf(val)))
	})
}

func testIsZero(test func(bool, any)) {
	test(true, nil)
	test(true, ``)
	test(true, 0)
	test(true, (*string)(nil))
	test(true, new(string))
	test(true, new(*string))
	test(true, stringPtr(``))
	test(true, []string(nil))

	test(false, `one`)
	test(false, 1)
	test(false, stringPtr(`one`))
	test(false, intPtr(1))
	test(false, []string{})
}

func TestIsKindNilable(t *testing.T) {
	eq(t, true, IsKindNilable(r.Chan))
	eq(t, true, IsKindNilable(r.Func))
	eq(t, true, IsKindNilable(r.Interface))
	eq(t, true, IsKindNilable(r.Map))
	eq(t, true, IsKindNilable(r.Ptr))
	eq(t, true, IsKindNilable(r.Slice))

	eq(t, false, IsKindNilable(r.Invalid))
	eq(t, false, IsKindNilable(r.Bool))
	eq(t, false, IsKindNilable(r.Int))
	eq(t, false, IsKindNilable(r.Int8))
	eq(t, false, IsKindNilable(r.Int16))
	eq(t, false, IsKindNilable(r.Int32))
	eq(t, false, IsKindNilable(r.Int64))
	eq(t, false, IsKindNilable(r.Uint))
	eq(t, false, IsKindNilable(r.Uint8))
	eq(t, false, IsKindNilable(r.Uint16))
	eq(t, false, IsKindNilable(r.Uint32))
	eq(t, false, IsKindNilable(r.Uint64))
	eq(t, false, IsKindNilable(r.Uintptr))
	eq(t, false, IsKindNilable(r.Float32))
	eq(t, false, IsKindNilable(r.Float64))
	eq(t, false, IsKindNilable(r.Complex64))
	eq(t, false, IsKindNilable(r.Complex128))
	eq(t, false, IsKindNilable(r.Array))
	eq(t, false, IsKindNilable(r.String))
	eq(t, false, IsKindNilable(r.Struct))
}

func TestIsKindColl(t *testing.T) {
	eq(t, true, IsKindColl(r.Array))
	eq(t, true, IsKindColl(r.Chan))
	eq(t, true, IsKindColl(r.Map))
	eq(t, true, IsKindColl(r.Slice))
	eq(t, true, IsKindColl(r.String))

	eq(t, false, IsKindColl(r.Invalid))
	eq(t, false, IsKindColl(r.Bool))
	eq(t, false, IsKindColl(r.Int))
	eq(t, false, IsKindColl(r.Int8))
	eq(t, false, IsKindColl(r.Int16))
	eq(t, false, IsKindColl(r.Int32))
	eq(t, false, IsKindColl(r.Int64))
	eq(t, false, IsKindColl(r.Uint))
	eq(t, false, IsKindColl(r.Uint8))
	eq(t, false, IsKindColl(r.Uint16))
	eq(t, false, IsKindColl(r.Uint32))
	eq(t, false, IsKindColl(r.Uint64))
	eq(t, false, IsKindColl(r.Uintptr))
	eq(t, false, IsKindColl(r.Float32))
	eq(t, false, IsKindColl(r.Float64))
	eq(t, false, IsKindColl(r.Complex64))
	eq(t, false, IsKindColl(r.Complex128))
	eq(t, false, IsKindColl(r.Func))
	eq(t, false, IsKindColl(r.Interface))
	eq(t, false, IsKindColl(r.Ptr))
	eq(t, false, IsKindColl(r.Struct))
	eq(t, false, IsKindColl(r.UnsafePointer))
}

func TestIsColl(t *testing.T) {
	test := func(exp bool, val any) {
		t.Helper()
		eq(t, exp, IsColl(val))
	}

	test(true, [0]string{})
	test(true, []string(nil))
	test(true, []string{})
	test(true, map[string]int(nil))
	test(true, map[string]int{})
	test(true, ``)

	test(false, nil)
	test(false, stringPtr(``))
	test(false, 0)
	test(false, struct{}{})
	test(false, IsColl)
}

func TestIsEmptyColl(t *testing.T) {
	testIsEmptyColl(func(exp bool, val any) {
		t.Helper()
		eq(t, exp, IsEmptyColl(val))
	})
}

func TestIsValueEmptyColl(t *testing.T) {
	testIsEmptyColl(func(exp bool, val any) {
		t.Helper()
		eq(t, exp, IsValueEmptyColl(r.ValueOf(val)))
	})
}

func testIsEmptyColl(test func(bool, any)) {
	test(true, ``)
	test(true, []string(nil))
	test(true, []string{})
	test(true, [0]string{})
	test(true, map[string]int(nil))
	test(true, map[string]int{})

	test(false, nil)
	test(false, 0)
	test(false, `one`)
	test(false, []string{`one`})
	test(false, [1]string{})
	test(false, new([]string))
	test(false, &[]string{})
	test(false, map[string]int{`one`: 10})
}

func TestNormNil(t *testing.T) {
	test := func(exp, src any) {
		t.Helper()
		eq(t, exp, NormNil(src))
	}

	test(nil, nil)
	test(nil, (*string)(nil))
	test(stringPtr(``), new(string))
	test(stringPtr(`one`), stringPtr(`one`))
}

func TestLen(t *testing.T) {
	test := func(exp int, val any) {
		t.Helper()
		eq(t, exp, Len(val))
	}

	test(0, nil)
	test(0, ``)
	test(3, `one`)
	test(0, 0)
	test(0, 10)
	test(0, stringPtr(``))
	test(3, stringPtr(`one`))
	test(0, [0]string{})
	test(3, [3]string{})
	test(0, &[0]string{})
	test(3, &[3]string{})
	test(0, []string{})
	test(3, []string{`one`, `two`, `three`})
	test(0, &[]string{})
	test(3, &[]string{`one`, `two`, `three`})
	test(0, map[string]int{})
	test(3, map[string]int{`one`: 10, `two`: 20, `three`: 30})
	test(0, &map[string]int{})
	test(3, &map[string]int{`one`: 10, `two`: 20, `three`: 30})
	test(0, struct{}{})
	test(0, Outer{})
	test(0, &struct{}{})
	test(0, &Outer{})
}

func TestValueLen(t *testing.T) {
	test := func(exp int, val any) {
		t.Helper()
		eq(t, exp, ValueLen(r.ValueOf(val)))
	}

	test(0, nil)
	test(0, ``)
	test(3, `one`)
	test(0, 0)
	test(0, 10)
	test(0, stringPtr(``))
	test(0, stringPtr(`one`))
	test(0, [0]string{})
	test(3, [3]string{})
	test(0, &[0]string{})
	test(0, &[3]string{})
	test(0, []string{})
	test(3, []string{`one`, `two`, `three`})
	test(0, &[]string{})
	test(0, &[]string{`one`, `two`, `three`})
	test(0, map[string]int{})
	test(3, map[string]int{`one`: 10, `two`: 20, `three`: 30})
	test(0, &map[string]int{})
	test(0, &map[string]int{`one`: 10, `two`: 20, `three`: 30})
	test(0, struct{}{})
	test(0, Outer{})
	test(0, &struct{}{})
	test(0, &Outer{})
}

func TestSliceType(t *testing.T) {
	test := func(exp, src any) {
		t.Helper()
		eq(t, r.TypeOf(exp), SliceType(src))
	}

	test([]string(nil), ``)
	test([]string(nil), (*string)(nil))
	test([]string(nil), (**string)(nil))
	test([][]string(nil), (*[]string)(nil))
	test([][]string(nil), (**[]string)(nil))
}

func TestTypeFilter(t *testing.T) {
	test := func(exp byte, visTyp r.Type, fil Filter) {
		t.Helper()
		eq(t, exp, fil.Visit(visTyp, r.StructField{}))
	}

	test(VisDesc, nil, TypeFilter[string]{})
	test(VisDesc, nil, TypeFilter[int]{})
	test(VisBoth, Type[string](), TypeFilter[string]{})
	test(VisDesc, Type[string](), TypeFilter[int]{})
	test(VisDesc, Type[int](), TypeFilter[string]{})
	test(VisBoth, Type[int](), TypeFilter[int]{})
}

func TestTagFilter(t *testing.T) {
	test := func(exp byte, key, val string, tag r.StructTag) {
		t.Helper()
		eq(t, exp, TagFilter{key, val}.Visit(nil, r.StructField{Tag: tag}))
	}

	test(VisDesc, ``, ``, ``)
	test(VisDesc, ``, ``, `json:"one" db:"two"`)
	test(VisDesc, `json`, ``, `json:"one" db:"two"`)
	test(VisDesc, `db`, ``, `json:"one" db:"two"`)
	test(VisDesc, `json`, `two`, `json:"one" db:"two"`)
	test(VisDesc, `db`, `one`, `json:"one" db:"two"`)
	test(VisBoth, `json`, `one`, `json:"one" db:"two"`)
	test(VisBoth, `db`, `two`, `json:"one" db:"two"`)
}

func TestIfaceFilter(t *testing.T) {
	test := func(exp byte, visTyp r.Type, fil Filter) {
		t.Helper()
		eq(t, exp, fil.Visit(visTyp, r.StructField{}))
	}

	test(VisNone, nil, IfaceFilter[fmt.Stringer]{})
	test(VisBoth, Type[time.Time](), IfaceFilter[fmt.Stringer]{})
	test(VisDesc, Type[*time.Time](), IfaceFilter[fmt.Stringer]{})
	test(VisDesc, Type[string](), IfaceFilter[fmt.Stringer]{})
	test(VisDesc, Type[int](), IfaceFilter[fmt.Stringer]{})
}

func TestShallowIfaceFilter(t *testing.T) {
	test := func(exp byte, visTyp r.Type, fil Filter) {
		t.Helper()
		eq(t, exp, fil.Visit(visTyp, r.StructField{}))
	}

	test(VisNone, nil, ShallowIfaceFilter[fmt.Stringer]{})
	test(VisSelf, Type[time.Time](), ShallowIfaceFilter[fmt.Stringer]{})
	test(VisDesc, Type[*time.Time](), ShallowIfaceFilter[fmt.Stringer]{})
	test(VisDesc, Type[string](), ShallowIfaceFilter[fmt.Stringer]{})
	test(VisDesc, Type[int](), ShallowIfaceFilter[fmt.Stringer]{})
}

func TestAppender(t *testing.T) {
	var tar Appender[string]

	tar.Visit(r.ValueOf(``), r.StructField{})
	tar.Visit(r.ValueOf(`one`), r.StructField{})
	tar.Visit(r.ValueOf(``), r.StructField{})
	tar.Visit(r.ValueOf(`two`), r.StructField{})
	tar.Visit(r.ValueOf(``), r.StructField{})

	eq(t, Appender[string]{`one`, `two`}, tar)
}

func TestGetWalker_nil(t *testing.T) {
	isNil(t, GetWalker(nil, nil))
	isNil(t, GetWalker(r.TypeOf(``), nil))
	isNil(t, GetWalker(nil, All{}))
}

func TestGetWalker_caching(t *testing.T) {
	var filter0 TypeFilter[string]
	var filter1 TypeFilter[int]
	typeOuter := r.TypeOf(Outer{})
	typeInner := r.TypeOf(Inner{})

	is(t, GetWalker(typeOuter, filter0), GetWalker(typeOuter, filter0))
	is(t, GetWalker(typeOuter, filter1), GetWalker(typeOuter, filter1))
	is(t, GetWalker(typeInner, filter0), GetWalker(typeInner, filter0))
	is(t, GetWalker(typeInner, filter1), GetWalker(typeInner, filter1))
}

func Test_walking(t *testing.T) {
	{
		var tar Appender[string]

		Walk(testOuterVal, tar.Filter(), &tar)

		eq(
			t,
			Appender[string]{`embed val`, `embed ptr val`, `outer val`, `inner val`, `inner ptr val`, `outer iface`},
			tar,
		)
	}

	{
		var tar Appender[int]

		Walk(testOuterVal, tar.Filter(), &tar)

		eq(
			t,
			Appender[int]{10, 20, 30, 40},
			tar,
		)
	}
}

func TestWalkPtr_invalid(t *testing.T) {
	panics(t, `expected kind ptr`, func() {
		WalkPtr(0, All{}, PanicVis{})
	})
	panics(t, `expected kind ptr`, func() {
		WalkPtr(``, All{}, PanicVis{})
	})
	panics(t, `expected kind ptr`, func() {
		WalkPtr(Outer{}, All{}, PanicVis{})
	})
}

func TestWalkPtr_valid_nil(t *testing.T) {
	WalkPtr(nil, All{}, PanicVis{})
	WalkPtr((*Outer)(nil), All{}, PanicVis{})
}

func TestWalkPtr_valid_non_nil(t *testing.T) {
	testWalkPtr(t, func(val any) {
		WalkPtr(val, TypeFilter[string]{}, VisitorFunc(func(val r.Value, _ r.StructField) {
			val.SetString(`val`)
		}))
	})
}

func TestWalkPtrFunc_invalid(t *testing.T) {
	panics(t, `expected kind ptr`, func() {
		WalkPtrFunc(0, All{}, func(r.Value, r.StructField) {})
	})
	panics(t, `expected kind ptr`, func() {
		WalkPtrFunc(``, All{}, func(r.Value, r.StructField) {})
	})
	panics(t, `expected kind ptr`, func() {
		WalkPtrFunc(Outer{}, All{}, func(r.Value, r.StructField) {})
	})
}

func TestWalkPtrFunc_valid_nil(t *testing.T) {
	WalkPtrFunc(nil, All{}, nil)
	WalkPtrFunc((*Outer)(nil), All{}, nil)
	WalkPtrFunc(nil, All{}, func(r.Value, r.StructField) {})
	WalkPtrFunc((*Outer)(nil), All{}, func(r.Value, r.StructField) {})
}

func TestWalkPtrFunc_valid_non_nil(t *testing.T) {
	testWalkPtr(t, func(val any) {
		WalkPtrFunc(val, TypeFilter[string]{}, func(val r.Value, _ r.StructField) {
			val.SetString(`val`)
		})
	})
}

func testWalkPtr(t testing.TB, fun func(any)) {
	var tar Outer
	fun(&tar)

	var exp Outer
	exp.Embed.EmbedStr = `val`
	exp.OuterStr = `val`
	exp.Inner.InnerStr = `val`

	eq(t, exp, tar)
}

/**
The inner value is missing because we currently don't walk inner occurrences of
a cyclic type, only the outermost occurrence. This is a technical limitation
and we would like to fix this in the future.
*/
func Test_walking_cyclic_by_ptr(t *testing.T) {
	type Type = CyclicByPtr

	testGetWalkerCyclic[Type](t)

	testWalkCyclic(t, &Type{
		Inner: &Type{Value: `inner_val`},
		Value: `outer_val`,
	})
}

// The inner value is missing for the same reason as with `CyclicByPtr`.
func Test_walking_cyclic_by_slice(t *testing.T) {
	type Type = CyclicBySlice

	testGetWalkerCyclic[Type](t)

	testWalkCyclic(t, &Type{
		Inner: []Type{{Value: `inner_val`}},
		Value: `outer_val`,
	})
}

/**
Unlike pointers and slices, we don't walk maps. In this test, the inner value
will be missing even if we someday implement proper walking of cyclic types.
*/
func Test_walking_cyclic_by_map(t *testing.T) {
	type Type = CyclicByMap

	testGetWalkerCyclic[Type](t)

	testWalkCyclic(t, &Type{
		Inner: map[string]CyclicByMap{`inner`: {Value: `inner_val`}},
		Value: `outer_val`,
	})
}

func testGetWalkerCyclic[A any](t *testing.T) {
	var zero A
	var ptr *A

	isNil(t, GetWalker(r.TypeOf(zero), nil))

	isNil(t, GetWalker(r.TypeOf(zero), TypeFilter[int]{}))
	isNil(t, GetWalker(r.TypeOf(ptr), TypeFilter[int]{}))

	isNotNil(t, GetWalker(r.TypeOf(zero), TypeFilter[string]{}))
	isNotNil(t, GetWalker(r.TypeOf(ptr), TypeFilter[string]{}))

	is(
		t,
		GetWalker(r.TypeOf(zero), TypeFilter[string]{}),
		GetWalker(r.TypeOf(zero), TypeFilter[string]{}),
	)

	is(
		t,
		GetWalker(r.TypeOf(ptr), TypeFilter[string]{}),
		GetWalker(r.TypeOf(ptr), TypeFilter[string]{}),
	)
}

func testWalkCyclic(t *testing.T, src any) {
	var tar Appender[string]
	WalkPtr(src, tar.Filter(), &tar)
	eq(t, Appender[string]{`outer_val`}, tar)
}

func TestIsEmbed(t *testing.T) {
	test := func(exp bool, name string) {
		t.Helper()
		field, ok := r.TypeOf(Outer{}).FieldByName(name)
		eq(t, true, ok)
		eq(t, exp, IsEmbed(field))
	}

	test(true, `Embed`)
	test(false, `EmbedPtr`)
	test(false, `OuterStr`)
	test(false, `Inner`)
	test(false, `InnerPtr`)
	test(false, `OuterIface`)
	test(false, `OuterDict`)
}

func TestFields(t *testing.T) {
	testFields(t, Fields)
}

func TestTypeFields(t *testing.T) {
	testFields(t, func(typ any) []r.StructField {
		return TypeFields(r.TypeOf(typ))
	})
}

func testFields(t *testing.T, get func(any) []r.StructField) {
	t.Run(`caching`, func(t *testing.T) {
		testFieldsCaching(t, get)
	})

	t.Run(`shallow`, func(t *testing.T) {
		testFieldsShallow(t, get)
	})

	t.Run(`avoid_expanding_embedded`, func(t *testing.T) {
		typ := r.TypeOf(Outer{})

		eq(
			t,
			[]r.StructField{
				typ.Field(0),
				typ.Field(1),
				typ.Field(2),
				typ.Field(3),
				typ.Field(4),
				typ.Field(5),
				typ.Field(6),
				typ.Field(7),
				typ.Field(8),
			},
			get((*Outer)(nil)),
		)
	})
}

func TestDeepFields(t *testing.T) {
	testDeepFields(t, DeepFields)
}

func TestTypeDeepFields(t *testing.T) {
	testDeepFields(t, func(typ any) []r.StructField {
		return TypeDeepFields(r.TypeOf(typ))
	})
}

func testDeepFields(t *testing.T, get func(any) []r.StructField) {
	t.Run(`caching`, func(t *testing.T) {
		testFieldsCaching(t, get)
	})

	t.Run(`shallow`, func(t *testing.T) {
		testFieldsShallow(t, get)
	})

	t.Run(`embed by value`, func(t *testing.T) {
		eq(
			t,
			[]r.StructField{
				{
					Name:      `One`,
					Type:      r.TypeOf(string(``)),
					Anonymous: false,
					Offset:    0,
					Index:     []int{0},
				},
				{
					Name:      `EmbedStr`,
					Type:      r.TypeOf(string(``)),
					Anonymous: false,
					Offset:    u.Sizeof(string(``)),
					Tag:       `json:"embedStr" db:"embed_str"`,
					Index:     []int{1, 0},
				},
				{
					Name:      `EmbedNum`,
					Type:      r.TypeOf(int(0)),
					Anonymous: false,
					Offset:    u.Sizeof(string(``)) + u.Sizeof(string(``)),
					Tag:       `json:"embedNum" db:"embed_num"`,
					Index:     []int{1, 1},
				},
				{
					Name:      `Inner`,
					Type:      r.TypeOf((*Inner)(nil)),
					Anonymous: true,
					Offset:    u.Sizeof(string(``)) + u.Sizeof(string(``)) + u.Sizeof(int(0)),
					Index:     []int{2},
				},
				{
					Name:      `Two`,
					Type:      r.TypeOf(int(0)),
					Anonymous: false,
					Offset:    u.Sizeof(string(``)) + u.Sizeof(string(``)) + u.Sizeof(int(0)) + u.Sizeof((*Inner)(nil)),
					Index:     []int{3},
				},
			},
			get(struct {
				One string
				Embed
				*Inner
				Two int
			}{}),
		)
	})
}

func testFieldsShallow(t *testing.T, get func(any) []r.StructField) {
	eq(t, []r.StructField(nil), get(nil))
	eq(t, []r.StructField{}, get(struct{}{}))
	eq(t, []r.StructField{}, get((*struct{})(nil)))

	eq(
		t,
		[]r.StructField{
			r.TypeOf(Inner{}).Field(0),
			r.TypeOf(Inner{}).Field(1),
		},
		get(Inner{}),
	)

	eq(
		t,
		[]r.StructField{
			r.TypeOf(Inner{}).Field(0),
			r.TypeOf(Inner{}).Field(1),
		},
		get((*Inner)(nil)),
	)

	t.Run(`no structs`, func(t *testing.T) {
		eq(
			t,
			[]r.StructField{
				{
					Name:   `One`,
					Type:   r.TypeOf(string(``)),
					Offset: 0,
					Index:  []int{0},
				},
				{
					Name:   `Two`,
					Type:   r.TypeOf(int(0)),
					Offset: u.Sizeof(string(``)),
					Index:  []int{1},
				},
			},
			get(struct {
				One string
				Two int
			}{}),
		)
	})

	t.Run(`non-embedded structs`, func(t *testing.T) {
		eq(
			t,
			[]r.StructField{
				{
					Name:   `One`,
					Type:   r.TypeOf(string(``)),
					Offset: 0,
					Index:  []int{0},
				},
				{
					Name:   `Two`,
					Type:   r.TypeOf(Inner{}),
					Offset: u.Sizeof(string(``)),
					Index:  []int{1},
				},
				{
					Name:   `Three`,
					Type:   r.TypeOf((*Inner)(nil)),
					Offset: u.Sizeof(string(``)) + u.Sizeof(Inner{}),
					Index:  []int{2},
				},
				{
					Name:   `Four`,
					Type:   r.TypeOf(int(0)),
					Offset: u.Sizeof(string(``)) + u.Sizeof(Inner{}) + u.Sizeof((*Inner)(nil)),
					Index:  []int{3},
				},
			},
			get(struct {
				One   string
				Two   Inner
				Three *Inner
				Four  int
			}{}),
		)
	})

	t.Run(`non-struct embed`, func(t *testing.T) {
		type Str = string
		type Int = int

		eq(
			t,
			[]r.StructField{
				{
					Name:      `Str`,
					Type:      r.TypeOf(Str(``)),
					Anonymous: true,
					Offset:    0,
					Index:     []int{0},
				},
				{
					Name:      `Int`,
					Type:      r.TypeOf(Int(0)),
					Anonymous: true,
					Offset:    u.Sizeof(Str(``)),
					Index:     []int{1},
				},
				{
					Name:      `Stringer`,
					Type:      r.TypeOf((*fmt.Stringer)(nil)).Elem(),
					Anonymous: true,
					Offset:    u.Sizeof(Str(``)) + u.Sizeof(Int(0)),
					Index:     []int{2},
				},
				{
					Name:      `Path`,
					Type:      r.TypeOf(Path(nil)),
					Anonymous: true,
					Offset:    u.Sizeof(Str(``)) + u.Sizeof(Int(0)) + u.Sizeof(fmt.Stringer(nil)),
					Index:     []int{3},
				},
			},
			get(struct {
				Str
				Int
				fmt.Stringer
				Path
			}{}),
		)
	})

	t.Run(`embed by pointer`, func(t *testing.T) {
		eq(
			t,
			[]r.StructField{
				{
					Name:      `One`,
					Type:      r.TypeOf(string(``)),
					Anonymous: false,
					Offset:    0,
					Index:     []int{0},
				},
				{
					Name:      `Embed`,
					Type:      r.TypeOf((*Embed)(nil)),
					Anonymous: true,
					Offset:    u.Sizeof(string(``)),
					Index:     []int{1},
				},
				{
					Name:      `Two`,
					Type:      r.TypeOf(int(0)),
					Anonymous: false,
					Offset:    u.Sizeof(string(``)) + u.Sizeof((*Embed)(nil)),
					Index:     []int{2},
				},
			},
			get(struct {
				One string
				*Embed
				Two int
			}{}),
		)
	})
}

func testFieldsCaching(t testing.TB, get func(any) []r.StructField) {
	test := func(typA, typB any) {
		t.Helper()

		valA := get(typA)
		valB := get(typB)

		is(t, &valA[0], &valB[0])
	}

	test((*Inner)(nil), (*Inner)(nil))
	test(Inner{}, (*Inner)(nil))
	test((*Inner)(nil), Inner{})
	test(Inner{}, Inner{})
}

func TestOffsetFields(t *testing.T) {
	testOffsetFields(t, OffsetFields)
}

func TestTypeOffsetFields(t *testing.T) {
	testOffsetFields(t, func(typ any) map[uintptr][]r.StructField {
		return TypeOffsetFields(r.TypeOf(typ))
	})
}

func testOffsetFields(t testing.TB, get func(any) map[uintptr][]r.StructField) {
	eq(
		t,
		map[uintptr][]r.StructField(nil),
		get(nil),
	)

	eq(
		t,
		map[uintptr][]r.StructField{},
		get(struct{}{}),
	)

	eq(
		t,
		map[uintptr][]r.StructField{},
		get((*struct{})(nil)),
	)

	eq(
		t,
		map[uintptr][]r.StructField{0: {{
			Name:  `One`,
			Type:  r.TypeOf(string(``)),
			Index: []int{0},
		}}},
		get(struct{ One string }{}),
	)

	type Empty2 struct{}
	type Empty1 struct{ Empty2 }
	type Empty0 struct {
		Empty1
		Empty2
	}

	eq(
		t,
		map[uintptr][]r.StructField{
			0: {{
				Name:  `One`,
				Type:  r.TypeOf(string(``)),
				Index: []int{0},
			}},
			u.Sizeof(string(``)): {{
				Name:   `Two`,
				Type:   r.TypeOf(int(0)),
				Offset: u.Sizeof(string(``)),
				Index:  []int{2},
			}},
		},
		get(struct {
			One string
			Empty0
			Two int
		}{}),
	)

	eq(
		t,
		map[uintptr][]r.StructField{
			0: {{
				Name:  `One`,
				Type:  r.TypeOf(string(``)),
				Index: []int{0},
			}},
			u.Sizeof(string(``)): {
				{
					Name:   `Two`,
					Type:   r.TypeOf(Empty0{}),
					Offset: u.Sizeof(string(``)),
					Index:  []int{1},
				},
				{
					Name:   `Three`,
					Type:   r.TypeOf(int(0)),
					Offset: u.Sizeof(string(``)),
					Index:  []int{2},
				},
			},
		},
		get(struct {
			One   string
			Two   Empty0
			Three int
		}{}),
	)
}

func TestInvertSelf(t *testing.T) {
	test := func(exp byte, val InvertSelf) {
		t.Helper()
		eq(t, exp, val.Visit(nil, r.StructField{}))
	}

	test(0b_0000_0000, InvertSelf{})
	test(0b_0000_0000, InvertSelf{Self{}})
	test(0b_0000_0011, InvertSelf{Desc{}})
	test(0b_1111_1110, InvertSelf{All{}})
}

func TestAnd(t *testing.T) {
	test := func(exp0, exp1 byte, val And) {
		t.Helper()
		eq(t, exp0, exp1)
		eq(t, exp0, val.Visit(nil, r.StructField{}))
		eq(t, exp1, val.Visit(nil, r.StructField{}))
	}

	test(0b_0000_0000, VisNone, And{})

	test(0b_1111_1111, VisAll, And{All{}})
	test(0b_0000_0001, VisSelf, And{Self{}})
	test(0b_0000_0010, VisDesc, And{Desc{}})
	test(0b_0000_0011, VisBoth, And{Both{}})

	test(0b_1111_1111, VisAll, And{All{}, All{}})
	test(0b_0000_0011, VisBoth, And{Both{}, Both{}})

	test(0b_0000_0001, VisSelf, And{All{}, Self{}})
	test(0b_0000_0001, VisSelf, And{Self{}, All{}})

	test(0b_0000_0001, VisSelf, And{Both{}, Self{}})
	test(0b_0000_0001, VisSelf, And{Self{}, Both{}})

	test(0b_0000_0010, VisDesc, And{All{}, Desc{}})
	test(0b_0000_0010, VisDesc, And{Desc{}, All{}})

	test(0b_0000_0010, VisDesc, And{Both{}, Desc{}})
	test(0b_0000_0010, VisDesc, And{Desc{}, Both{}})

	test(0b_0000_0000, VisNone, And{Self{}, Desc{}})
	test(0b_0000_0000, VisNone, And{Desc{}, Self{}})
}

func TestMaybeAnd(t *testing.T) {
	isNil(t, MaybeAnd())
	isNil(t, MaybeAnd(nil))
	isNil(t, MaybeAnd(nil, nil))
	isNil(t, MaybeAnd(nil, nil, nil))
	eq(t, Self{}, MaybeAnd(Self{}))
	eq(t, Self{}, MaybeAnd(Self{}, nil))
	eq(t, Self{}, MaybeAnd(nil, Self{}))
	eq(t, Self{}, MaybeAnd(nil, Self{}, nil))
	eq(t, Desc{}, MaybeAnd(nil, Desc{}, nil))
	eq(t, And{Self{}, Desc{}}, MaybeAnd(Self{}, Desc{}))
	eq(t, And{Self{}, Desc{}}, MaybeAnd(nil, Self{}, nil, Desc{}, nil))
}

func TestOr(t *testing.T) {
	test := func(exp0, exp1 byte, val Or) {
		t.Helper()
		eq(t, exp0, exp1)
		eq(t, exp0, val.Visit(nil, r.StructField{}))
		eq(t, exp1, val.Visit(nil, r.StructField{}))
	}

	test(0b_0000_0000, VisNone, Or{})

	test(0b_1111_1111, VisAll, Or{All{}})
	test(0b_0000_0001, VisSelf, Or{Self{}})
	test(0b_0000_0010, VisDesc, Or{Desc{}})
	test(0b_0000_0011, VisBoth, Or{Both{}})

	test(0b_1111_1111, VisAll, Or{All{}, All{}})
	test(0b_0000_0011, VisBoth, Or{Both{}, Both{}})

	test(0b_1111_1111, VisAll, Or{All{}, Self{}})
	test(0b_1111_1111, VisAll, Or{Self{}, All{}})

	test(0b_0000_0011, VisBoth, Or{Both{}, Self{}})
	test(0b_0000_0011, VisBoth, Or{Self{}, Both{}})

	test(0b_1111_1111, VisAll, Or{All{}, Desc{}})
	test(0b_1111_1111, VisAll, Or{Desc{}, All{}})

	test(0b_0000_0011, VisBoth, Or{Both{}, Desc{}})
	test(0b_0000_0011, VisBoth, Or{Desc{}, Both{}})

	test(0b_0000_0011, VisBoth, Or{Self{}, Desc{}})
	test(0b_0000_0011, VisBoth, Or{Desc{}, Self{}})
}

func TestMaybeOr(t *testing.T) {
	isNil(t, MaybeOr())
	isNil(t, MaybeOr(nil))
	isNil(t, MaybeOr(nil, nil))
	isNil(t, MaybeOr(nil, nil, nil))
	eq(t, Self{}, MaybeOr(Self{}))
	eq(t, Self{}, MaybeOr(Self{}, nil))
	eq(t, Self{}, MaybeOr(nil, Self{}))
	eq(t, Self{}, MaybeOr(nil, Self{}, nil))
	eq(t, Desc{}, MaybeOr(nil, Desc{}, nil))
	eq(t, Or{Self{}, Desc{}}, MaybeOr(Self{}, Desc{}))
	eq(t, Or{Self{}, Desc{}}, MaybeOr(nil, Self{}, nil, Desc{}, nil))
}

func Test_reflect_Type_FieldByIndex(t *testing.T) {
	typ := r.TypeOf(struct {
		One [4]string
		Two [8]int
		Embed
	}{})

	field := typ.FieldByIndex([]int{2, 1})

	eq(
		t,
		r.StructField{
			Name:      `EmbedNum`,
			Type:      r.TypeOf(int(0)),
			Anonymous: false,
			Tag:       `json:"embedNum" db:"embed_num"`,
			Offset:    u.Sizeof(string(``)), // Part of our problem.
			Index:     []int{1},             // Another part of our problem.
		},
		field,
	)
}

func Test_reflect_Type_FieldByName(t *testing.T) {
	typ := r.TypeOf(struct {
		One [4]string
		Two [8]int
		Embed
	}{})

	field, _ := typ.FieldByName(`EmbedNum`)

	eq(
		t,
		r.StructField{
			Name:      `EmbedNum`,
			Type:      r.TypeOf(int(0)),
			Anonymous: false,
			Tag:       `json:"embedNum" db:"embed_num"`,
			Offset:    u.Sizeof(string(``)), // This is our problem.
			Index:     []int{2, 1},
		},
		field,
	)
}
