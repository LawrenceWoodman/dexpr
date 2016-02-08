/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 */
package dexpr

import (
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"strconv"
)

type Expr struct {
	Expr string
	Node ast.Node
}

type ErrInvalidExpr string

func (e ErrInvalidExpr) Error() string {
	return string(e)
}

type ErrInvalidOp token.Token

func (e ErrInvalidOp) Error() string {
	return fmt.Sprintf("Invalid operator: %q", token.Token(e))
}

func New(expr string) (*Expr, error) {
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return &Expr{}, err
	}
	return &Expr{Expr: expr, Node: node}, nil
}

func (expr *Expr) EvalBool(vars map[string]*dlit.Literal) (bool, error) {
	var l *dlit.Literal
	inspector := func(n ast.Node) bool {
		l = nodeToLiteral(vars, n)
		return false
	}
	ast.Inspect(expr.Node, inspector)
	if b, isBool := l.Bool(); isBool {
		return b, nil
	} else if l.IsError() {
		return false, ErrInvalidExpr(l.String())
	} else {
		return false, ErrInvalidExpr("Expression doesn't return a bool")
	}
}

func nodeToLiteral(vars map[string]*dlit.Literal, n ast.Node) *dlit.Literal {
	var l *dlit.Literal
	var err error
	switch x := n.(type) {
	case *ast.BasicLit:
		switch x.Kind {
		case token.INT:
			l, err = dlit.New(x.Value)
			if err != nil {
				l, _ = dlit.New(err)
			}
		case token.FLOAT:
			l, err = dlit.New(x.Value)
			if err != nil {
				l, _ = dlit.New(err)
			}

		case token.STRING:
			uc, err := strconv.Unquote(x.Value)
			if err != nil {
				l, _ = dlit.New(err)
			}
			l, err = dlit.New(uc)
			if err != nil {
				l, _ = dlit.New(err)
			}
		}
	case *ast.Ident:
		l = vars[x.Name]
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

func evalBinaryExpr(lh *dlit.Literal, rh *dlit.Literal,
	op token.Token) *dlit.Literal {
	var r *dlit.Literal

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
		r, _ = dlit.New(ErrInvalidOp(op))
	}

	return r
}

func literalToQuotedString(l *dlit.Literal) string {
	_, isInt := l.Int()
	if isInt {
		return l.String()
	}
	_, isFloat := l.Float()
	if isFloat {
		return l.String()
	}
	_, isBool := l.Bool()
	if isBool {
		return l.String()
	}
	return fmt.Sprintf("\"%s\"", l.String())
}

func makeErrInvalidExprLiteral(errFormattedMsg string, l1 *dlit.Literal,
	l2 *dlit.Literal) *dlit.Literal {
	l1s := literalToQuotedString(l1)
	l2s := literalToQuotedString(l2)
	err := ErrInvalidExpr(fmt.Sprintf(errFormattedMsg, l1s, l2s))
	r, _ := dlit.New(err)
	return r
}

func checkNewLitError(l *dlit.Literal, err error, errFormattedMsg string,
	l1 *dlit.Literal, l2 *dlit.Literal) *dlit.Literal {
	if err != nil {
		return makeErrInvalidExprLiteral(errFormattedMsg, l1, l2)
	}
	return l
}

func opLss(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid comparison: %s < %s"

	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		l, err := dlit.New(lhInt < rhInt)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	rhFloat, rhIsFloat := rh.Float()
	if lhIsInt && rhIsFloat {
		l, err := dlit.New(lhInt < int64(math.Ceil(rhFloat)))
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat && rhIsInt {
		l, err := dlit.New(int64(math.Floor(lhFloat)) < rhInt)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	if lhIsFloat && rhIsFloat {
		l, err := dlit.New(lhFloat < rhFloat)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	return makeErrInvalidExprLiteral(errMsg, lh, rh)
}

func opLeq(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid comparison: %s <= %s"
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		l, err := dlit.New(lhInt <= rhInt)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	rhFloat, rhIsFloat := rh.Float()
	if lhIsInt && rhIsFloat {
		l, err := dlit.New(lhInt <= int64(math.Floor(rhFloat)))
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat && rhIsInt {
		l, err := dlit.New(int64(math.Floor(lhFloat)) <= rhInt)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	if lhIsFloat && rhIsFloat {
		l, err := dlit.New(lhFloat <= rhFloat)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	return makeErrInvalidExprLiteral(errMsg, lh, rh)
}

func opGtr(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid comparison: %s > %s"

	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		l, err := dlit.New(lhInt > rhInt)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	rhFloat, rhIsFloat := rh.Float()
	if lhIsInt && rhIsFloat {
		l, err := dlit.New(lhInt > int64(math.Floor(rhFloat)))
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat && rhIsInt {
		l, err := dlit.New(int64(math.Ceil(lhFloat)) > rhInt)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	if lhIsFloat && rhIsFloat {
		l, err := dlit.New(lhFloat > rhFloat)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	return makeErrInvalidExprLiteral(errMsg, lh, rh)
}

func opEql(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid comparison: %s == %s"

	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		l, err := dlit.New(lhInt == rhInt)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		l, err := dlit.New(lhFloat == rhFloat)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	// Don't compare bools as otherwise with the way that floats or ints
	// are cast to bools you would find that "True" == 1.0 because they would
	// both convert to true bools

	if lh.IsError() || rh.IsError() {
		return makeErrInvalidExprLiteral(errMsg, lh, rh)
	}

	l, err := dlit.New(lh.String() == rh.String())
	return checkNewLitError(l, err, errMsg, lh, rh)
}

func opAdd(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid operation: %s + %s"

	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		l, err := dlit.New(lhInt + rhInt)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		l, err := dlit.New(lhFloat + rhFloat)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}
	return makeErrInvalidExprLiteral(errMsg, lh, rh)
}

func opLand(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid operation: %s && %s"

	lhBool, lhIsBool := lh.Bool()
	rhBool, rhIsBool := rh.Bool()
	if lhIsBool && rhIsBool {
		l, err := dlit.New(lhBool && rhBool)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	return makeErrInvalidExprLiteral(errMsg, lh, rh)
}
