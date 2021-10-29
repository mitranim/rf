package rf

import (
	r "reflect"
	"testing"
	"unsafe"
)

type Outer struct {
	Embed
	EmbedPtr
	OuterStr string `json:"outerStr" db:"outer_str"`
	Inner    Inner  `json:"inner"    db:"inner"`
	InnerPtr *Inner `json:"innerPtr" db:"inner_ptr"`
}

type Inner struct {
	InnerStr string `json:"innerStr" db:"inner_str"`
	InnerNum int    `json:"innerNum" db:"inner_num"`
}

type EmbedPtr = *Embed

type Embed struct {
	EmbedStr string `json:"embedStr" db:"embed_str"`
	EmbedNum int    `json:"embedNum" db:"embed_num"`
}

var testValOuter = r.ValueOf(Outer{
	Embed:    Embed{EmbedStr: `embed val`, EmbedNum: 10},
	EmbedPtr: &Embed{EmbedStr: `embed ptr val`, EmbedNum: 20},
	OuterStr: `outer val`,
	Inner:    Inner{InnerStr: `inner val`, InnerNum: 30},
	InnerPtr: &Inner{InnerStr: `inner ptr val`, InnerNum: 40},
})

func eq(t testing.TB, exp, act interface{}) {
	t.Helper()
	if !r.DeepEqual(exp, act) {
		t.Fatalf(`
expected (detailed):
	%#[1]v
actual (detailed):
	%#[2]v
expected (simple):
	%[1]v
actual (simple):
	%[2]v
`, exp, act)
	}
}

func is(t testing.TB, exp, act interface{}) {
	t.Helper()

	expIface := *(*iface)(unsafe.Pointer(&exp))
	actIface := *(*iface)(unsafe.Pointer(&act))

	if expIface != actIface {
		t.Fatalf(`
expected (interface):
	%#[1]v
actual (interface):
	%#[2]v
expected (detailed):
	%#[3]v
actual (detailed):
	%#[4]v
expected (simple):
	%[3]v
actual (simple):
	%[4]v
`, expIface, actIface, exp, act)
	}
}

// nolint:structcheck
type iface struct {
	typ unsafe.Pointer
	dat unsafe.Pointer
}

func stringPtr(val string) *string { return &val }
func intPtr(val int) *int          { return &val }
