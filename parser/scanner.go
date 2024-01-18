// Copyright 2023-2024 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package parser

import (
	"cee/ast"
	. "cee/internal"
	"cee/scanner"
	"cee/token"
)

type Scanner struct {
	scanner.Scanner

	Token ast.Token

	DelimiterStack []Stack[int]
}

func (s *Scanner) Setup() {
	s.Scanner.Delimiters = token.Delimiters
	s.Scanner.Whitespaces = token.Whitespaces
}

func (s *Scanner) ScanToken() (ast.Token, error) {
	begin, kind, lit, err := s.Scanner.ScanToken()
	if err != nil {
		return ast.Token{}, err
	}

	switch kind {
	case scanner.WORD:
		if k, ok := token.KeywordEnums[lit]; ok {
			kind = k
		} else {
			kind = token.IDENT
		}

		return ast.Token{
			PosRange: ast.PosRange{From: begin, To: s.Position},
			Kind:     kind,
			Literal:  lit,
		}, nil

	case scanner.MARKS:
		if k, ok := token.KeywordEnums[lit]; ok {
			kind = k
		} else {
			panic(UnknownOperatorError{
				Token: ast.Token{
					PosRange: ast.PosRange{},
					Kind:     kind,
					Literal:  lit,
				},
			})
		}

		// Semicolon auto complete.
		if kind == token.NEWLINE {
			switch s.Token.Kind {
			case token.IDENT:
			case token.RBRACE:
			case token.RBRACK:
			case token.RPAREN:
			case token.BREAK:
			case token.CONTINUE:
			case token.RETURN:
			case token.INC:
			case token.DEC:
			default:
				return s.ScanToken()
			}

			return ast.Token{
				PosRange: ast.PosRange{From: begin, To: s.Position},
				Kind:     token.SEMICOLON,
				Literal:  ";",
			}, nil
		}

		return ast.Token{
			PosRange: ast.PosRange{From: begin, To: s.Position},
			Kind:     kind,
			Literal:  lit,
		}, nil

	case scanner.COMMENT:
		return s.ScanToken()

	default:
		kind = [...]int{
			scanner.INT:    token.INT,
			scanner.CHAR:   token.CHAR,
			scanner.STRING: token.STRING,
		}[kind]

		if kind == 0 {
			panic("impossible")
		}

		return ast.Token{
			PosRange: ast.PosRange{From: begin, To: s.Position},
			Kind:     kind,
			Literal:  lit,
		}, nil
	}
}

func (s *Scanner) Scan() {
	tok, err := s.ScanToken()
	if err != nil {
		panic(err)
	}

	s.Token = tok
}
