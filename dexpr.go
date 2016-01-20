/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 */
package dexpr

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"math"
	"strconv"
)

// TODO: Remove all log references
type Literal struct {
	Value string
	Kind  Kind
}

type Kind int

type ErrInvalidKind struct {
	expected Kind
	got      Kind
}

func (e ErrInvalidKind) Error() string {
	return fmt.Sprintf("Invalid kind, expected: %s, got: %s", e.expected, e.got)
}

type ErrInvalidKinds struct {
	expected []Kind
	got      []Kind
}

type ErrInvalidExpr string

func (e ErrInvalidExpr) Error() string {
	return string(e)
}

func (e ErrInvalidKinds) Error() string {
	return fmt.Sprintf("Invalid kinds, expected: %s, got: %s", e.expected, e.got)
}

type ErrInvalidOp token.Token

func (e ErrInvalidOp) Error() string {
	return fmt.Sprintf("Invalid operator: %q", token.Token(e))
}

const (
	Illegal Kind = iota
	Error
	Int
	Float
	String
	Bool
)

var kinds = [...]string{
	Illegal: "Illegal",
	Error:   "Error",
	Int:     "Int",
	Float:   "Float",
	String:  "String",
	Bool:    "Bool",
}

func (k Kind) String() string {
	var s string
	if k >= 0 && k < Kind(len(kinds)) {
		s = kinds[k]
	} else {
		s = fmt.Sprintf("kind(%q)", k)
	}
	return s
}

func newLiteralError(e error) Literal {
	return Literal{Value: e.Error(), Kind: Error}
}

func newLiteralBool(b bool) Literal {
	if b {
		return Literal{Value: "1", Kind: Bool}
	} else {
		return Literal{Value: "0", Kind: Bool}
	}
}

func (l *Literal) ValueAsFloat() float64 {
	var err error
	var f float64
	if l.Kind == Float || l.Kind == Int {
		f, err = strconv.ParseFloat(l.Value, 64)
		if err != nil {
			log.Fatal("Can't parse as float: ", l.Value)
		}
	} else {
		log.Fatal("Incompatible type: ", l.Kind)
	}
	return f
}

func (l *Literal) ValueAsInt() int64 {
	var err error
	var i int64
	if l.Kind == Int {
		i, err = strconv.ParseInt(l.Value, 10, 64)
		if err != nil {
			log.Fatal("Can't parse as int: ", l.Value)
		}
	} else {
		log.Fatal("Incompatible type: ", l.Kind)
	}
	return i
}

func (l *Literal) ValueAsBool() bool {
	var b bool
	if l.Value == "1" {
		b = true
	} else if l.Value == "0" {
		b = false
	} else {
		log.Fatal("Unrecognized value: ", l.Value)
	}
	return b
}

func EvalBool(vars map[string]Literal, node ast.Node) (bool, error) {
	var l Literal
	inspector := func(n ast.Node) bool {
		l = nodeToLiteral(vars, n)
		return false
	}
	ast.Inspect(node, inspector)
	if l.Kind == Bool {
		return l.ValueAsBool(), nil
	} else if l.Kind == Error {
		return false, ErrInvalidExpr(l.Value)
	} else {
		return false, ErrInvalidExpr("Expression doesn't return a bool")
	}
}

func nodeToLiteral(vars map[string]Literal, n ast.Node) Literal {
	var l Literal
	var k Kind
	switch x := n.(type) {
	case *ast.BasicLit:
		switch x.Kind {
		case token.INT:
			k = Int
			l = Literal{Value: x.Value, Kind: k}
		case token.FLOAT:
			k = Float
			l = Literal{Value: x.Value, Kind: k}
		case token.STRING:
			uc, err := strconv.Unquote(x.Value)
			if err != nil {
				log.Fatal(err)
			}
			k = String
			l = Literal{Value: uc, Kind: k}
		default:
			k = Illegal
			l = Literal{Value: "", Kind: k}
		}
	case *ast.Ident:
		l = Literal{Value: vars[x.Name].Value, Kind: vars[x.Name].Kind}
	case *ast.ParenExpr:
		l = nodeToLiteral(vars, x.X)
	case *ast.BinaryExpr:
		lh := nodeToLiteral(vars, x.X)
		rh := nodeToLiteral(vars, x.Y)
		l = evalBinaryExpr(lh, rh, x.Op)
	case *ast.CallExpr:
		fmt.Printf("CallExpr - expr: %q, args: %q\n", x.Fun, x.Args)
	default:
		fmt.Println("UNRECOGNIZED TYPE - x: ", x)
	}
	return l
}

func evalBinaryExpr(lh Literal, rh Literal, op token.Token) Literal {
	var r Literal

	switch op {
	case token.LSS:
		r = opLss(lh, rh)
	case token.LEQ:
		r = opLeq(lh, rh)
	case token.EQL:
		r = opEql(lh, rh)
	case token.GTR:
		r = opGtr(lh, rh)
	case token.ADD:
		r = opAdd(lh, rh)
	case token.LAND:
		r = opLand(lh, rh)
	default:
		err := ErrInvalidOp(op)
		return newLiteralError(err)
	}

	return r
}

func opLss(lh Literal, rh Literal) Literal {
	var b bool
	switch true {
	case lh.Kind == Int && rh.Kind == Int:
		b = lh.ValueAsInt() < rh.ValueAsInt()
	case lh.Kind == Int && rh.Kind == Float:
		b = lh.ValueAsInt() < int64(math.Ceil(rh.ValueAsFloat()))
	case lh.Kind == Float && rh.Kind == Int:
		b = int64(math.Floor(lh.ValueAsFloat())) < rh.ValueAsInt()
	case lh.Kind == Float && rh.Kind == Float:
		b = lh.ValueAsFloat() < rh.ValueAsFloat()
	default:
		expected := []Kind{Int, Float}
		got := []Kind{lh.Kind, rh.Kind}
		err := ErrInvalidKinds{expected, got}
		return newLiteralError(err)
	}
	return newLiteralBool(b)
}

func opLeq(lh Literal, rh Literal) Literal {
	var b bool
	switch true {
	case lh.Kind == Int && rh.Kind == Int:
		b = lh.ValueAsInt() <= rh.ValueAsInt()
	case lh.Kind == Int && rh.Kind == Float:
		b = lh.ValueAsInt() <= int64(math.Floor(rh.ValueAsFloat()))
	case lh.Kind == Float && rh.Kind == Int:
		b = int64(math.Floor(lh.ValueAsFloat())) <= rh.ValueAsInt()
	case lh.Kind == Float && rh.Kind == Float:
		b = lh.ValueAsFloat() <= rh.ValueAsFloat()
	default:
		expected := []Kind{Int, Float}
		got := []Kind{lh.Kind, rh.Kind}
		err := ErrInvalidKinds{expected, got}
		return newLiteralError(err)
	}
	return newLiteralBool(b)
}

func opGtr(lh Literal, rh Literal) Literal {
	var b bool
	switch true {
	case lh.Kind == Int && rh.Kind == Int:
		b = lh.ValueAsInt() > rh.ValueAsInt()
	case lh.Kind == Int && rh.Kind == Float:
		b = lh.ValueAsInt() > int64(math.Floor(rh.ValueAsFloat()))
	case lh.Kind == Float && rh.Kind == Int:
		b = int64(math.Ceil(lh.ValueAsFloat())) > rh.ValueAsInt()
	case lh.Kind == Float && rh.Kind == Float:
		b = lh.ValueAsFloat() > rh.ValueAsFloat()
	default:
		expected := []Kind{Int, Float}
		got := []Kind{lh.Kind, rh.Kind}
		err := ErrInvalidKinds{expected, got}
		return newLiteralError(err)
	}
	return newLiteralBool(b)
}

func opEql(lh Literal, rh Literal) Literal {
	var b bool
	switch true {
	case lh.Kind == Int && rh.Kind == Int:
		b = lh.ValueAsInt() == rh.ValueAsInt()
	case lh.Kind == Float && rh.Kind == Float:
		b = lh.ValueAsFloat() == rh.ValueAsFloat()
	case lh.Kind == String && rh.Kind == String:
		b = lh.Value == rh.Value
	default:
		b = false
	}
	return newLiteralBool(b)
}

func opAdd(lh Literal, rh Literal) Literal {
	var r Literal
	switch true {
	case lh.Kind == Int && rh.Kind == Int:
		i := lh.ValueAsInt() + rh.ValueAsInt()
		r = Literal{Value: strconv.FormatInt(i, 10), Kind: Int}
	case lh.Kind == Int && rh.Kind == Float:
		fallthrough
	case lh.Kind == Float && rh.Kind == Int:
		f := lh.ValueAsFloat() + rh.ValueAsFloat()
		r = Literal{Value: strconv.FormatFloat(f, 'E', -1, 64), Kind: Float}
	case lh.Kind == Float && rh.Kind == Float:
		f := lh.ValueAsFloat() + rh.ValueAsFloat()
		r = Literal{Value: strconv.FormatFloat(f, 'E', -1, 64), Kind: Float}
	default:
		expected := []Kind{Int, Float}
		got := []Kind{lh.Kind, rh.Kind}
		err := ErrInvalidKinds{expected, got}
		return newLiteralError(err)
	}
	return r
}

func opLand(lh Literal, rh Literal) Literal {
	if lh.Kind != Bool {
		err := ErrInvalidKind{Bool, lh.Kind}
		return newLiteralError(err)
	}
	if rh.Kind != Bool {
		err := ErrInvalidKind{Bool, rh.Kind}
		return newLiteralError(err)
	}
	return newLiteralBool(lh.ValueAsBool() && rh.ValueAsBool())
}
