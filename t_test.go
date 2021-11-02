package rf

import (
	r "reflect"
	"testing"
	"time"
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
	eq(t, nil, DerefType(nil))
	eq(t, r.TypeOf(``), DerefType(``))
	eq(t, r.TypeOf(``), DerefType((*string)(nil)))
	eq(t, r.TypeOf(``), DerefType((**string)(nil)))
	eq(t, r.TypeOf([]string(nil)), DerefType([]string(nil)))
	eq(t, r.TypeOf([]string(nil)), DerefType([]string{}))
	eq(t, r.TypeOf([]string(nil)), DerefType((*[]string)(nil)))
	eq(t, r.TypeOf([]string(nil)), DerefType((**[]string)(nil)))
}

func TestTypeDeref(t *testing.T) {
	eq(t, nil, TypeDeref(r.TypeOf(nil)))
	eq(t, r.TypeOf(``), TypeDeref(r.TypeOf((**string)(nil))))
	eq(t, r.TypeOf([]string(nil)), TypeDeref(r.TypeOf(([]string)(nil))))
	eq(t, r.TypeOf([]string(nil)), TypeDeref(r.TypeOf((**[]string)(nil))))
	eq(t, r.TypeOf([]*string(nil)), TypeDeref(r.TypeOf((**[]*string)(nil))))
}

func TestDerefValue(t *testing.T) {
	eq(t, r.Value{}, DerefValue(nil))
	eq(t, r.Value{}, DerefValue((*string)(nil)))
	eq(t, r.Value{}, DerefValue((*[]string)(nil)))

	test := func(exp, src interface{}) {
		t.Helper()
		eq(t, r.ValueOf(exp).Interface(), DerefValue(src).Interface())
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
	eq(t, nil, ElemType(nil))
	eq(t, r.TypeOf(nil), ElemType(nil))
	eq(t, r.TypeOf((*interface{})(nil)).Elem(), ElemType((*interface{})(nil)))
	eq(t, r.TypeOf((*interface{})(nil)).Elem(), ElemType((**interface{})(nil)))
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
	eq(t, nil, TypeElem(r.TypeOf(nil)))
	eq(t, r.TypeOf(``), TypeElem(r.TypeOf((**[]**[]**string)(nil))))
	eq(t, r.TypeOf([0]string{}), TypeElem(r.TypeOf((**[]**[0]string)(nil))))
}

func TestValueType(t *testing.T) {
	eq(t, nil, ValueType(r.Value{}))
	eq(t, r.TypeOf(``), ValueType(r.ValueOf(``)))
	eq(t, r.TypeOf((*string)(nil)), ValueType(r.ValueOf((*string)(nil))))
	eq(t, r.TypeOf((*interface{})(nil)), ValueType(r.ValueOf((*interface{})(nil))))
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
	testIsNil(func(exp bool, val interface{}) {
		t.Helper()
		eq(t, exp, IsNil(val))
	})
}

func TestValueIsNil(t *testing.T) {
	testIsNil(func(exp bool, val interface{}) {
		t.Helper()
		eq(t, exp, IsValueNil(r.ValueOf(val)))
	})
}

func testIsNil(test func(bool, interface{})) {
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
	typ := r.TypeOf(struct {
		Public  string
		private string
	}{})

	fieldPublic, _ := typ.FieldByName(`Public`)
	fieldPrivate, _ := typ.FieldByName(`private`)

	eq(t, true, IsPublic(fieldPublic.PkgPath))
	eq(t, false, IsPublic(fieldPrivate.PkgPath))
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
	testIsZero(func(exp bool, val interface{}) {
		t.Helper()
		eq(t, exp, IsZero(val))
	})
}

func TestValueIsZero(t *testing.T) {
	testIsZero(func(exp bool, val interface{}) {
		t.Helper()
		eq(t, exp, IsValueZero(r.ValueOf(val)))
	})
}

func testIsZero(test func(bool, interface{})) {
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
	test := func(exp bool, val interface{}) {
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
	testIsEmptyColl(func(exp bool, val interface{}) {
		t.Helper()
		eq(t, exp, IsEmptyColl(val))
	})
}

func TestIsValueEmptyColl(t *testing.T) {
	testIsEmptyColl(func(exp bool, val interface{}) {
		t.Helper()
		eq(t, exp, IsValueEmptyColl(r.ValueOf(val)))
	})
}

func testIsEmptyColl(test func(bool, interface{})) {
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
	test := func(exp, src interface{}) {
		t.Helper()
		eq(t, exp, NormNil(src))
	}

	test(nil, nil)
	test(nil, (*string)(nil))
	test(stringPtr(``), new(string))
	test(stringPtr(`one`), stringPtr(`one`))
}

func TestDerefLen(t *testing.T) {
	test := func(exp int, val interface{}) {
		t.Helper()
		eq(t, exp, DerefLen(val))
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
	test := func(exp int, val interface{}) {
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
	test := func(exp, src interface{}) {
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
	test := func(exp bool, filTyp, visTyp interface{}) {
		t.Helper()
		eq(t, exp, TypeFilter{r.TypeOf(filTyp)}.ShouldVisit(r.TypeOf(visTyp), r.StructField{}))
	}

	test(true, nil, nil)
	test(true, ``, ``)
	test(false, ``, nil)
	test(false, nil, ``)
}

func TestTagFilter(t *testing.T) {
	test := func(exp bool, key, val string, tag r.StructTag) {
		t.Helper()
		eq(t, exp, TagFilter{key, val}.ShouldVisit(nil, r.StructField{Tag: tag}))
	}

	test(false, ``, ``, ``)
	test(false, ``, ``, `json:"one" db:"two"`)
	test(false, `json`, ``, `json:"one" db:"two"`)
	test(false, `db`, ``, `json:"one" db:"two"`)
	test(false, `json`, `two`, `json:"one" db:"two"`)
	test(false, `db`, `one`, `json:"one" db:"two"`)
	test(true, `json`, `one`, `json:"one" db:"two"`)
	test(true, `db`, `two`, `json:"one" db:"two"`)
}

func TestAppenderFor(t *testing.T) {
	test := func(exp, val interface{}) {
		t.Helper()
		eq(t, exp, AppenderFor(val).Interface())
	}

	test([]string(nil), ``)
	test([]string(nil), (*string)(nil))
	test([]string(nil), (**string)(nil))
	test([]int(nil), 0)
	test([]int(nil), (*int)(nil))
	test([]int(nil), (**int)(nil))

	eq(t, true, AppenderFor((*string)(nil))[0].CanSet())
}

func TestAppender(t *testing.T) {
	val := AppenderFor((*string)(nil))

	val.Visit(r.ValueOf(``), r.StructField{})
	val.Visit(r.ValueOf(`one`), r.StructField{})
	val.Visit(r.ValueOf(``), r.StructField{})
	val.Visit(r.ValueOf(`two`), r.StructField{})
	val.Visit(r.ValueOf(``), r.StructField{})

	eq(t, []string{`one`, `two`}, val.Interface())
}

func TestGetWalker_nil(t *testing.T) {
	eq(t, nil, GetWalker(nil, nil))
	eq(t, nil, GetWalker(r.TypeOf(``), nil))
	eq(t, nil, GetWalker(nil, True{}))
}

func TestGetWalker_caching(t *testing.T) {
	filter0 := TypeFilter{r.TypeOf(string(``))}
	filter1 := TypeFilter{r.TypeOf(int(0))}
	typeOuter := r.TypeOf(Outer{})
	typeInner := r.TypeOf(Inner{})

	is(t, GetWalker(typeOuter, filter0), GetWalker(typeOuter, filter0))
	is(t, GetWalker(typeOuter, filter1), GetWalker(typeOuter, filter1))
	is(t, GetWalker(typeInner, filter0), GetWalker(typeInner, filter0))
	is(t, GetWalker(typeInner, filter1), GetWalker(typeInner, filter1))
}

func Test_walking(t *testing.T) {
	{
		vis := AppenderFor((*string)(nil))

		Walk(testOuterVal, vis.Filter(), vis)

		eq(
			t,
			[]string{`embed val`, `embed ptr val`, `outer val`, `inner val`, `inner ptr val`, `outer iface`},
			vis.Interface(),
		)
	}

	{
		vis := AppenderFor((*int)(nil))

		Walk(testOuterVal, vis.Filter(), vis)

		eq(
			t,
			[]int{10, 20, 30, 40},
			vis.Interface(),
		)
	}
}

func TestMaybeOr(t *testing.T) {
	eq(t, nil, MaybeOr())
	eq(t, nil, MaybeOr(nil, nil, nil))
	eq(t, Nop{}, MaybeOr(Nop{}))
	eq(t, Nop{}, MaybeOr(nil, Nop{}, nil))
	eq(t, Or{Nop{}, True{}}, MaybeOr(nil, Nop{}, nil, True{}))
	eq(t, Or{Nop{}, True{}, False{}}, MaybeOr(nil, Nop{}, nil, True{}, nil, False{}))
}

func TestMaybeAnd(t *testing.T) {
	eq(t, nil, MaybeAnd())
	eq(t, nil, MaybeAnd(nil, nil, nil))
	eq(t, Nop{}, MaybeAnd(Nop{}))
	eq(t, Nop{}, MaybeAnd(nil, Nop{}, nil))
	eq(t, And{Nop{}, True{}}, MaybeAnd(nil, Nop{}, nil, True{}))
	eq(t, And{Nop{}, True{}, False{}}, MaybeAnd(nil, Nop{}, nil, True{}, nil, False{}))
}
