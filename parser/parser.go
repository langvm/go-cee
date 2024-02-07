// Copyright 2023-2024 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package parser

import (
	. "cee/ast"
	"cee/token"
	"strings"
)

func ParsePackageName(canonicalName string) string {
	split := strings.Split(canonicalName, "/")
	return split[len(split)-1]
}

// NOTICE:
// ALL Expect* functions start from the cursor position and end at the NEXT token.

type Parser struct {
	Scanner
}

func (p *Parser) MatchTerms(terms ...int) error {
	for _, term := range terms {
		if p.Token.Kind != term {
			panic(UnexpectedNodeError{
				Node:   p.Token,
				Expect: terms})
		}
		p.Scan()
	}
}

func (p *Parser) ExpectIdent() (Ident, error) {
	// Idents -> {Idents.Kind == IDENT}

	if p.Token.Kind != token.IDENT {
		panic(UnexpectedNodeError{
			Node:   p.Token,
			Expect: []int{token.IDENT}})
	}

	ident := p.Token
	p.Scan()

	return Ident{Token: ident}
}

func (p *Parser) ExpectIdentList(terminator int) (idents []Ident, _ error) {
	// IdentList -> Idents
	//            | IdentList + ',' + Idents
	//            | IdentList + ',' + terminator

	for {
		ident, err := p.ExpectIdent()
		if err != nil {
			return nil, err
		}

		idents = append(idents, ident)
		switch p.Token.Kind {
		case token.COMMA:
			p.Scan()
			if p.Token.Kind == terminator {
				return
			}
			continue
		default:
			return
		}
	}
}

// ExpectFuncType starts from left paren.
func (p *Parser) ExpectFuncType() FuncType {
	var (
		params  []GenDecl
		results []Type
	)

	begin := p.Token.From

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
		PosRange: PosRange{From: begin, To: p.Position},
		Params:   params,
		Results:  results,
	}
}

func (p *Parser) ExpectFuncDecl() FuncDecl {
	begin := p.Token.From

	p.MatchTerms(token.FUNC)

	var pIdent *Ident

	if p.Token.Kind == token.IDENT {
		ident := p.ExpectIdent()
		pIdent = &ident
	}

	typ := p.ExpectFuncType()

	var pExpr *StmtBlockExpr

	if p.Token.Kind == token.LBRACE {
		expr := p.ExpectStmtBlockExpr()
		pExpr = &expr
	}

	return FuncDecl{
		PosRange: PosRange{From: begin, To: p.Position},
		Type:     typ,
		Ident:    pIdent,
		Stmt:     pExpr,
	}
}

func (p *Parser) ExpectStructType() StructType {
	begin := p.Token.From

	p.MatchTerms(token.STRUCT, token.LBRACE)

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
	case token.FUNC:
		p.Scan()
		return p.ExpectFuncType()
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
	op := p.Token

	if !token.IsOperator(p.Token.Kind) {
		panic(UnexpectedNodeError{
			Node:   p.Token,
			Expect: []int{token.OPERATOR_BEGIN}})
	}

	p.Scan()

	return op
}

func (p *Parser) ExpectLiteralValue() LiteralValue {
	tok := p.Token
	p.Scan()
	return LiteralValue{Token: tok}
}

func (p *Parser) ExpectIndexExpr(expr Expr) IndexExpr {
	// IndexExpr -> Expr + '[' + Expr + ']'

	begin := p.Token.From

	p.MatchTerms(token.LBRACK)

	indexExpr := p.ExpectExpr()
	p.MatchTerms(token.RBRACK)

	return IndexExpr{
		PosRange: PosRange{From: begin, To: p.Position},
		Expr:     expr,
		Index:    indexExpr,
	}
}

func (p *Parser) ExpectMemberSelectExpr(expr Expr) MemberSelectExpr {
	// MemberSelectExpr -> Expr + '.' + Idents

	begin := p.Token.From

	p.MatchTerms(token.MEMBER_SELECT)
	m := p.ExpectIdent()

	return MemberSelectExpr{
		PosRange: PosRange{From: begin, To: p.Position},
		Member:   m,
	}
}

func (p *Parser) ExpectCallExpr(callee Expr) CallExpr {
	// CallExpr -> Expr + '(' + ExprList + ')'

	begin := p.Token.From

	p.MatchTerms(token.LPAREN)

	params := p.ExpectExprList(token.RPAREN)

	p.MatchTerms(token.RPAREN)

	return CallExpr{
		PosRange: PosRange{From: begin, To: p.Position},
		Callee:   callee,
		Params:   params,
	}
}

// LookAheadLeftAssociativeOperatorAndExpr continues to parse, when current token is left-associative that leads to a new expression.
func (p *Parser) LookAheadLeftAssociativeOperatorAndExpr(expr Expr) Expr {
	newExpr := func() Expr {
		switch p.Token.Kind {
		case token.MEMBER_SELECT:
			return p.ExpectMemberSelectExpr(expr)
		case token.LBRACK:
			return p.ExpectIndexExpr(expr)
		case token.LPAREN:
			return p.ExpectCallExpr(expr)
		default:
		}

		if token.PostfixUnaryOperators[p.Token.Kind] {
			op := p.ExpectOperator()

			return UnaryExpr{
				PosRange: PosRange{From: expr.GetPosRange().From, To: op.GetPosRange().To},
				Operator: op,
				Expr:     expr,
			}
		}

		return nil
	}()

	if newExpr == nil {
		return expr
	}

	return p.LookAheadLeftAssociativeOperatorAndExpr(newExpr)
}

func (p *Parser) ExpectLeftAssociativeExpr() Expr {
	expr := func() Expr {
		if token.IsLiteralValue(p.Token.Kind) {
			return p.ExpectLiteralValue()
		}

		switch p.Token.Kind {
		case token.IDENT:
			return p.ExpectIdent()
		case token.FUNC:
			return p.ExpectFuncDecl()
		case token.LPAREN:
			p.Scan()
			expr := p.ExpectExpr()
			p.MatchTerms(token.RPAREN)
			return expr
		default:
			panic(UnexpectedNodeError{
				Node:   p.Token,
				Expect: []int{}}) // TODO
		}
	}()

	if newExpr := p.LookAheadLeftAssociativeOperatorAndExpr(expr); newExpr != nil {
		return newExpr
	}

	return expr
}

func (p *Parser) ExpectPrefixUnaryExpr() UnaryExpr {
	begin := p.Token.From

	op := p.ExpectOperator()
	expr := p.ExpectLeftAssociativeExpr()

	return UnaryExpr{
		PosRange: PosRange{From: begin, To: p.Position},
		Operator: op,
		Expr:     expr,
	}
}

func (p *Parser) ExpectShortExpr() Expr {
	if token.PrefixUnaryOperators[p.Token.Kind] {
		return p.ExpectPrefixUnaryExpr()
	}

	return p.ExpectLeftAssociativeExpr()
}

// ExpectBinaryExpr parses bi-operand expressions in left-associative approach.
func (p *Parser) ExpectBinaryExpr(exprL Expr) BinaryExpr {
	op := p.ExpectOperator()
	exprR := p.ExpectShortExpr()

	expr := BinaryExpr{
		PosRange: PosRange{From: exprL.GetPosRange().From, To: exprR.GetPosRange().To},
		Operator: op,
		Exprs:    [2]Expr{exprL, exprR},
	}

	if token.BinaryOperators[p.Token.Kind] != 0 {
		return p.ExpectBinaryExpr(expr)
	}

	return expr
}

func (p *Parser) ExpectExpr() Expr {
	expr := p.ExpectShortExpr()
	if token.BinaryOperators[p.Token.Kind] != 0 {
		return p.ExpectBinaryExpr(expr)
	}
	return expr
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

func (p *Parser) ExpectImportDecl() ImportDecl {
	begin := p.Token.From

	p.MatchTerms(token.IMPORT)

	var alias *Ident

	if p.Token.Kind == token.IDENT {
		ident := p.ExpectIdent()
		alias = &ident // Escape.
	}

	lit := p.ExpectLiteralValue()

	if lit.Kind != token.STRING {
		panic(UnexpectedNodeError{
			Node:   lit,
			Expect: []int{token.STRING}})
	}

	return ImportDecl{
		PosRange:      PosRange{From: begin, To: p.Position},
		CanonicalName: lit,
		Alias:         alias,
	}
}

func (p *Parser) ExpectGenDecl() GenDecl {
	// GenDecl -> Idents + Type

	begin := p.Token.From

	idents := p.ExpectIdentList(0)
	typ := p.ExpectType()

	return GenDecl{
		PosRange: PosRange{From: begin, To: p.Position},
		Idents:   idents,
		Type:     typ,
	}
}

func (p *Parser) ExpectValDecl() ValDecl {
	// ValDecl -> 'val' + Idents + '=' + Expr

	begin := p.Token.From

	p.MatchTerms(token.VAL)

	ident := p.ExpectIdent()

	p.MatchTerms(token.ASSIGN)

	expr := p.ExpectExpr()

	return ValDecl{
		PosRange: PosRange{From: begin, To: p.Position},
		Name:     ident,
		Value:    expr,
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

	begin := p.Token.From

	p.MatchTerms(token.RETURN)

	exprs := p.ExpectExprList(0)

	return ReturnStmt{
		PosRange: PosRange{From: begin, To: p.Position},
		Exprs:    exprs,
	}
}

func (p *Parser) ExpectLoopStmt() LoopStmt {
	p.MatchTerms(token.FOR)
	p.ExpectExpr()

	return LoopStmt{} // TODO
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
	case token.VAL:
	case token.CONTINUE:
	case token.BREAK:
	case token.RETURN:
		p.ExpectReturnStmt()
	default:
	}
	return nil // TODO
}
