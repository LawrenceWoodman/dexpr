/*
 * A package for evaluating dynamic expressions
 *
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

package dexpr

import (
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"go/ast"
	"go/token"
	"math"
	"strconv"
)

type Expr struct {
	Expr string
	Node ast.Node
}

type CallFun func([]*dlit.Literal) (*dlit.Literal, error)

func New(expr string) (*Expr, error) {
	node, err := parseExpr(expr)
	if err != nil {
		return &Expr{}, ErrInvalidExpr{expr, ErrSyntax}
	}
	return &Expr{Expr: expr, Node: node}, nil
}

func MustNew(expr string) *Expr {
	e, err := New(expr)
	if err != nil {
		panic(err.Error())
	}
	return e
}

func (expr *Expr) Eval(
	vars map[string]*dlit.Literal,
	callFuncs map[string]CallFun,
) *dlit.Literal {
	var l *dlit.Literal
	inspector := func(n ast.Node) bool {
		// The eltStore is where the composite types store there elements
		eltStore := map[int64][]*dlit.Literal{}
		eltStoreNum := int64(0)
		l = nodeToLiteral(vars, callFuncs, eltStore, eltStoreNum, n)
		return false
	}
	ast.Inspect(expr.Node, inspector)
	if err := l.Err(); err != nil {
		return dlit.MustNew(ErrInvalidExpr{expr.Expr, err})
	}
	return l
}

func (expr *Expr) EvalBool(
	vars map[string]*dlit.Literal,
	callFuncs map[string]CallFun,
) (bool, error) {
	l := expr.Eval(vars, callFuncs)
	if b, isBool := l.Bool(); isBool {
		return b, nil
	} else if err := l.Err(); err != nil {
		return false, err
	}
	return false, ErrInvalidExpr{expr.Expr, ErrIncompatibleTypes}
}

func (expr *Expr) String() string {
	return expr.Expr
}

var kinds = map[string]*dlit.Literal{
	"lit": dlit.NewString("lit"),
}

func nodeToLiteral(
	vars map[string]*dlit.Literal,
	callFuncs map[string]CallFun,
	eltStore map[int64][]*dlit.Literal,
	eltStoreNum int64,
	n ast.Node,
) *dlit.Literal {
	switch x := n.(type) {
	case *ast.BasicLit:
		switch x.Kind {
		case token.INT:
			return dlit.MustNew(x.Value)
		case token.FLOAT:
			return dlit.MustNew(x.Value)
		case token.CHAR:
			fallthrough
		case token.STRING:
			uc, err := strconv.Unquote(x.Value)
			if err != nil {
				return dlit.MustNew(ErrSyntax)
			}
			return dlit.NewString(uc)
		}
	case *ast.Ident:
		if l, exists := vars[x.Name]; !exists {
			return dlit.MustNew(ErrVarNotExist(x.Name))
		} else {
			return l
		}
	case *ast.ParenExpr:
		return nodeToLiteral(vars, callFuncs, eltStore, eltStoreNum, x.X)
	case *ast.BinaryExpr:
		return binaryExprToLiteral(vars, callFuncs, eltStore, eltStoreNum, x)
	case *ast.UnaryExpr:
		rh := nodeToLiteral(vars, callFuncs, eltStore, eltStoreNum, x.X)
		if err := rh.Err(); err != nil {
			return rh
		}
		return evalUnaryExpr(rh, x.Op)
	case *ast.CallExpr:
		args := exprSliceToDLiterals(vars, callFuncs, eltStore, eltStoreNum, x.Args)
		return callFun(callFuncs, x.Fun, args)
	case *ast.CompositeLit:
		kind := nodeToLiteral(kinds, callFuncs, eltStore, eltStoreNum, x.Type)
		if kind.String() != "lit" {
			return dlit.MustNew(ErrInvalidCompositeType)
		}
		elts := exprSliceToDLiterals(vars, callFuncs, eltStore, eltStoreNum, x.Elts)
		eltStore[eltStoreNum] = elts
		rNum := eltStoreNum
		eltStoreNum++
		return dlit.MustNew(rNum)
	case *ast.IndexExpr:
		return indexExprToLiteral(vars, callFuncs, eltStore, eltStoreNum, x)
	case *ast.ArrayType:
		return nodeToLiteral(vars, callFuncs, eltStore, eltStoreNum, x.Elt)
	}
	return dlit.MustNew(ErrSyntax)
}

func indexExprToLiteral(
	vars map[string]*dlit.Literal,
	callFuncs map[string]CallFun,
	eltStore map[int64][]*dlit.Literal,
	eltStoreNum int64,
	ie *ast.IndexExpr,
) *dlit.Literal {
	indexX := nodeToLiteral(vars, callFuncs, eltStore, eltStoreNum, ie.X)
	indexIndex := nodeToLiteral(vars, callFuncs, eltStore, eltStoreNum, ie.Index)

	if indexX.Err() != nil {
		return indexX
	} else if indexIndex.Err() != nil {
		return indexIndex
	}
	ii, isInt := indexIndex.Int()
	if !isInt {
		return dlit.MustNew(ErrSyntax)
	}
	if bl, ok := ie.X.(*ast.BasicLit); ok {
		if bl.Kind != token.STRING {
			return dlit.MustNew(ErrTypeNotIndexable)
		}
		return dlit.NewString(string(indexX.String()[ii]))
	}

	ix, isInt := indexX.Int()
	if !isInt {
		return dlit.MustNew(ErrSyntax)
	}
	if ii >= int64(len(eltStore[ix])) {
		return dlit.MustNew(ErrInvalidIndex)
	}
	return eltStore[ix][ii]
}

func exprSliceToDLiterals(
	vars map[string]*dlit.Literal,
	callFuncs map[string]CallFun,
	eltStore map[int64][]*dlit.Literal,
	eltStoreNum int64,
	callArgs []ast.Expr,
) []*dlit.Literal {
	r := make([]*dlit.Literal, len(callArgs))
	for i, arg := range callArgs {
		r[i] = nodeToLiteral(vars, callFuncs, eltStore, eltStoreNum, arg)
	}
	return r
}

func callFun(
	callFuncs map[string]CallFun,
	name ast.Expr,
	args []*dlit.Literal) *dlit.Literal {
	// TODO: Find more direct way of getting name as a string
	nameString := fmt.Sprintf("%s", name)
	f, exists := callFuncs[nameString]
	if !exists {
		return dlit.MustNew(ErrFunctionNotExist(nameString))
	}
	l, err := f(args)
	if err != nil {
		return dlit.MustNew(ErrFunctionError{nameString, err})
	}
	return l
}

func evalUnaryExpr(rh *dlit.Literal, op token.Token) *dlit.Literal {
	var r *dlit.Literal
	switch op {
	case token.SUB:
		r = opNeg(rh)
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

func opNeg(l *dlit.Literal) *dlit.Literal {
	lInt, lIsInt := l.Int()
	if lIsInt {
		return dlit.MustNew(0 - lInt)
	}

	strMinInt64 := strconv.FormatInt(int64(math.MinInt64), 10)
	posMinInt64 := strMinInt64[1:]
	if l.String() == posMinInt64 {
		return dlit.MustNew(int64(math.MinInt64))
	}

	lFloat, lIsFloat := l.Float()
	if lIsFloat {
		return dlit.MustNew(0 - lFloat)
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}
