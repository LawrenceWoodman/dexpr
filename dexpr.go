/*
 * A package for evaluating dynamic expressions
 *
 * Copyright (C) 2016-2017 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

package dexpr

import (
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"go/ast"
	"go/token"
	"strconv"
)

type Expr struct {
	Expr string
	Node *ENode
}

type CallFun func([]*dlit.Literal) (*dlit.Literal, error)

func New(expr string, callFuncs map[string]CallFun) (*Expr, error) {
	node, err := parseExpr(expr)
	if err != nil {
		return &Expr{}, InvalidExprError{expr, ErrSyntax}
	}

	en := compile(node, callFuncs)
	if err := en.Err(); err != nil {
		return &Expr{}, InvalidExprError{expr, err}
	}
	return &Expr{Expr: expr, Node: en}, nil
}

func MustNew(expr string, callFuncs map[string]CallFun) *Expr {
	e, err := New(expr, callFuncs)
	if err != nil {
		panic(err.Error())
	}
	return e
}

func (expr *Expr) Eval(vars map[string]*dlit.Literal) *dlit.Literal {
	l := expr.Node.Eval(vars)
	if err := l.Err(); err != nil {
		return dlit.MustNew(InvalidExprError{expr.Expr, err})
	}
	return l
}

func (expr *Expr) EvalBool(vars map[string]*dlit.Literal) (bool, error) {
	l := expr.Eval(vars)
	if b, isBool := l.Bool(); isBool {
		return b, nil
	} else if err := l.Err(); err != nil {
		return false, err
	}
	return false, InvalidExprError{expr.Expr, ErrIncompatibleTypes}
}

func (expr *Expr) String() string {
	return expr.Expr
}

// kinds are the kinds of composite type
var kinds = map[string]*dlit.Literal{
	"lit": dlit.NewString("lit"),
}

// TODO: use kind or separate types rather than isError, isFunction, etc
// TODO: Make private
type ENode struct {
	isError    bool
	isFunction bool
	isLiteral  bool
	isVar      bool
	err        error /* TODO: Consider keeping errors as literals */
	function   Fn
	literal    *dlit.Literal
	varName    string
}

// TODO: Make private
type Fn func(map[string]*dlit.Literal) *dlit.Literal

func (n *ENode) Err() error {
	return n.err
}

func (n *ENode) String() string {
	if n.isLiteral {
		return n.literal.String()
	}
	return ""
}

func (n *ENode) Eval(vars map[string]*dlit.Literal) *dlit.Literal {
	if n.isLiteral {
		return n.literal
	} else if n.isError {
		return dlit.MustNew(n.err)
	} else if n.isFunction {
		return n.function(vars)
	} else if n.isVar {
		if l, ok := vars[n.varName]; !ok {
			return dlit.MustNew(VarNotExistError(n.varName))
		} else {
			return l
		}
	}
	panic("ENode incorrectly configured")
}

func (n *ENode) Int() (int64, bool) {
	if !n.isLiteral {
		return 0, false
	}
	i, isInt := n.literal.Int()
	return i, isInt
}

func compile(node ast.Node, callFuncs map[string]CallFun) *ENode {
	var en *ENode
	inspector := func(n ast.Node) bool {
		eltStore := newEltStore()
		en = nodeToENode(callFuncs, eltStore, n)
		return false
	}
	ast.Inspect(node, inspector)
	if err := en.Err(); err != nil {
		return en
	}
	return en
}

func nodeToENode(
	callFuncs map[string]CallFun,
	eltStore *eltStore,
	n ast.Node,
) *ENode {
	switch x := n.(type) {
	case *ast.BasicLit:
		switch x.Kind {
		case token.INT:
			fallthrough
		case token.FLOAT:
			return &ENode{isLiteral: true, literal: dlit.NewString(x.Value)}
		case token.CHAR:
			fallthrough
		case token.STRING:
			uc, err := strconv.Unquote(x.Value)
			if err != nil {
				return &ENode{isError: true, err: ErrSyntax}
			}
			return &ENode{isLiteral: true, literal: dlit.NewString(uc)}
		}
	case *ast.Ident:
		return &ENode{isVar: true, varName: x.Name}
	case *ast.ParenExpr:
		return nodeToENode(callFuncs, eltStore, x.X)
	case *ast.BinaryExpr:
		return binaryExprToENode(callFuncs, eltStore, x)
	case *ast.UnaryExpr:
		return unaryExprToENode(callFuncs, eltStore, x)
	case *ast.CallExpr:
		args := exprSliceToENodes(callFuncs, eltStore, x.Args)
		return &ENode{
			isFunction: true,
			function: func(vars map[string]*dlit.Literal) *dlit.Literal {
				lits := eNodesToDLiterals(vars, args)
				return callFun(callFuncs, x.Fun, lits)
			},
		}
	case *ast.CompositeLit:
		kindNode := nodeToENode(callFuncs, eltStore, x.Type)
		kind := kindNode.Eval(kinds)
		if kind.String() != "lit" {
			return &ENode{isError: true, err: ErrInvalidCompositeType}
		}
		elts := exprSliceToENodes(callFuncs, eltStore, x.Elts)
		rNum := eltStore.Add(elts)
		return &ENode{isLiteral: true, literal: dlit.MustNew(rNum)}
	case *ast.IndexExpr:
		return indexExprToENode(callFuncs, eltStore, x)
	case *ast.ArrayType:
		return nodeToENode(callFuncs, eltStore, x.Elt)
	}
	return &ENode{isError: true, err: ErrSyntax}
}

func indexExprToENode(
	callFuncs map[string]CallFun,
	eltStore *eltStore,
	ie *ast.IndexExpr,
) *ENode {
	indexX := nodeToENode(callFuncs, eltStore, ie.X)
	indexIndex := nodeToENode(callFuncs, eltStore, ie.Index)

	if indexX.Err() != nil {
		return indexX
	} else if indexIndex.Err() != nil {
		return indexIndex
	}
	ii, isInt := indexIndex.Int()
	if !isInt {
		return &ENode{isError: true, err: ErrSyntax}
	}
	if bl, ok := ie.X.(*ast.BasicLit); ok {
		if bl.Kind != token.STRING {
			return &ENode{isError: true, err: ErrTypeNotIndexable}
		}
		return &ENode{
			isLiteral: true,
			literal:   dlit.MustNew(string(indexX.String()[ii])),
		}
	}

	ix, isInt := indexX.Int()
	if !isInt {
		return &ENode{isError: true, err: ErrSyntax}
	}
	elts := eltStore.Get(ix)
	if ii >= int64(len(elts)) {
		return &ENode{isError: true, err: ErrInvalidIndex}
	}
	return elts[ii]
}

func exprSliceToENodes(
	callFuncs map[string]CallFun,
	eltStore *eltStore,
	callArgs []ast.Expr,
) []*ENode {
	r := make([]*ENode, len(callArgs))
	for i, arg := range callArgs {
		r[i] = nodeToENode(callFuncs, eltStore, arg)
	}
	return r
}

func eNodesToDLiterals(
	vars map[string]*dlit.Literal,
	ens []*ENode,
) []*dlit.Literal {
	r := make([]*dlit.Literal, len(ens))
	for i, en := range ens {
		r[i] = en.Eval(vars)
	}
	return r
}

func callFun(
	callFuncs map[string]CallFun,
	name ast.Expr,
	args []*dlit.Literal,
) *dlit.Literal {
	id, ok := name.(*ast.Ident)
	if !ok {
		panic(fmt.Errorf("can't get name as *ast.Ident: %s", name))
	}
	f, exists := callFuncs[id.Name]
	if !exists {
		return dlit.MustNew(FunctionNotExistError(id.Name))
	}
	l, err := f(args)
	if err != nil {
		return dlit.MustNew(FunctionError{id.Name, err})
	}
	return l
}
