package dexpr

import (
	"go/parser"
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
		{"1.0 == 1", true},
		{"a == 4", true},
		{"a == 5", false},
		{"a == a", true},
		{"a == b", false},
		{"\"hello\" == \"hello\"", true},
		{"\"hllo\" == \"hello\"", false},
		{"\"hllo\" == 7", false},
		{"str == \"hello\"", true},
		{"str == \"helo\"", false},
		{"6 < 7", true},
		{"7 < 7", false},
		{"8 < 7", false},
		{"6.7 < 7", true},
		{"6.7 < 7.7", true},
		{"7 < 7.2", true},
		{"7 < 6.7", false},
		{"b < a", true},
		{"b < b", false},
		{"a < b", false},
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
		{"6 > 7", false},
		{"7 > 7", false},
		{"8 > 7", true},
		{"6.7 > 7", false},
		{"6.7 > 7.7", false},
		{"7 > 7.2", false},
		{"b > a", false},
		{"a > b", true},
		{"5 + 1.5 > 6", true},
		{"5 + 1 > 6", false},
		{"a + b > 6", true},
		{"a + b > 7", false},
		{"a + b > 8", false},
		{"9 > 8 && 2 < 3", true},
		{"9 > 9 && 2 < 3", false},
		{"9 > 8 && 3 < 3", false},
		{"9 > 9 && 3 < 3", false},
		{"9 > 8 && 2 < 3 && 7 > 2", true},
		{"9 > 8 && 2 < 3 && 7 > 7", false},
		{"9 + (8 + 2) > 18", true},
		{"9 + (8 + 2) > 19", false},
		{"isFrom(5)", true},
		{"isFrom(true)", true},
	}
	vars := map[string]Literal{
		"a":   Literal{Value: "4", Kind: Int},
		"b":   Literal{Value: "3", Kind: Int},
		"str": Literal{Value: "hello", Kind: String},
	}
	for _, c := range cases {
		node, err := parser.ParseExpr(c.in)
		if err != nil {
			t.Error(err)
		}
		got, err := EvalBool(vars, node)
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
			ErrInvalidExpr("Invalid kinds, expected: [Int Float], got: [Int String]")},
		{"10 & 101", false, ErrInvalidExpr("Invalid operator: \"&\"")},
	}
	vars := map[string]Literal{}
	for _, c := range cases {
		node, err := parser.ParseExpr(c.in)
		if err != nil {
			t.Error(err)
		}
		got, err := EvalBool(vars, node)
		if got != c.want {
			t.Errorf("EvalBool(vars, %q) == %q, want %q", c.in, got, c.want)
		}
		if err.Error() != c.wantError.Error() {
			t.Errorf("EvalBool(vars, %q) err == %q, wantError %q",
				c.in, err, c.wantError)
		}
	}
}
