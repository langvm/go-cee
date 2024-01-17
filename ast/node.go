// Copyright 2023-2024 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package ast

import (
	"cee/internal"
	"cee/scanner"
)

const (
	_ = iota

	UNARY_BEGIN

	OP_NOT

	OP_INCREASE
	OP_DECREASE

	OP_REFERENCE
	OP_DEREFERENCE
	OP_DEREFERENCE_GUARDED

	OP_PANIC

	UNARY_END

	BINARY_BEGIN

	OP_NEQ
	OP_EQ
	OP_LT
	OP_LE

	OP_AND
	OP_OR
	OP_XOR
	OP_SHIFT_LEFT
	OP_SHIFT_RIGHT

	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV
	OP_MOD

	BINARY_END
)

type Node interface {
	GetPosRange() PosRange
	Print(b *internal.StringBuffer)
}

type PosRange struct {
	From, To scanner.Position
}

func (pos PosRange) GetPosRange() PosRange { return pos }

type Token struct {
	PosRange
	Kind    int
	Literal string
}

type (
	Type interface {
		Node
	}

	StructType struct {
		PosRange
		Fields []GenDecl
	}

	TraitType struct {
		PosRange
		// TODO
	}

	TypeAlias struct {
		Ident
	}

	FuncType struct {
		PosRange
		Params  []GenDecl
		Results []Type
	}
)

type (
	Expr interface {
		Node
	}

	BadExpr struct {
		PosRange
	}

	LiteralValue struct {
		Token
	}

	Ident struct {
		Token
	}

	UnaryExpr struct {
		PosRange
		Operator Token
		Expr     Expr
	}

	BinaryExpr struct {
		PosRange
		Operator Token
		Exprs    [2]Expr
	}

	EllipsisExpr struct {
		PosRange
		Array Expr
	}

	CallExpr struct {
		PosRange
		Callee Expr
		Params []Expr
	}

	IndexExpr struct {
		PosRange
		Expr  Expr
		Index Expr
	}

	CastExpr struct {
		PosRange
	}

	BranchExpr struct {
		PosRange
		Cond       Expr
		Branch     StmtBlockExpr
		ElseBranch StmtBlockExpr
	}

	MatchExpr struct {
		PosRange
		Subject  Expr
		Patterns []StmtBlockExpr
	}

	StmtBlockExpr struct {
		PosRange
		Type  Type // nil for void
		Stmts []Stmt
	}

	MemberSelectExpr struct {
		PosRange
		Member Ident
		Expr   Expr
	}
)

type (
	Stmt interface {
		Node
	}

	ImportDecl struct {
		PosRange
		CanonicalName LiteralValue
		Alias         *Ident
	}

	ValDecl struct {
		PosRange
		Name  Ident
		Value Expr
	}

	GenDecl struct {
		PosRange
		Idents []Ident
		Type   Type
	}

	FuncDecl struct {
		PosRange
		Type  FuncType
		Ident *Ident
		Stmt  *StmtBlockExpr
	}

	ReturnStmt struct {
		PosRange
		Exprs []Expr
	}

	AssignStmt struct {
		PosRange
		ExprL, ExprR Expr
	}

	BreakStmt struct {
		PosRange
	}

	ContinueStmt struct {
		PosRange
	}

	LoopStmt struct {
		PosRange
		Cond Expr
		Stmt StmtBlockExpr
	}

	ForeachStmt struct {
		PosRange
		IdentList []Ident
		Expr      Expr
	}

	EndlessForStmt struct {
		Stmt StmtBlockExpr
	}
)
