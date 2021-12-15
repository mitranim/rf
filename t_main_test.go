package rf

import (
	"fmt"
	r "reflect"
	"runtime"
	"strings"
	"testing"
	"unsafe"
)

type Outer struct {
	Embed
	EmbedPtr
	OuterStr   string            `json:"outerStr"   db:"outer_str"`
	Inner      Inner             `json:"inner"      db:"inner"`
	InnerPtr   *Inner            `json:"innerPtr"   db:"inner_ptr"`
	OuterIface interface{}       `json:"outerIface" db:"outer_iface"`
	OuterDict  map[string]string `json:"outerDict"  db:"outer_dict"`
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

var testOuter = Outer{
	Embed:      Embed{EmbedStr: `embed val`, EmbedNum: 10},
	EmbedPtr:   &Embed{EmbedStr: `embed ptr val`, EmbedNum: 20},
	OuterStr:   `outer val`,
	Inner:      Inner{InnerStr: `inner val`, InnerNum: 30},
	InnerPtr:   &Inner{InnerStr: `inner ptr val`, InnerNum: 40},
	OuterIface: `outer iface`,
	OuterDict:  testDict,
}

var testOuterVal = r.ValueOf(testOuter)

var testDict = map[string]string{
	`685d3b7e4d534b4c850e3ddf5701c3ba`: `3adcd66e3c274602a1d6bae0f1398829`,
	`658313fdc97040f09e4ea1d912e7f7e1`: `5b1393f9883a4944ae59d533c5f9320b`,
	`c5b594971e924b63bd9b0a3ea2cdfc09`: `95821e89aaca472eac1ff10720b6d2cd`,
	`7644e53b08a941589ad86fb4764a66b2`: `5ee91be1df594ff68c9d4e4c5f714c9a`,
	`3479510cb4c94447a0cd8b414eea72b7`: `a16d40f182c84ff29785ffd9c3ac73c0`,
	`e0824d8903964a13b8dc3c761ab92566`: `6daf9798230a4e78979f0a10b0be058f`,
	`9480ddd46ef242eb9c7079e9cd45e23a`: `633d04d68b444e27ae82b9f18d5e7797`,
	`479e32fe1b0240d0ae316a5b17391513`: `0875dd0cecd64f98811039d97807f70e`,
	`d60c9adfd28942898ba491a50c72b7d3`: `19a24133a05a4e049957daa96ee52b45`,
	`32fea384071642dbbd6c7256e319eac5`: `bac7dd09e423442c8844bd1d89808c5d`,
	`9638949aed5f43eb9dad81019be12349`: `2769dbf34f064ae6acf07dcf68ff5403`,
	`ac15c9390cb443268dd16ad40385ae17`: `f890f664f85d4b0dbf713d6b84c4f6ef`,
	`4efeea434d12475fb493020f058fa02c`: `17861ccc60d24bc189544856efe295c5`,
	`2072e3349d0b4856ae5ecfb09a9755d2`: `f2c2b99677f4429cbe27e24073e5711e`,
	`ff5f133712fe4b318c430b9ccb81127a`: `11bdedc50e894e1b84af7dde01cc6038`,
	`f076c4dbcbb8443d834568fd701848be`: `ffdfa755a16343bb9f0019d01c5ecfd5`,
	`dce5de9593ed4fad8d077dc7f3bd4ec8`: `b46cd3fef926461e86f182b06e9b0ef9`,
	`dc4233ee223046cd8acd8a430ac61283`: `6c6ba813c4e94ba1bc3f7ab8e59aee4d`,
	`4145068c5d754809b8ac1e8971b9d77e`: `2bfa620589c94dd1bbb4504a0a10d4e7`,
	`11c5e23de6dc4f62914e757f01357f7d`: `525b25113eee45c49f145c9589fe2523`,
}

var testDictVal = r.ValueOf(testDict)

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

	// nolint:structcheck
	type iface struct {
		typ unsafe.Pointer
		dat unsafe.Pointer
	}

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

func panics(t testing.TB, msg string, fun func()) {
	t.Helper()
	val := catchAny(fun)

	if val == nil {
		t.Fatalf(`expected %v to panic, found no panic`, funcName(fun))
	}

	str := fmt.Sprint(val)
	if !strings.Contains(str, msg) {
		t.Fatalf(`
expected %v to panic with a message containing:
	%v
found the following message:
	%v
`, funcName(fun), msg, str)
	}
}

func funcName(val interface{}) string {
	return runtime.FuncForPC(r.ValueOf(val).Pointer()).Name()
}

func catchAny(fun func()) (val interface{}) {
	defer recAny(&val)
	fun()
	return
}

func recAny(ptr *interface{}) { *ptr = recover() }

func stringPtr(val string) *string { return &val }
func intPtr(val int) *int          { return &val }

var testSlice = func() (out []Outer) {
	for range Iter(1024) {
		out = append(out, testOuter)
	}
	return
}()

type PanicVis struct{}

func (self PanicVis) Visit(r.Value, r.StructField) {
	panic(fmt.Errorf(`unexpected call to %q`, funcName(self.Visit)))
}
