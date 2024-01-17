// Copyright 2023-2024 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package ast

import . "cee/internal"

func (t Token) Print(b *StringBuffer) {
	b.Print(t.Literal)
}

func (t StructType) Print(b *StringBuffer) {
	b.Println("struct {")
	for _, field := range t.Fields {
		field.Print(b)
	}
	b.Println("}")
}

func (t TraitType) Print(b *StringBuffer) {
	b.Println("trait {")
	// TODO
	b.Println("}")
}

func (t FuncType) Print(b *StringBuffer) {
	b.Println("(")
	for _, param := range t.Params {
		for _, ident := range param.Idents {
			b.Println(ident.Literal, ",")
		}
		param.Type.Print(b)
	}
	b.Println(")(")
	for _, result := range t.Results {
		result.Print(b)
	}
	b.Println(")")
}

func (e LiteralValue) Print(b *StringBuffer) {
	b.Print(e.Literal)
}

func (i Ident) Print(b *StringBuffer) {
	i.Token.Print(b)
}

func (e UnaryExpr) Print(b *StringBuffer) {
	e.Operator.Print(b)
	e.Expr.Print(b)
}

func (e BinaryExpr) Print(b *StringBuffer) {
	e.Exprs[0].Print(b)
	e.Operator.Print(b)
	e.Exprs[1].Print(b)
}

func (e CallExpr) Print(b *StringBuffer) {
	e.Callee.Print(b)
	b.Println("(")
	for _, param := range e.Params {
		param.Print(b)
		b.Println(",")
	}
	b.Println(")")
}

func (e IndexExpr) Print(b *StringBuffer) {
	e.Expr.Print(b)
	b.Print("[")
	e.Index.Print(b)
	b.Print("]")
}

func (e MemberSelectExpr) Print(b *StringBuffer) {
	e.Expr.Print(b)
	b.Print(".")
	e.Member.Print(b)
}

func (d GenDecl) Print(b *StringBuffer) {
	for _, ident := range d.Idents {
		b.Println(ident.Literal, ",")
	}
	d.Type.Print(b)
}

func (d FuncDecl) Print(b *StringBuffer) {
	if d.Ident == nil {
		b.Print("fun ")
	} else {
		b.Println("fun ", d.Ident.Literal, " ")
	}
	d.Type.Print(b)
	d.Stmt.Print(b)
}

func (e StmtBlockExpr) Print(b *StringBuffer) {
	b.Println("{")
	for _, stmt := range e.Stmts {
		stmt.Print(b)
	}
	b.Println("}")
}
