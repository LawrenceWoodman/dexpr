package dexpr

import (
	"errors"
	"github.com/lawrencewoodman/dlit"
	"testing"
)

/*************************
       Benchmarks
*************************/
func BenchmarkOpEql(b *testing.B) {
	b.StopTimer()
	for n := 0; n < b.N; n++ {
		cases := []struct {
			lh   *dlit.Literal
			rh   *dlit.Literal
			want *dlit.Literal
		}{
			{lh: dlit.MustNew(5), rh: dlit.MustNew(5), want: dlit.MustNew(true)},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(7), want: dlit.MustNew(false)},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(5.3), want: dlit.MustNew(false)},
			{lh: dlit.MustNew(5),
				rh:   dlit.MustNew("fred"),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew(5.3), rh: dlit.MustNew(5.3), want: dlit.MustNew(true)},
			{lh: dlit.MustNew(6.3),
				rh:   dlit.MustNew(5.3),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew(6.3),
				rh:   dlit.MustNew(5.3),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew(6.3), rh: dlit.MustNew(5), want: dlit.MustNew(false)},
			{lh: dlit.MustNew(6.3),
				rh:   dlit.MustNew("fred"),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew("fred"),
				rh:   dlit.MustNew("fred"),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew("fred"),
				rh:   dlit.MustNew("bob"),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew("fred"),
				rh:   dlit.MustNew(8),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew("fred"),
				rh:   dlit.MustNew(8.2),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew(errors.New("an error")),
				rh:   dlit.MustNew(errors.New("an error")),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(errors.New("an error")),
				rh:   dlit.MustNew(errors.New("another error")),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(errors.New("an error")),
				rh:   dlit.MustNew(5),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(errors.New("an error")),
				rh:   dlit.MustNew(5.3),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(errors.New("an error")),
				rh:   dlit.MustNew("bob"),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(errors.New("an error")),
				rh:   dlit.MustNew("an error"),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
		}
		for _, c := range cases {
			b.StartTimer()
			got := opEql(c.lh, c.rh)
			b.StopTimer()
			if got.String() != c.want.String() {
				b.Errorf("opEql(%s, %s) - got: %s, want: %s", c.lh, c.rh, got, c.want)
			}
		}
	}
}

func BenchmarkOpNeq(b *testing.B) {
	b.StopTimer()
	for n := 0; n < b.N; n++ {
		cases := []struct {
			lh   *dlit.Literal
			rh   *dlit.Literal
			want *dlit.Literal
		}{
			{lh: dlit.MustNew(5), rh: dlit.MustNew(5), want: dlit.MustNew(false)},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(7), want: dlit.MustNew(true)},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(5.3), want: dlit.MustNew(true)},
			{lh: dlit.MustNew(5),
				rh:   dlit.MustNew("fred"),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(5.3), rh: dlit.MustNew(5.3), want: dlit.MustNew(false)},
			{lh: dlit.MustNew(6.3),
				rh:   dlit.MustNew(5.3),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(6.3),
				rh:   dlit.MustNew(5.3),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(6.3), rh: dlit.MustNew(5), want: dlit.MustNew(true)},
			{lh: dlit.MustNew(6.3),
				rh:   dlit.MustNew("fred"),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew("fred"),
				rh:   dlit.MustNew("fred"),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew("fred"),
				rh:   dlit.MustNew("bob"),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew("fred"),
				rh:   dlit.MustNew(8),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew("fred"),
				rh:   dlit.MustNew(8.2),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(errors.New("an error")),
				rh:   dlit.MustNew(errors.New("an error")),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(errors.New("an error")),
				rh:   dlit.MustNew(errors.New("another error")),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(errors.New("an error")),
				rh:   dlit.MustNew(5),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(errors.New("an error")),
				rh:   dlit.MustNew(5.3),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(errors.New("an error")),
				rh:   dlit.MustNew("bob"),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(errors.New("an error")),
				rh:   dlit.MustNew("an error"),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
		}
		for _, c := range cases {
			b.StartTimer()
			got := opNeq(c.lh, c.rh)
			b.StopTimer()
			if got.String() != c.want.String() {
				b.Errorf("opNeq(%s, %s) - got: %s, want: %s", c.lh, c.rh, got, c.want)
			}
		}
	}
}

func BenchmarkOpLand(b *testing.B) {
	b.StopTimer()
	for n := 0; n < b.N; n++ {
		cases := []struct {
			lh   *dlit.Literal
			rh   *dlit.Literal
			want *dlit.Literal
		}{
			{lh: dlit.MustNew(5), rh: dlit.MustNew(5),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(true), rh: dlit.MustNew(5),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(true),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(true), rh: dlit.MustNew(true),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(false), rh: dlit.MustNew(true),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew(false), rh: dlit.MustNew(false),
				want: dlit.MustNew(false),
			},
		}
		for _, c := range cases {
			b.StartTimer()
			got := opLand(c.lh, c.rh)
			b.StopTimer()
			if got.String() != c.want.String() {
				b.Errorf("opLand(%s, %s) - got: %s, want: %s", c.lh, c.rh, got, c.want)
			}
		}
	}
}

func BenchmarkOpLor(b *testing.B) {
	b.StopTimer()
	for n := 0; n < b.N; n++ {
		cases := []struct {
			lh   *dlit.Literal
			rh   *dlit.Literal
			want *dlit.Literal
		}{
			{lh: dlit.MustNew(5), rh: dlit.MustNew(5),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(true), rh: dlit.MustNew(5),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(true),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(true), rh: dlit.MustNew(true),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(false), rh: dlit.MustNew(true),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(true), rh: dlit.MustNew(false),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(false), rh: dlit.MustNew(false),
				want: dlit.MustNew(false),
			},
		}
		for _, c := range cases {
			b.StartTimer()
			got := opLor(c.lh, c.rh)
			b.StopTimer()
			if got.String() != c.want.String() {
				b.Errorf("opLor(%s, %s) - got: %s, want: %s", c.lh, c.rh, got, c.want)
			}
		}
	}
}

func BenchmarkOpLss(b *testing.B) {
	b.StopTimer()
	for n := 0; n < b.N; n++ {
		cases := []struct {
			lh   *dlit.Literal
			rh   *dlit.Literal
			want *dlit.Literal
		}{
			{lh: dlit.MustNew(5), rh: dlit.MustNew(6),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(5.5),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(5),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew(6), rh: dlit.MustNew(5),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew(5.5), rh: dlit.MustNew(5),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew(true), rh: dlit.MustNew(5),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(true),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
		}
		for _, c := range cases {
			b.StartTimer()
			got := opLss(c.lh, c.rh)
			b.StopTimer()
			if got.String() != c.want.String() {
				b.Errorf("opLss(%s, %s) - got: %s, want: %s", c.lh, c.rh, got, c.want)
			}
		}
	}
}

func BenchmarkOpLeq(b *testing.B) {
	b.StopTimer()
	for n := 0; n < b.N; n++ {
		cases := []struct {
			lh   *dlit.Literal
			rh   *dlit.Literal
			want *dlit.Literal
		}{
			{lh: dlit.MustNew(5), rh: dlit.MustNew(6),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(5.5),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(5),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(5.5), rh: dlit.MustNew(5.5),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(6), rh: dlit.MustNew(5),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew(5.5), rh: dlit.MustNew(5),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew(true), rh: dlit.MustNew(5),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(true),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
		}
		for _, c := range cases {
			b.StartTimer()
			got := opLeq(c.lh, c.rh)
			b.StopTimer()
			if got.String() != c.want.String() {
				b.Errorf("opLeq(%s, %s) - got: %s, want: %s", c.lh, c.rh, got, c.want)
			}
		}
	}
}

func BenchmarkOpGtr(b *testing.B) {
	b.StopTimer()
	for n := 0; n < b.N; n++ {
		cases := []struct {
			lh   *dlit.Literal
			rh   *dlit.Literal
			want *dlit.Literal
		}{
			{lh: dlit.MustNew(5), rh: dlit.MustNew(6),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(5.5),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(5),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew(5.5), rh: dlit.MustNew(5.5),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew(6), rh: dlit.MustNew(5),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(5.5), rh: dlit.MustNew(5),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(5.5), rh: dlit.MustNew(5.1),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(true), rh: dlit.MustNew(5),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(true),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
		}
		for _, c := range cases {
			b.StartTimer()
			got := opGtr(c.lh, c.rh)
			b.StopTimer()
			if got.String() != c.want.String() {
				b.Errorf("opGtr(%s, %s) - got: %s, want: %s", c.lh, c.rh, got, c.want)
			}
		}
	}
}

func BenchmarkOpGeq(b *testing.B) {
	b.StopTimer()
	for n := 0; n < b.N; n++ {
		cases := []struct {
			lh   *dlit.Literal
			rh   *dlit.Literal
			want *dlit.Literal
		}{
			{lh: dlit.MustNew(5), rh: dlit.MustNew(6),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(5.5),
				want: dlit.MustNew(false),
			},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(5),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(5.5), rh: dlit.MustNew(5.5),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(6), rh: dlit.MustNew(5),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(5.5), rh: dlit.MustNew(5),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(5.5), rh: dlit.MustNew(5.1),
				want: dlit.MustNew(true),
			},
			{lh: dlit.MustNew(true), rh: dlit.MustNew(5),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
			{lh: dlit.MustNew(5), rh: dlit.MustNew(true),
				want: dlit.MustNew(ErrIncompatibleTypes),
			},
		}
		for _, c := range cases {
			b.StartTimer()
			got := opGeq(c.lh, c.rh)
			b.StopTimer()
			if got.String() != c.want.String() {
				b.Errorf("opGeq(%s, %s) - got: %s, want: %s", c.lh, c.rh, got, c.want)
			}
		}
	}
}
