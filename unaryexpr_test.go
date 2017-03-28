package dexpr

import (
	"github.com/lawrencewoodman/dlit"
	"testing"
)

/*************************
       Benchmarks
*************************/
func BenchmarkOpNot(b *testing.B) {
	b.StopTimer()
	for n := 0; n < b.N; n++ {
		cases := []struct {
			in   *dlit.Literal
			want *dlit.Literal
		}{
			{in: dlit.MustNew(6), want: dlit.MustNew(ErrIncompatibleTypes)},
			{in: dlit.MustNew(true), want: dlit.MustNew(false)},
			{in: dlit.MustNew(false), want: dlit.MustNew(true)},
		}
		for _, c := range cases {
			b.StartTimer()
			got := opNot(c.in)
			b.StopTimer()
			if got.String() != c.want.String() {
				b.Errorf("opNot(%s) - got: %s, want: %s", c.in, got, c.want)
			}
		}
	}
}
