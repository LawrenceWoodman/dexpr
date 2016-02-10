package dexpr

import (
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"testing"
)

func TestEvalBool_noErrors(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"1 == 1", true},
		{"1 == 2", false},
		{"2.5 == 2.5", true},
		{"2.5 == 3.5", false},
		{"1 == 1.5", false},

		/* Becareful of float == int comparison */
		{"1.0 == 1", true},
		{"1 == 1.0", true},

		{"numStrB == 3", true},
		{"numStrB == 3.0", true},
		{"3 == numStrB", true},
		{"3.0 == numStrB", true},
		{"a == 4", true},
		{"a == 5", false},
		{"a == a", true},
		{"a == b", false},
		{"\"hello\" == \"hello\"", true},
		{"\"hllo\" == \"hello\"", false},
		{"\"hllo\" == 7", false},
		{"str == \"hello\"", true},
		{"str == \"helo\"", false},
		{"numStrA == 3", false},
		{"numStrA == 4", true},
		{"numStrA == numStrA", true},
		{"numStrA == numStrB", false},
		{"numStrC == numStrC", true},
		{"numStrC == numStrD", false},

		/* Ensure that bools are not used for comparison */
		{"\"true\" == 1", false},
		{"\"true\" == 1.0", false},
		{"\"true\" == \"TRUE\"", false},
		{"\"TRUE\" == \"TRUE\"", true},
		{"\"false\" == \"FALSE\"", false},
		{"\"FALSE\" == \"FALSE\"", true},
		{"\"false\" ==  0", false},
		{"\"false\" ==  0.0", false},

		{"6 < 7", true},
		{"7 < 7", false},
		{"8 < 7", false},
		{"6.7 < 7", true},
		{"6.7 < 7.7", true},
		{"7 < 7.2", true},
		{"7 < 6.7", false},
		{"3 < a", true},
		{"4 < a", false},
		{"a < 5", true},
		{"a < 4", false},
		{"b < a", true},
		{"b < b", false},
		{"a < b", false},
		{"3 < numStrA", true},
		{"4 < numStrA", false},
		{"numStrA < 5", true},
		{"numStrA < 4", false},
		{"numStrB < numStrA", true},
		{"numStrB < numStrB", false},
		{"numStrA < numStrB", false},
		{"numStrA < numStrC", true},
		{"numStrD < numStrC", true},
		{"6 <= 7", true},
		{"7 <= 7", true},
		{"8 <= 7", false},
		{"6.7 <= 7", true},
		{"6.7 <= 7.7", true},
		{"7 <= 7.2", true},
		{"7 <= 6.7", false},
		{"b <= a", true},
		{"a <= a", true},
		{"a <= b", false},
		{"3 <= numStrA", true},
		{"4 <= numStrA", true},
		{"5 <= numStrA", false},
		{"5.5 <= numStrA", false},
		{"numStrA <= 5", true},
		{"numStrA <= 4", true},
		{"numStrA <= 3", false},
		{"numStrB <= numStrA", true},
		{"numStrB <= numStrB", true},
		{"numStrA <= numStrB", false},
		{"numStrA <= numStrC", true},
		{"numStrD <= numStrC", true},
		{"6 > 7", false},
		{"7 > 7", false},
		{"8 > 7", true},
		{"6.7 > 7", false},
		{"6.7 > 7.7", false},
		{"7 > 7.2", false},
		{"b > a", false},
		{"a > b", true},
		{"3 > numStrA", false},
		{"4 > numStrA", false},
		{"5 > numStrA", true},
		{"5.5 > numStrA", true},
		{"numStrA > 5", false},
		{"numStrA > 4", false},
		{"numStrA > 3", true},
		{"numStrB > numStrA", false},
		{"numStrB > numStrB", false},
		{"numStrA > numStrB", true},
		{"numStrA > numStrC", false},
		{"numStrD > numStrC", false},
		{"5 + 1.5 > 6", true},
		{"5 + 1 > 6", false},
		{"a + b > 6", true},
		{"a + b > 7", false},
		{"a + b > 8", false},
		{"numStrA + numStrB > 6", true},
		{"numStrA + numStrB > 7", false},
		{"numStrA + numStrB > 8", false},
		{"numStrC + numStrD > 7", true},
		{"numStrC + numStrD == 8.0", true},
		{"numStrC + numStrD == 8", true},
		{"numStrC + numStrD > 8", false},
		{"9 > 8 && 2 < 3", true},
		{"9 > 9 && 2 < 3", false},
		{"9 > 8 && 3 < 3", false},
		{"9 > 9 && 3 < 3", false},
		{"9 > 8 && 2 < 3 && 7 > 2", true},
		{"9 > 8 && 2 < 3 && 7 > 7", false},
		{"9 + (8 + 2) > 18", true},
		{"9 + (8 + 2) > 19", false},

		/*
			{"isFrom(5)", true},
			{"isFrom(true)", true},
		*/
	}
	vars := map[string]*dlit.Literal{
		"a":       makeLit(4),
		"b":       makeLit(3),
		"c":       makeLit(4.5),
		"d":       makeLit(3.5),
		"str":     makeLit("hello"),
		"numStrA": makeLit("4"),
		"numStrB": makeLit("3"),
		"numStrC": makeLit("4.5"),
		"numStrD": makeLit("3.5"),
	}
	for _, c := range cases {
		dexpr, err := New(c.in)
		got, err := dexpr.EvalBool(vars)
		if err != nil {
			t.Errorf("EvalBool(vars, %q) err == %q", c.in, err)
		}
		if got != c.want {
			t.Errorf("EvalBool(vars, %q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestEvalBool_errors(t *testing.T) {
	cases := []struct {
		in        string
		want      bool
		wantError error
	}{
		{"7 + 8", false, ErrInvalidExpr("Expression doesn't return a bool")},
		{"7 < \"hello\"", false,
			ErrInvalidExpr("Invalid comparison: 7 < \"hello\"")},
		{"\"world\" > 2.1", false,
			ErrInvalidExpr("Invalid comparison: \"world\" > 2.1")},
		{"10 & 101", false, ErrInvalidExpr("Invalid operator: \"&\"")},
		{"7 && 9", false, ErrInvalidExpr("Invalid operation: 7 && 9")},
		{"total > 20", false,
			ErrInvalidExpr("Variable doesn't exist: total")},
		{"20 < total", false,
			ErrInvalidExpr("Variable doesn't exist: total")},
	}
	vars := map[string]*dlit.Literal{}
	for _, c := range cases {
		dexpr, err := New(c.in)
		got, err := dexpr.EvalBool(vars)
		if got != c.want {
			t.Errorf("EvalBool(vars, %q) == %q, want %q", c.in, got, c.want)
		}
		if err == nil {
			t.Errorf("EvalBool(vars, %q) err == nil, wantError %q",
				c.in, c.wantError)
		} else if err.Error() != c.wantError.Error() {
			t.Errorf("EvalBool(vars, %q) err == %q, wantError %q",
				c.in, err, c.wantError)
		}
	}
}

func makeLit(v interface{}) *dlit.Literal {
	l, err := dlit.New(v)
	if err != nil {
		panic(fmt.Sprintf("MakeLit(%q) gave err: %q", v, err))
	}
	return l
}
