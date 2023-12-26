// Copyright 2023-2023 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package parser

import (
	. "cee/ast"
	"cee/token"
)

// NOTICE:
// ALL Expect* functions start from the cursor position and end at the NEXT token.

type Parser struct {
	Scanner
}

func (p *Parser) MatchTerms(terms ...int) {
	for _, term := range terms {
		if p.Token.Kind != term {
			panic(UnexpectedNodeError{
				Node:   p.Token,
				Expect: terms})
		}
		p.Scan()
	}
}

func (p *Parser) ExpectToken() Token {
	tok := p.Token
	p.Scan()
	return tok
}

func (p *Parser) ExpectIdent() Ident {
	// Ident -> {Ident.Kind == IDENT}

	if p.Token.Kind != token.IDENT {
		panic(UnexpectedNodeError{
			Node:   p.Token,
			Expect: []int{token.IDENT}})
	}

	ident := p.Token
	p.Scan()

	return Ident{Token: ident}
}

func (p *Parser) ExpectIdentList(terminator int) (idents []Ident) {
	// IdentList -> Ident
	//            | IdentList + ',' + Ident
	//            | IdentList + ',' + terminator

	for {
		ident := p.ExpectIdent()
		idents = append(idents, ident)
		switch p.Token.Kind {
		case token.COMMA:
			p.Scan()
			if p.Token.Kind == terminator {
				// NOTE: INTENTIONAL DESIGN-BREAKING IMPLEMENTATION
				// To avoid this out-of-duty branch:
				// - Ident prefetch must be applied to the scanner for comma erasing, or
				// - Modify syntax, deny any comma at the end of list.
				return
			}
			continue
		default:
			return
		}
	}
}

func (p *Parser) ExpectFuncType() FuncType {
	var (
		params  []GenDecl
		results []Type
	)

	p.Scan()
	p.MatchTerms(token.LPAREN)

	if p.Token.Kind == token.IDENT {
		params = p.ExpectGenDeclList(0)
	}

	p.MatchTerms(token.RPAREN)

	switch p.Token.Kind {
	case token.LPAREN:
		p.Scan()
		results = p.ExpectTypeList(token.RPAREN)
		p.MatchTerms(token.RPAREN)
	default:
		results = []Type{p.ExpectFuncType()}
	}

	return FuncType{
		Params:  params,
		Results: results,
	}
}

func (p *Parser) ExpectFuncDecl() FuncDecl {
	begin := p.Position

	typ := p.ExpectFuncType()

	p.ExpectStmtBlockExpr()

	return FuncDecl{
		PosRange: PosRange{From: begin, To: p.Position},
		Type:     typ,
		Name:     "",
		Block:    StmtBlockExpr{},
	}
}

func (p *Parser) ExpectStructType() StructType {
	begin := p.Position

	p.Scan() // 'struct'
	p.MatchTerms(token.LBRACE)

	var fields []GenDecl
	for {
		if p.Token.Kind != token.IDENT {
			break
		}
		idents := p.ExpectIdentList(token.RPAREN)
		if p.Token.Kind == token.SEMICOLON {
			if len(idents) != 1 {
				panic(UnexpectedNodeError{
					Node:   p.Token,
					Expect: []int{}})
			}
			fields = append(fields, GenDecl{Idents: nil, Type: TypeAlias{Ident: idents[0]}})
			p.Scan()
		} else {
			fields = append(fields, GenDecl{Idents: idents, Type: p.ExpectType()})
			p.MatchTerms(token.SEMICOLON)
		}
	}
	p.MatchTerms(token.RBRACE)

	return StructType{
		PosRange: PosRange{From: begin, To: p.Position},
		Fields:   fields,
	}
}

func (p *Parser) ExpectTraitType() TraitType {
	// TODO parse trait
	return TraitType{} // TODO
}

func (p *Parser) ExpectType() Type {
	switch p.Token.Kind {
	case token.STRUCT:
		return p.ExpectStructType()
	case token.TRAIT:
		return p.ExpectTraitType()
	case token.IDENT:
		return TypeAlias{Ident: p.ExpectIdent()}
	default:
		panic(UnexpectedNodeError{
			Node:   p.Token,
			Expect: []int{token.STRUCT, token.TRAIT, token.IDENT}})
	}
}

func (p *Parser) ExpectTypeList(terminator int) (types []Type) {
	// TypeList -> Type
	//           | TypeList + ',' + Type
	//           | TypeList + ',' + terminator

	for {
		types = append(types, p.ExpectType())
		if p.Token.Kind != token.COMMA {
			if p.Token.Kind == terminator {
				return
			}
			return
		}
		p.Scan()
	}
}

func (p *Parser) ExpectOperator() Token {
	if !token.IsOperator(p.Token.Kind) {
		panic(UnexpectedNodeError{
			Node:   p.Token,
			Expect: []int{token.OPERATOR_BEGIN}})
	}
	p.Scan()

	return p.Token
}

func (p *Parser) ExpectLiteralValue() LiteralValue {
	tok := p.Token
	p.Scan()
	return LiteralValue{Token: tok}
}

func (p *Parser) ExpectIndexExpr(expr Expr) IndexExpr {
	// IndexExpr -> Expr + '[' + Expr + ']'

	begin := p.Position

	p.Scan()

	indexExpr := p.ExpectExpr()
	p.MatchTerms(token.RBRACK)

	return IndexExpr{
		PosRange: PosRange{From: begin, To: p.Position},
		Expr:     expr,
		Index:    indexExpr,
	}
}

func (p *Parser) ExpectMemberSelectExpr(expr Expr) MemberSelectExpr {
	// MemberSelectExpr -> Expr + '.' + Ident

	begin := p.Position

	p.Scan() // '.'
	m := p.ExpectIdent()

	return MemberSelectExpr{
		PosRange: PosRange{From: begin, To: p.Position},
		Member:   m,
	}
}

func (p *Parser) ExpectCallExpr(callee Expr) CallExpr {
	// CallExpr -> Expr + '(' + ExprList + ')'

	begin := p.Position

	p.Scan() // '('

	params := p.ExpectExprList(token.RPAREN)

	p.MatchTerms(token.RPAREN)

	return CallExpr{
		PosRange: PosRange{From: begin, To: p.Position},
		Callee:   callee,
		Params:   params,
	}
}

// ExpectExprSuccessor continues to parse if the next token can lead to a new expression.
func (p *Parser) ExpectExprSuccessor(expr Expr) Expr {
	switch p.Token.Kind {
	case token.MEMBER_SELECT:
		return p.ExpectMemberSelectExpr(expr)
	case token.LBRACK:
		return p.ExpectIndexExpr(expr)
	case token.LPAREN:
		return p.ExpectCallExpr(expr)
	default:
		if token.PostfixUnaryOperators[p.Token.Kind] {
			op := p.ExpectOperator()

			return UnaryExpr{
				PosRange: PosRange{From: expr.Pos(), To: op.End()},
				Operator: op,
				Expr:     expr,
			}
		}

		panic(UnexpectedNodeError{
			Node:   p.Token,
			Expect: []int{token.MEMBER_SELECT, token.LBRACK, token.LPAREN}})
	}
}

func (p *Parser) ExpectExpr() (expr Expr) {
	// Expr -> '(' + Expr + ')'
	//       | BinaryExpr
	//       | UnaryExpr
	//       | CallExpr
	//       | Ident
}

func (p *Parser) ExpectExprList(terminator int) (exprs []Expr) {
	// ExprList -> Expr
	//           | ExprList + ',' + Expr
	//           | ExprList + ',' + terminator

	for {
		switch p.Token.Kind {
		case token.COMMA:
			p.Scan()
			if p.Token.Kind == terminator {
				return
			}
		case terminator:
			return
		default:
			exprs = append(exprs, p.ExpectExpr())
		}
	}
}

func (p *Parser) ExpectGenDecl() GenDecl {
	// GenDecl -> Ident + Type

	begin := p.Position

	idents := p.ExpectIdentList(0)
	typ := p.ExpectType()

	return GenDecl{
		PosRange: PosRange{From: begin, To: p.Position},
		Idents:   idents,
		Type:     typ,
	}
}

func (p *Parser) ExpectGenDeclList(terminator int) (genDecls []GenDecl) {
	// GenDeclList -> GenDecl
	//              | GenDeclList + ',' + GenDecl
	//              | GenDeclList + ',' + terminator

	for {
		genDecls = append(genDecls, p.ExpectGenDecl())
		switch p.Token.Kind {
		case token.COMMA:
			p.Scan()
			if p.Token.Kind == terminator {
				return
			}
		default:
			return
		}
	}
}

func (p *Parser) ExpectStmtBlockExpr() StmtBlockExpr {
	// StmtBlockExpr -> '{' + []Stmt + '}'

	var (
		begin = p.Position
		stmts []Stmt
	)

	p.MatchTerms(token.LBRACE)
	for p.Token.Kind != token.RBRACE {
		stmts = append(stmts, p.ExpectStmt())
	}
	return StmtBlockExpr{
		PosRange: PosRange{From: begin, To: p.Position},
		Stmts:    stmts,
	}
}

func (p *Parser) ExpectReturnStmt() ReturnStmt {
	// ReturnStmt -> 'return' + ExprList

	begin := p.Position

	p.Scan()
	return ReturnStmt{
		PosRange: PosRange{From: begin, To: p.Position},
		Exprs:    p.ExpectExprList(0),
	}
}

func (p *Parser) ExpectStmt() Stmt {
	// Stmt -> ReturnStmt
	//       | GenDecl
	//       | CallExpr
	//       | StmtBlockExpr

	switch p.Token.Kind {
	case token.LBRACE: // TODO
		return p.ExpectStmtBlockExpr()
	case token.IDENT:
	case token.VAR:
	case token.RETURN:
		p.ExpectReturnStmt()
	default:
	}
	return nil // TODO
}
