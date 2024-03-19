// Copyright 2024 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package parser

import (
	"cee/ast"
	"cee/diagnosis"
	"cee/stack"
	"cee/token"
	scanner "github.com/langvm/go-cee-scanner"
	"strings"
)

func ParsePackageName(canonicalName string) string {
	split := strings.Split(canonicalName, "/")
	return split[len(split)-1]
}

// NOTICE:
// ALL Expect* functions start from the cursor position and end at the NEXT token.

type Parser struct {
	scanner.Scanner
	ReachedEOF bool

	Token ast.Token

	QuoteStack []int

	Diagnosis []diagnosis.Diagnosis
}

func NewParser(buffer []rune) Parser {
	return Parser{
		Scanner: scanner.Scanner{
			BufferScanner: scanner.BufferScanner{
				Buffer: buffer,
			},
			Whitespaces: token.Whitespaces,
			Delimiters:  token.Delimiters,
		},
	}
}

func (p *Parser) Scan() {
	begin := p.Position

	bt, err := p.Scanner.Scan()
	if err != nil {
		panic(err)
	}

	var (
		kind = 0
		lit  = string(bt.Literal)
	)

	switch bt.Kind {
	case scanner.IDENT:
		kind = token.Keyword2Enum[lit]
		if kind == 0 {
			kind = token.IDENT
		}
	case scanner.OPERATOR:
		kind = token.Keyword2Enum[lit]
		if kind == 0 {
			kind = token.IDENT
		}
	case scanner.DELIMITER:
		kind = token.Keyword2Enum[lit]
		switch kind {
		case token.LBRACE:
			p.QuoteStack = append(p.QuoteStack, token.RBRACE)
		case token.LPAREN:
			p.QuoteStack = append(p.QuoteStack, token.RPAREN)
		case token.LBRACK:
			p.QuoteStack = append(p.QuoteStack, token.RBRACK)
		case token.RBRACE:
			fallthrough
		case token.RPAREN:
			fallthrough
		case token.RBRACK:
			p.QuoteStack = stack.Pop(p.QuoteStack)
		default:
		}
	case scanner.INT:
		kind = token.INT
	case scanner.CHAR:
		kind = token.CHAR
	case scanner.STRING:
		kind = token.STRING
	case scanner.COMMENT:
		p.Scan()
		return
	default:
		// TODO
	}

	p.Token = ast.Token{
		PosRange: ast.PosRange{From: begin, To: p.Position},
		Kind:     kind,
		Literal:  lit,
	}
}

func (p *Parser) Report(d diagnosis.Diagnosis) {
	p.Diagnosis = append(p.Diagnosis, d)
}

func (p *Parser) ReportAndRecover(d diagnosis.Diagnosis) {
	p.Diagnosis = append(p.Diagnosis, d)

	if len(p.QuoteStack) != 0 {
		term := stack.Top(p.QuoteStack)
		for p.Token.Kind != term {
			p.Scan()
		}
	}
}

func (p *Parser) MatchTerm(term int) {
	if p.Token.Kind != term {
		p.Report(diagnosis.Diagnosis{
			Kind: diagnosis.UnexpectedNode,
			Error: diagnosis.UnexpectedNodeError{
				Have: p.Token,
				Want: term,
			},
		})
	}
}

func ExpectList[T any](p *Parser, expectFunc func(p *Parser) T, kind int, delimiter int, terminate int) ast.List[T] {
	begin := p.Position

	var list []T

	if p.Token.Kind == delimiter {
		p.ReportAndRecover(diagnosis.Diagnosis{
			Kind: diagnosis.UnexpectedNode,
			Error: diagnosis.UnexpectedNodeError{
				Have: p.Token,
				Want: kind,
			},
		})
	}

	for {
		switch p.Token.Kind {
		case delimiter:
			p.MatchTerm(delimiter)
		case terminate:
			p.Scan()
			return ast.List[T]{
				PosRange: ast.PosRange{From: begin, To: p.Position},
				List:     list,
			}
		default:
			list = append(list, expectFunc(p))
		}
	}
}

func (p *Parser) ExpectIdent() ast.Ident {}

func (p *Parser) ExpectBranchExpr() ast.BranchExpr {}

func (p *Parser) ExpectCallExpr() ast.CallExpr {

}

func (p *Parser) ExpectAssignStmt() ast.AssignStmt {}

func (p *Parser) ExpectStmtBlock() ast.StmtBlockExpr {}

func (p *Parser) ExpectExpr() ast.Expr {
}
