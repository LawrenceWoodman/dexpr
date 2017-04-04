package dexpr

import (
	"testing"
)

func TestUse_concurrent(t *testing.T) {
	vs := newValStore()
	numGoRoutines := 100
	for i := 0; i < numGoRoutines; i++ {
		go func() {
			for i := 0; i < 1000; i++ {
				l := vs.Use(string(i))
				if l.String() != string(i) {
					t.Errorf("Use - got: %s, want: %d", l, i)
				}
			}
		}()
	}
}
