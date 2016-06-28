/*
 * A package for evaluating dynamic expressions
 *
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */
package dexpr

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"go/ast"
	"go/token"
	"math"
	"strconv"
)

var ErrUnderflowOverflow = errors.New("underflow/overflow")
var ErrDivByZero = errors.New("divide by zero")
var ErrIncompatibleTypes = errors.New("incompatible types")
var ErrInvalidCompositeType = errors.New("invalid composite type")
var ErrInvalidIndex = errors.New("index out of range")
var ErrTypeNotIndexable = errors.New("type does not support indexing")
var ErrSyntax = errors.New("syntax error")

type Expr struct {
	Expr string
	Node ast.Node
}

type ErrInvalidExpr struct {
	Expr string
	Err  error
}

func (e ErrInvalidExpr) Error() string {
	return fmt.Sprintf("invalid expression: %s (%s)", e.Expr, e.Err)
}

type ErrInvalidOp token.Token

func (e ErrInvalidOp) Error() string {
	return fmt.Sprintf("invalid operator: %s", token.Token(e))
}

type ErrFunctionNotExist string

func (e ErrFunctionNotExist) Error() string {
	return fmt.Sprintf("function doesn't exist: %s", string(e))
}

type ErrVarNotExist string

func (e ErrVarNotExist) Error() string {
	return fmt.Sprintf("variable doesn't exist: %s", string(e))
}

type ErrFunctionError struct {
	FnName string
	Err    error
}

func (e ErrFunctionError) Error() string {
	return fmt.Sprintf("function: %s, returned error: %s", e.FnName, e.Err)
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
		lh := nodeToLiteral(vars, callFuncs, eltStore, eltStoreNum, x.X)
		rh := nodeToLiteral(vars, callFuncs, eltStore, eltStoreNum, x.Y)
		if err := lh.Err(); err != nil {
			return lh
		} else if err := rh.Err(); err != nil {
			return rh
		}
		return evalBinaryExpr(lh, rh, x.Op)
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
	}
	if indexIndex.Err() != nil {
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
	case token.NEQ:
		r = opNeq(lh, rh)
	case token.GTR:
		r = opGtr(lh, rh)
	case token.GEQ:
		r = opGeq(lh, rh)
	case token.LAND:
		r = opLand(lh, rh)
	case token.LOR:
		r = opLor(lh, rh)
	case token.ADD:
		r = opAdd(lh, rh)
	case token.SUB:
		r = opSub(lh, rh)
	case token.MUL:
		r = opMul(lh, rh)
	case token.QUO:
		r = opQuo(lh, rh)
	default:
		r, _ = dlit.New(ErrInvalidOp(op))
	}

	return r
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

func opLss(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		return dlit.MustNew(lhInt < rhInt)
	}

	rhFloat, rhIsFloat := rh.Float()
	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat < rhFloat)
	}

	return dlit.MustNew(ErrIncompatibleTypes)
}

func opLeq(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		return dlit.MustNew(lhInt <= rhInt)
	}

	rhFloat, rhIsFloat := rh.Float()
	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat <= rhFloat)
	}

	return dlit.MustNew(ErrIncompatibleTypes)
}

func opGtr(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		return dlit.MustNew(lhInt > rhInt)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat > rhFloat)
	}

	return dlit.MustNew(ErrIncompatibleTypes)
}

func opGeq(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		return dlit.MustNew(lhInt >= rhInt)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat >= rhFloat)
	}

	return dlit.MustNew(ErrIncompatibleTypes)
}

func opEql(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		return dlit.MustNew(lhInt == rhInt)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat == rhFloat)
	}

	// Don't compare bools as otherwise with the way that floats or ints
	// are cast to bools you would find that "True" == 1.0 because they would
	// both convert to true bools
	lhErr := lh.Err()
	rhErr := rh.Err()
	if lhErr != nil || rhErr != nil {
		return dlit.MustNew(ErrIncompatibleTypes)
	}

	return dlit.MustNew(lh.String() == rh.String())
}

func opNeq(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		return dlit.MustNew(lhInt != rhInt)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		dlit.MustNew(lhFloat != rhFloat)
	}

	// Don't compare bools as otherwise with the way that floats or ints
	// are cast to bools you would find that "True" == 1.0 because they would
	// both convert to true bools
	lhErr := lh.Err()
	rhErr := rh.Err()
	if lhErr != nil || rhErr != nil {
		return dlit.MustNew(ErrIncompatibleTypes)
	}

	return dlit.MustNew(lh.String() != rh.String())
}

func opLand(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhBool, lhIsBool := lh.Bool()
	rhBool, rhIsBool := rh.Bool()
	if lhIsBool && rhIsBool {
		return dlit.MustNew(lhBool && rhBool)
	}

	return dlit.MustNew(ErrIncompatibleTypes)
}

func opLor(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhBool, lhIsBool := lh.Bool()
	rhBool, rhIsBool := rh.Bool()
	if lhIsBool && rhIsBool {
		return dlit.MustNew(lhBool || rhBool)
	}

	return dlit.MustNew(ErrIncompatibleTypes)
}

func opAdd(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		r := lhInt + rhInt
		if (r < lhInt) != (rhInt < 0) {
			return dlit.MustNew(ErrUnderflowOverflow)
		}
		return dlit.MustNew(r)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat + rhFloat)
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}

func opSub(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		r := lhInt - rhInt
		if (r > lhInt) != (rhInt < 0) {
			return dlit.MustNew(ErrUnderflowOverflow)
		}
		return dlit.MustNew(r)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat - rhFloat)
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}

func opMul(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		// Overflow detection inspired by suggestion from Rob Pike on Go-nuts group:
		//   https://groups.google.com/d/msg/Golang-nuts/h5oSN5t3Au4/KaNQREhZh0QJ
		if lhInt == 0 || rhInt == 0 || lhInt == 1 || rhInt == 1 {
			return dlit.MustNew(lhInt * rhInt)
		}
		if lhInt == math.MinInt64 || rhInt == math.MinInt64 {
			return dlit.MustNew(ErrUnderflowOverflow)
		}
		r := lhInt * rhInt
		if r/rhInt != lhInt {
			return dlit.MustNew(ErrUnderflowOverflow)
		}
		return dlit.MustNew(r)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat * rhFloat)
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}

func opQuo(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()

	if rhIsInt && rhInt == 0 {
		return dlit.MustNew(ErrDivByZero)
	}
	if lhIsInt && rhIsInt && lhInt%rhInt == 0 {
		return dlit.MustNew(lhInt / rhInt)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat / rhFloat)
	}
	return dlit.MustNew(ErrIncompatibleTypes)
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
