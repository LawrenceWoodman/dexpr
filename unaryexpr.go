/*
 * Copyright (C) 2017 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

package dexpr

import (
	"github.com/lawrencewoodman/dlit"
	"go/token"
	"math"
	"strconv"
)

func evalUnaryExpr(rh *dlit.Literal, op token.Token) *dlit.Literal {
	var r *dlit.Literal
	switch op {
	case token.NOT:
		r = opNot(rh)
	case token.SUB:
		r = opNeg(rh)
	default:
		r = dlit.MustNew(InvalidOpError(op))
	}
	return r
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
