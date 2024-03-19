// Copyright 2024 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package ast

import (
	"cee"
	"github.com/langvm/go-cee-scanner"
)

const (
	_ = iota
)

type Node interface {
	GetPosRange() PosRange
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

type List[T any] struct {
	PosRange
	List []T
}

type TypeKind byte

const (
	_ TypeKind = iota

	TypeNone
	TypeStruct
	TypeTrait

	TypeI8 // builtin
	TypeI16
	TypeI32
	TypeI64
	TypeU8
	TypeU16
	TypeU32
	TypeU64
)

type Type struct {
	cee.Union[TypeKind]
}

type (
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

type ExprKind int

const (
	_ = iota

	ExprIdent
	ExprLiteralValue
	ExprUnary
	ExprBinary
)

type Expr struct {
	cee.Union[ExprKind]
}

type (
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

type StmtKind byte

const (
	_ = iota
)

type Stmt struct {
}

type (
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
