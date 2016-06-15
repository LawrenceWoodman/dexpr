package dexpr

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"math"
	"testing"
)

func TestMustNew(t *testing.T) {
	cases := []struct {
		in      string
		wantStr string
	}{
		{"a+b", "a+b"},
		{"income", "income"},
		{"6.6", "6.6"},
	}

	for _, c := range cases {
		got := MustNew(c.in)
		if got.String() != c.wantStr {
			t.Errorf("MustNew(%q) - got: %s, want: %s", c.in, got, c.wantStr)
		}
	}
}

func TestMustNew_panic(t *testing.T) {
	cases := []struct {
		in        string
		wantPanic string
	}{
		{"/bob harry", ErrInvalidExpr("Invalid expression: /bob harry").Error()},
	}

	for _, c := range cases {
		paniced := false
		defer func() {
			if r := recover(); r != nil {
				if r.(string) == c.wantPanic {
					paniced = true
				} else {
					t.Errorf("MustNew(%q) - got panic: %s, wanted: %s",
						c.in, r, c.wantPanic)
				}
			}
		}()
		MustNew(c.in)
		if c.wantPanic != "" && !paniced {
			t.Errorf("MustNew(%q) - failed to panic with: %s", c.in, c.wantPanic)
		}
	}
}

func TestNew_errors(t *testing.T) {
	cases := []struct {
		in        string
		wantError error
	}{
		{"7 {} 3", ErrInvalidExpr("Invalid expression: 7 {} 3")},
		{"8/cot££t", ErrInvalidExpr("Invalid expression: 8/cot££t")},
	}
	for _, c := range cases {
		_, err := New(c.in)
		if err == nil {
			t.Errorf("New(%s) no error, wanted: %s", c.in, err)
		}
		if err.Error() != c.wantError.Error() {
			t.Errorf("New(%s) got error: %s, wanted: %s", c.in, err, c.wantError)
		}
	}
}

func TestEval_noerrors(t *testing.T) {
	cases := []struct {
		in   string
		want *dlit.Literal
	}{
		{"1 == 1", dlit.MustNew(true)},
		{"1 == 2", dlit.MustNew(false)},
		{"2.6 + 2.5", dlit.MustNew(5.1)},
		{"-2 + -2", dlit.MustNew(-4)},
		{"-2.5 + -2.6", dlit.MustNew(-5.1)},
		{"-2 - 3", dlit.MustNew(-5)},
		{"-2.5 - 3.6", dlit.MustNew(-6.1)},
		{"8 - 9", dlit.MustNew(-1)},
		{"a + numStrB", dlit.MustNew(7)},
		{"8/4", dlit.MustNew(2)},
		{"1/4", dlit.MustNew(0.25)},
		{"8*4", dlit.MustNew(32)},
		{fmt.Sprintf("%d * 1", int64(math.MinInt64)),
			dlit.MustNew(int64(math.MinInt64))},
		{fmt.Sprintf("%d * 1", int64(math.MaxInt64)),
			dlit.MustNew(int64(math.MaxInt64))},
		{fmt.Sprintf("(%d / 2) * 2", int64(math.MinInt64)),
			dlit.MustNew(int64(math.MinInt64))},
		{fmt.Sprintf("((%d+-1) / 2) * 2", int64(math.MaxInt64)),
			dlit.MustNew(int64(math.MaxInt64) - 1)},

		/* Tests that unary negation works properly */
		{fmt.Sprintf("%d + 0", int64(math.MinInt64)),
			dlit.MustNew(int64(math.MinInt64))},

		{"roundto(5.567, 2)", dlit.MustNew(5.57)},
		{"roundto(-17.5, 0)", dlit.MustNew(-17)},
	}
	vars := map[string]*dlit.Literal{
		"a":       dlit.MustNew(4),
		"numStrB": dlit.MustNew("3"),
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
		if _, gotIsErr := got.Err(); gotIsErr || got.String() != c.want.String() {
			t.Errorf("Eval(vars) in: %q, got: %s, want: %s", c.in, got, c.want)
		}
	}
}

func TestEval_errors(t *testing.T) {
	cases := []struct {
		in   string
		want *dlit.Literal
	}{
		{"8/bob", dlit.MustNew(
			ErrInvalidExpr("Variable doesn't exist: bob")),
		},
		{"8/(1 == 1)", dlit.MustNew(
			ErrInvalidExpr("Invalid operation: 8 / true")),
		},
		{"8/0", dlit.MustNew(
			ErrInvalidExpr("Invalid operation: 8 / 0 (Divide by zero)")),
		},
		{"bob(5.567, 2)", dlit.MustNew(
			ErrInvalidExpr("Function doesn't exist: bob")),
		},
		{fmt.Sprintf("%d+%d", int64(math.MaxInt64), int64(math.MaxInt64)),
			dlit.MustNew(
				ErrInvalidExpr("Invalid operation: 9223372036854775807 + 9223372036854775807 (Underflow/Overflow)")),
		},
		{fmt.Sprintf("%d+1", int64(math.MaxInt64)), dlit.MustNew(
			ErrInvalidExpr("Invalid operation: 9223372036854775807 + 1 (Underflow/Overflow)")),
		},
		{fmt.Sprintf("%d + -1", int64(math.MinInt64)), dlit.MustNew(
			ErrInvalidExpr("Invalid operation: -9223372036854775808 + -1 (Underflow/Overflow)")),
		},
		{fmt.Sprintf("%d - %d",
			int64(math.MaxInt64), int64(math.MinInt64)), dlit.MustNew(
			ErrInvalidExpr("Invalid operation: 9223372036854775807 - -9223372036854775808 (Underflow/Overflow)")),
		},
		{fmt.Sprintf("%d - -1", int64(math.MaxInt64)), dlit.MustNew(
			ErrInvalidExpr("Invalid operation: 9223372036854775807 - -1 (Underflow/Overflow)")),
		},
		{fmt.Sprintf("%d - 1", int64(math.MinInt64)), dlit.MustNew(
			ErrInvalidExpr("Invalid operation: -9223372036854775808 - 1 (Underflow/Overflow)")),
		},
		{fmt.Sprintf("%d*2", int64(math.MaxInt64)), dlit.MustNew(
			ErrInvalidExpr("Invalid operation: 9223372036854775807 * 2 (Underflow/Overflow)")),
		},
		{fmt.Sprintf("%d*2", int64(math.MinInt64)), dlit.MustNew(
			ErrInvalidExpr("Invalid operation: -9223372036854775808 * 2 (Underflow/Overflow)")),
		},
		/* TODO: implement this
		{fmt.Sprintf("%f+1", float64(math.MaxFloat64)), dlit.MustNew(
			ErrInvalidExpr("Invalid operation: 9223372036854775807 + 1, Overflow")),
		},
		*/
		// TODO: Add test for overflow - largest int divided by 0.5
		// TODO: Add test for overflow - largest float divided by 0.5
	}
	vars := map[string]*dlit.Literal{
		"a":       dlit.MustNew(4),
		"numStrB": dlit.MustNew("3"),
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
		_, gotIsErr := got.Err()
		_, wantIsErr := c.want.Err()
		if gotIsErr != wantIsErr || got.String() != c.want.String() {
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
		{"9 > 8 || 2 < 3", true},
		{"9 > 9 || 2 < 3", true},
		{"9 > 8 || 3 < 3", true},
		{"9 > 9 || 3 < 3", false},
		{"9 > 8 || 2 < 3 || 7 > 2", true},
		{"8 > 8 || 3 < 3 || 7 > 7", false},
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
		"a":           dlit.MustNew(4),
		"b":           dlit.MustNew(3),
		"c":           dlit.MustNew(4.5),
		"d":           dlit.MustNew(3.5),
		"str":         dlit.MustNew("hello"),
		"numStrA":     dlit.MustNew("4"),
		"numStrB":     dlit.MustNew("3"),
		"numStrC":     dlit.MustNew("4.5"),
		"numStrD":     dlit.MustNew("3.5"),
		"break":       dlit.MustNew(1),
		"case":        dlit.MustNew(2),
		"chan":        dlit.MustNew(3),
		"const":       dlit.MustNew(4),
		"continue":    dlit.MustNew(5),
		"default":     dlit.MustNew(6),
		"defer":       dlit.MustNew(7),
		"else":        dlit.MustNew(8),
		"fallthrough": dlit.MustNew(9),
		"for":         dlit.MustNew(10),
		"func":        dlit.MustNew(11),
		"go":          dlit.MustNew(12),
		"goto":        dlit.MustNew(13),
		"if":          dlit.MustNew(14),
		"import":      dlit.MustNew(15),
		"interface":   dlit.MustNew(16),
		"map":         dlit.MustNew(17),
		"package":     dlit.MustNew(18),
		"range":       dlit.MustNew(19),
		"return":      dlit.MustNew(20),
		"select":      dlit.MustNew(21),
		"struct":      dlit.MustNew(22),
		"switch":      dlit.MustNew(23),
		"type":        dlit.MustNew(24),
		"var":         dlit.MustNew(25),
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

func roundTo(args []*dlit.Literal) (*dlit.Literal, error) {
	if len(args) > 2 {
		err := errors.New("Too many arguments")
		return dlit.MustNew(err), err
	}
	x, isFloat := args[0].Float()
	if !isFloat {
		err := errors.New("Can't convert to float")
		return dlit.MustNew(err), err
	}
	p, isInt := args[1].Int()
	if !isInt {
		err := errors.New("Can't convert to int")
		return dlit.MustNew(err), err
	}
	// This uses round half-up to tie-break
	shift := math.Pow(10, float64(p))
	return dlit.MustNew(math.Floor(.5+x*shift) / shift), nil
}
