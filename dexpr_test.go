package dexpr

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/dlit_go"
	"math"
	"testing"
)

func TestEval_noerrors(t *testing.T) {
	cases := []struct {
		in   string
		want *dlit.Literal
	}{
		{"1 == 1", makeLit(true)},
		{"1 == 2", makeLit(false)},
		{"2.6 + 2.5", makeLit(5.1)},
		{"-2 + -2", makeLit(-4)},
		{"-2.5 + -2.6", makeLit(-5.1)},
		{"a + numStrB", makeLit(7)},
		{"8/4", makeLit(2)},
		{"1/4", makeLit(0.25)},
		{"8*4", makeLit(32)},
		{fmt.Sprintf("%d * 1", int64(math.MinInt64)),
			makeLit(int64(math.MinInt64))},
		{fmt.Sprintf("%d * 1", int64(math.MaxInt64)),
			makeLit(int64(math.MaxInt64))},
		{fmt.Sprintf("(%d / 2) * 2", int64(math.MinInt64)),
			makeLit(int64(math.MinInt64))},
		{fmt.Sprintf("((%d+-1) / 2) * 2", int64(math.MaxInt64)),
			makeLit(int64(math.MaxInt64) - 1)},

		/* Tests that unary negation works properly */
		{fmt.Sprintf("%d + 0", int64(math.MinInt64)),
			makeLit(int64(math.MinInt64))},

		{"roundto(5.567, 2)", makeLit(5.57)},
		{"roundto(-17.5, 0)", makeLit(-17)},
	}
	vars := map[string]*dlit.Literal{
		"a":       makeLit(4),
		"numStrB": makeLit("3"),
	}
	funcs := map[string]CallFun{
		"roundto": roundTo,
	}
	for _, c := range cases {
		dexpr, err := New(c.in)
		if err != nil {
			t.Errorf("New(%s) err: %s", c.in, err)
		}
		got := dexpr.Eval(vars, funcs)
		if got.IsError() || got.String() != c.want.String() {
			t.Errorf("Eval(vars) in: %q, got: %s, want: %s", c.in, got, c.want)
		}
	}
}

func TestEval_errors(t *testing.T) {
	cases := []struct {
		in   string
		want *dlit.Literal
	}{
		{"8/bob", makeLit(
			ErrInvalidExpr("Variable doesn't exist: bob")),
		},
		{"8/(1 == 1)", makeLit(
			ErrInvalidExpr("Invalid operation: 8 / true")),
		},
		{"8/0", makeLit(
			ErrInvalidExpr("Invalid operation: 8 / 0 (Divide by zero)")),
		},
		{"bob(5.567, 2)", makeLit(
			ErrInvalidExpr("Function doesn't exist: bob")),
		},
		{"9223372036854775807 + 9223372036854775807", makeLit(
			ErrInvalidExpr("Invalid operation: 9223372036854775807 + 9223372036854775807 (Underflow/Overflow)")),
		},
		{fmt.Sprintf("%d+1", int64(math.MaxInt64)), makeLit(
			ErrInvalidExpr("Invalid operation: 9223372036854775807 + 1 (Underflow/Overflow)")),
		},
		{fmt.Sprintf("%d + -1", int64(math.MinInt64)), makeLit(
			ErrInvalidExpr("Invalid operation: -9223372036854775808 + -1 (Underflow/Overflow)")),
		},
		{fmt.Sprintf("%d*2", int64(math.MaxInt64)), makeLit(
			ErrInvalidExpr("Invalid operation: 9223372036854775807 * 2 (Underflow/Overflow)")),
		},
		{fmt.Sprintf("%d*2", int64(math.MinInt64)), makeLit(
			ErrInvalidExpr("Invalid operation: -9223372036854775808 * 2 (Underflow/Overflow)")),
		},
		/* TODO: implement this
		{fmt.Sprintf("%f+1", float64(math.MaxFloat64)), makeLit(
			ErrInvalidExpr("Invalid operation: 9223372036854775807 + 1, Overflow")),
		},
		*/
		// TODO: Add test for overflow - largest int divided by 0.5
		// TODO: Add test for overflow - largest float divided by 0.5
	}
	vars := map[string]*dlit.Literal{
		"a":       makeLit(4),
		"numStrB": makeLit("3"),
	}
	funcs := map[string]CallFun{
		"roundto": roundTo,
	}
	for _, c := range cases {
		dexpr, err := New(c.in)
		if err != nil {
			t.Errorf("New(%s) err: %s", c.in, err)
		}
		got := dexpr.Eval(vars, funcs)
		if got.IsError() != c.want.IsError() || got.String() != c.want.String() {
			t.Errorf("Eval(vars) in: %q, got: %s, want: %s", c.in, got, c.want)
		}
	}
}

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
		{"numStrC == numStrD", false},

		/* Check that keyword tokens are parsed as variables */
		{"break == 1", true},
		{"break == 2", false},
		{"case == 2", true},
		{"case == 3", false},
		{"chan == 3", true},
		{"chan == 4", false},
		{"const == 4", true},
		{"const == 5", false},
		{"continue == 5", true},
		{"continue == 6", false},
		{"default == 6", true},
		{"default == 7", false},
		{"defer == 7", true},
		{"defer == 8", false},
		{"else == 8", true},
		{"else == 9", false},
		{"fallthrough == 9", true},
		{"fallthrough == 10", false},
		{"for == 10", true},
		{"for == 11", false},
		{"func == 11", true},
		{"func == 12", false},
		{"go == 12", true},
		{"go == 13", false},
		{"goto == 13", true},
		{"goto == 14", false},
		{"if == 14", true},
		{"if == 15", false},
		{"import == 15", true},
		{"import == 16", false},
		{"interface == 16", true},
		{"interface == 17", false},
		{"map == 17", true},
		{"map == 18", false},
		{"package == 18", true},
		{"package == 19", false},
		{"range == 19", true},
		{"range == 20", false},
		{"return == 20", true},
		{"return == 21", false},
		{"select == 21", true},
		{"select == 22", false},
		{"struct == 22", true},
		{"struct == 23", false},
		{"switch == 23", true},
		{"switch == 24", false},
		{"type == 24", true},
		{"type == 25", false},
		{"var == 25", true},
		{"var == 26", false},

		{"a != 4", false},
		{"a != 5", true},
		{"a != a", false},
		{"a != b", true},
		{"\"hello\" != \"hello\"", false},
		{"\"hllo\" != \"hello\"", true},
		{"\"hllo\" != 7", true},
		{"str != \"hello\"", false},
		{"str != \"helo\"", true},
		{"numStrA != 3", true},
		{"numStrA != 4", false},
		{"numStrA != numStrA", false},
		{"numStrA != numStrB", true},
		{"numStrC != numStrD", true},

		/* Ensure that bools are not used for comparison */
		{"\"true\" == 1", false},
		{"\"true\" == 1.0", false},
		{"\"true\" == \"TRUE\"", false},
		{"\"TRUE\" == \"TRUE\"", true},
		{"\"false\" == \"FALSE\"", false},
		{"\"FALSE\" == \"FALSE\"", true},
		{"\"false\" ==  0", false},
		{"\"false\" ==  0.0", false},
		{"\"true\" != 0", true},
		{"\"true\" != 1.0", true},
		{"\"true\" != \"TRUE\"", true},
		{"\"TRUE\" != \"TRUE\"", false},
		{"\"false\" != \"FALSE\"", true},
		{"\"FALSE\" != \"FALSE\"", false},
		{"\"false\" !=  0", true},
		{"\"false\" !=  0.0", true},

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
		{"6 >= 7", false},
		{"7 >= 7", true},
		{"8 >= 7", true},
		{"6.7 >= 7", false},
		{"6.7 >= 7.7", false},
		{"7.2 >= 7", true},
		{"7.2 >= 7.2", true},
		{"b >= a", false},
		{"a >= b", true},
		{"3 >= numStrA", false},
		{"4 >= numStrA", true},
		{"5 >= numStrA", true},
		{"5.5 >= numStrA", true},
		{"numStrA >= 5", false},
		{"numStrA >= 4", true},
		{"numStrA >= 3", true},
		{"numStrB >= numStrA", false},
		{"numStrB >= numStrB", true},
		{"numStrA >= numStrB", true},
		{"numStrA >= numStrC", false},
		{"numStrD >= numStrC", false},
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
		{"roundto(8+2.25, 1) == 10.3", true},
		{"roundto(8+2.25, 1) == 10.25", false},

		/*
			{"isFrom(5)", true},
			{"isFrom(true)", true},
		*/
	}
	vars := map[string]*dlit.Literal{
		"a":           makeLit(4),
		"b":           makeLit(3),
		"c":           makeLit(4.5),
		"d":           makeLit(3.5),
		"str":         makeLit("hello"),
		"numStrA":     makeLit("4"),
		"numStrB":     makeLit("3"),
		"numStrC":     makeLit("4.5"),
		"numStrD":     makeLit("3.5"),
		"break":       makeLit(1),
		"case":        makeLit(2),
		"chan":        makeLit(3),
		"const":       makeLit(4),
		"continue":    makeLit(5),
		"default":     makeLit(6),
		"defer":       makeLit(7),
		"else":        makeLit(8),
		"fallthrough": makeLit(9),
		"for":         makeLit(10),
		"func":        makeLit(11),
		"go":          makeLit(12),
		"goto":        makeLit(13),
		"if":          makeLit(14),
		"import":      makeLit(15),
		"interface":   makeLit(16),
		"map":         makeLit(17),
		"package":     makeLit(18),
		"range":       makeLit(19),
		"return":      makeLit(20),
		"select":      makeLit(21),
		"struct":      makeLit(22),
		"switch":      makeLit(23),
		"type":        makeLit(24),
		"var":         makeLit(25),
	}
	funcs := map[string]CallFun{
		"roundto": roundTo,
	}
	for _, c := range cases {
		dexpr, err := New(c.in)
		if err != nil {
			t.Errorf("New(%s) err: %s", c.in, err)
		}
		got, err := dexpr.EvalBool(vars, funcs)
		if err != nil {
			t.Errorf("EvalBool(vars, %q) err == %q", c.in, err)
		}
		if got != c.want {
			t.Errorf("EvalBool(vars, %q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestString(t *testing.T) {
	cases := []string{
		"1 == 1",
		"2.5 == 2.5",
		"1.0 == 1",
		"numStr == 3",
		"\"true\" == \"TRUE\"",
		"5 + 1.5 > 6",
	}
	for _, c := range cases {
		dexpr, err := New(c)
		if err != nil {
			t.Errorf("New(%s) err: %s", c, err)
		}
		got := dexpr.String()
		if got != c {
			t.Errorf("String() got %s, want: %s", got, c)
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
		{"bob(8+2.257) == 7", false,
			ErrInvalidExpr("Function doesn't exist: bob")},
	}
	vars := map[string]*dlit.Literal{}
	funcs := map[string]CallFun{}
	for _, c := range cases {
		dexpr, err := New(c.in)
		if err != nil {
			t.Errorf("New(%s) err: %s", c.in, err)
		}
		got, err := dexpr.EvalBool(vars, funcs)
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

/**********************************
 *    Helper functions
 **********************************/

func makeLit(v interface{}) *dlit.Literal {
	l, err := dlit.New(v)
	if err != nil {
		panic(fmt.Sprintf("MakeLit(%q) gave err: %q", v, err))
	}
	return l
}

func roundTo(args []*dlit.Literal) (*dlit.Literal, error) {
	if len(args) > 2 {
		err := errors.New("Too many arguments")
		return makeLit(err), err
	}
	x, isFloat := args[0].Float()
	if !isFloat {
		err := errors.New("Can't convert to float")
		return makeLit(err), err
	}
	p, isInt := args[1].Int()
	if !isInt {
		err := errors.New("Can't convert to int")
		return makeLit(err), err
	}
	// This uses round half-up to tie-break
	shift := math.Pow(10, float64(p))
	return makeLit(math.Floor(.5+x*shift) / shift), nil
}
