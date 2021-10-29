package rf

import (
	r "reflect"
	"testing"
)

func BenchmarkGetWalker(b *testing.B) {
	for range Iter(b.N) {
		benchGetWalker()
	}
}

func benchGetWalker() {
	GetWalker(r.TypeOf(&testValOuter), True{})
}

func BenchmarkWalk(b *testing.B) {
	for range Iter(b.N) {
		benchWalk()
	}
}

func benchWalk() {
	Walk(r.ValueOf(&testValOuter), True{}, Nop{})
}
