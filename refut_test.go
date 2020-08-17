package refut

import (
	"reflect"
	"testing"
)

type Composite struct {
	Exported0   string
	Exported1   string
	unexported0 string
	unexported1 string
	Embedded
	Named    Named
	NamedPtr *NamedPtr
}

type Embedded struct {
	Exported2   string
	Exported3   string
	unexported2 string
	unexported3 string
}

type Named struct {
	Exported4 string
	Exported5 string
}

type NamedPtr struct {
	Exported6 string
	Exported7 string
}

type Field struct {
	Name  string
	Value interface{}
}

func makeComposite() Composite {
	val := Composite{NamedPtr: &NamedPtr{}}
	val.Exported0 = "Exported0"
	val.Exported1 = "Exported1"
	val.unexported0 = "unexported0"
	val.unexported1 = "unexported1"
	val.Exported2 = "Exported2"
	val.Exported3 = "Exported3"
	val.unexported2 = "unexported2"
	val.unexported3 = "unexported3"
	val.Named.Exported4 = "Exported4"
	val.Named.Exported5 = "Exported5"
	val.NamedPtr.Exported6 = "Exported6"
	val.NamedPtr.Exported7 = "Exported7"
	return val
}

func TestTraverseStruct(t *testing.T) {
	value := makeComposite()

	var fields []Field

	err := TraverseStruct(value, func(rval reflect.Value, sfield reflect.StructField, path []int) error {
		fields = append(fields, Field{Name: sfield.Name, Value: rval.Interface()})
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	expected := []Field{
		{"Exported0", value.Exported0},
		{"Exported1", value.Exported1},
		{"Exported2", value.Exported2},
		{"Exported3", value.Exported3},
		{"Named", value.Named},
		{"NamedPtr", value.NamedPtr},
	}

	if !reflect.DeepEqual(expected, fields) {
		t.Fatalf(`
Expected to visit the following fields:
%#v
Got the following:
%#v
`, expected, fields)
	}
}

func TestTraverseStructType(t *testing.T) {
	var fieldNames []string

	err := TraverseStructType(Composite{}, func(sfield reflect.StructField, path []int) error {
		fieldNames = append(fieldNames, sfield.Name)
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"Exported0", "Exported1", "Exported2", "Exported3", "Named", "NamedPtr"}

	if !reflect.DeepEqual(expected, fieldNames) {
		t.Fatalf(`
Expected to visit the following fields:
%#v
Got the following:
%#v
`, expected, fieldNames)
	}
}

func TestTagIdent(t *testing.T) {
	type Tagged struct {
		Empty           string `json:""`
		Hyphen          string `json:"-"`
		Ident           string `json:"ident"`
		IdentOmitempty  string `json:"ident,omitempty"`
		EmptyOmitempty  string `json:",omitempty"`
		HyphenOmitempty string `json:"-,omitempty"`
	}

	rtype := reflect.TypeOf(Tagged{})

	test := func(fieldName string, expected string) {
		sfield, ok := rtype.FieldByName(fieldName)
		if !ok {
			t.Fatalf(`field %q not found`, fieldName)
		}
		tag := sfield.Tag.Get("json")
		ident := TagIdent(tag)
		if expected != ident {
			t.Fatalf(`tag: %q; expected ident: %q; received ident: %q`, tag, expected, ident)
		}
	}

	test("Empty", "")
	test("Hyphen", "")
	test("Ident", "ident")
	test("IdentOmitempty", "ident")
	test("EmptyOmitempty", "")
	test("HyphenOmitempty", "")
}

func TestRtypeDeref(t *testing.T) {
	test := func(actual reflect.Type, expected reflect.Type) {
		if expected != actual {
			t.Fatalf(`expected RtypeDeref to produce type %v, got %v`, expected, actual)
		}
	}

	rt := reflect.TypeOf

	test(RtypeDeref(rt("")), rt(""))
	test(RtypeDeref(rt((***string)(nil))), rt(""))
	test(RtypeDeref(rt(nil)), rt(nil))
}

func TestRvalDeref(t *testing.T) {
	test := func(actual, expected interface{}) {
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf(`expected RtypeDeref to produce %#v, got %#v`, expected, actual)
		}
	}

	rv := reflect.ValueOf

	val0 := "hello world"
	val1 := &val0
	val2 := &val1

	test(RvalDeref(rv(val0)).Interface(), "hello world")
	test(RvalDeref(rv(val1)).Interface(), "hello world")
	test(RvalDeref(rv(val2)).Interface(), "hello world")
}

// This automatically tests `IsRkindNilable`.
func TestIsNilable(t *testing.T) {
	test := func(val interface{}, expected bool) {
		actual := IsNilable(val)
		if expected != actual {
			t.Fatalf(`expected IsNilable(%#v) = %v, got %v`, val, expected, actual)
		}
	}

	test(interface{}(nil), true)
	test((chan string)(nil), true)
	test((func())(nil), true)
	test(map[string]string(nil), true)
	test([]string(nil), true)
	test((*string)(nil), true)

	test(0, false)
	test("", false)
	test([2]string{}, false)
	test(struct{}{}, false)
	test(false, false)
}

// This automatically tests `IsRvalNil`.
func TestIsNil(t *testing.T) {
	test := func(val interface{}, expected bool) {
		actual := IsNil(val)
		if expected != actual {
			t.Fatalf(`expected IsNil(%#v) = %v, got %v`, val, expected, actual)
		}
	}

	str := ""

	test(interface{}(nil), true)
	test((chan string)(nil), true)
	test((func())(nil), true)
	test(map[string]string(nil), true)
	test([]string(nil), true)
	test((*string)(nil), true)

	test(interface{}(""), false)
	test(make(chan string, 0), false)
	test(func() {}, false)
	test(map[string]string{}, false)
	test([]string{}, false)
	test(&str, false)

	test(0, false)
	test("", false)
	test([2]string{}, false)
	test(struct{}{}, false)
	test(false, false)
}

// This automatically tests `IsRvalColl`.
func TestIsColl(t *testing.T) {
	test := func(val interface{}, expected bool) {
		actual := IsColl(val)
		if expected != actual {
			t.Fatalf(`expected IsColl(%#v) = %v, got %v`, val, expected, actual)
		}
	}

	test([2]string{}, true)
	test((chan string)(nil), true)
	test(map[string]string(nil), true)
	test([]string(nil), true)
	test("", true)

	test(interface{}(nil), false)
	test(0, false)
	test(struct{}{}, false)
	test(false, false)
	test((*string)(nil), false)
	test((func())(nil), false)
}

// This automatically tests `IsRvalEmptyColl`.
func TestIsEmptyColl(t *testing.T) {
	test := func(val interface{}, expected bool) {
		actual := IsEmptyColl(val)
		if expected != actual {
			t.Fatalf(`expected IsEmptyColl(%#v) = %v, got %v`, val, expected, actual)
		}
	}

	test([0]string{}, true)
	test((chan string)(nil), true)
	test(make(chan string, 1), true)
	test(map[string]string(nil), true)
	test(map[string]string{}, true)
	test([]string(nil), true)
	test([]string{}, true)
	test("", true)

	test([2]string{}, false)
	test(map[string]string{"": ""}, false)
	test([]string{""}, false)
	test("_", false)

	test(interface{}(nil), false)
	test(0, false)
	test(struct{}{}, false)
	test(false, false)
	test((*string)(nil), false)
	test((func())(nil), false)
}

func BenchmarkTraverseRtypeInline(b *testing.B) {
	value := makeComposite()
	b.ResetTimer()

	for range bn(b) {
		err := traverseRtypeInline(reflect.TypeOf(value))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func traverseRtypeInline(rtype reflect.Type) error {
	rtype = onlyStructRtype(rtype)

	for i := 0; i < rtype.NumField(); i++ {
		sfield := rtype.Field(i)
		if !IsSfieldExported(sfield) {
			continue
		}

		if sfield.Anonymous && RtypeDeref(sfield.Type).Kind() == reflect.Struct {
			err := traverseRtypeInline(sfield.Type)
			if err != nil {
				return err
			}
			continue
		}
	}

	return nil
}

func BenchmarkTraverseStructType(b *testing.B) {
	value := makeComposite()
	b.ResetTimer()

	for range bn(b) {
		err := TraverseStructType(value, func(reflect.StructField, []int) error {
			return nil
		})

		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTraverseStructRtype(b *testing.B) {
	value := makeComposite()
	b.ResetTimer()

	for range bn(b) {
		err := TraverseStructRtype(reflect.TypeOf(value), func(reflect.StructField, []int) error {
			return nil
		})

		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTraverseStruct(b *testing.B) {
	value := makeComposite()
	b.ResetTimer()

	for range bn(b) {
		err := TraverseStruct(value, func(reflect.Value, reflect.StructField, []int) error {
			return nil
		})

		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTraverseStructRval(b *testing.B) {
	value := makeComposite()
	b.ResetTimer()

	for range bn(b) {
		err := TraverseStructRval(reflect.ValueOf(value), func(reflect.Value, reflect.StructField, []int) error {
			return nil
		})

		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReflectValueOf(b *testing.B) {
	value := makeComposite()
	b.ResetTimer()
	for range bn(b) {
		_ = reflect.ValueOf(value)
	}
}

func BenchmarkReflectTypeOf(b *testing.B) {
	value := makeComposite()
	b.ResetTimer()
	for range bn(b) {
		_ = reflect.TypeOf(value)
	}
}

func BenchmarkReflectValueOfType(b *testing.B) {
	value := makeComposite()
	b.ResetTimer()
	for range bn(b) {
		_ = reflect.ValueOf(value).Type()
	}
}

func bn(b *testing.B) []struct{} { return make([]struct{}, b.N) }
