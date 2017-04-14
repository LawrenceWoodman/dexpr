/*
 * Copyright (C) 2017 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

package dexpr

import (
	"github.com/lawrencewoodman/dlit"
	"go/ast"
	"go/token"
	"math"
	"strconv"
)

func unaryExprToENode(
	callFuncs map[string]CallFun,
	eltStore *eltStore,
	ue *ast.UnaryExpr,
) *ENode {
	rh := nodeToENode(callFuncs, eltStore, ue.X)
	if rh.Err() != nil {
		return rh
	}
	switch ue.Op {
	case token.NOT:
		return &ENode{
			isFunction: true,
			function: func(vars map[string]*dlit.Literal) *dlit.Literal {
				return opNot(rh.Eval(vars))
			},
		}
	case token.SUB:
		return &ENode{
			isFunction: true,
			function: func(vars map[string]*dlit.Literal) *dlit.Literal {
				return opNeg(rh.Eval(vars))
			},
		}
	}
	return &ENode{isError: true, err: InvalidOpError(ue.Op)}
}

func opNot(l *dlit.Literal) *dlit.Literal {
	lBool, lIsBool := l.Bool()
	if !lIsBool {
		return dlit.MustNew(ErrIncompatibleTypes)
	}
	if lBool {
		return falseLiteral
	}
	return trueLiteral
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
