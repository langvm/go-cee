// Copyright 2023-2024 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package parser

import (
	"cee/ast"
	"cee/diagnosis"
	"cee/scanner"
	"cee/token"
	"runtime/debug"
	"testing"
)

func newParser(src string) Parser {
	p := Parser{
		Scanner: Scanner{
			Scanner: scanner.Scanner{
				Delimiters: token.Delimiters,
				BufferScanner: scanner.BufferScanner{
					Buffer: []rune(src)}}}}
	p.Setup()
	return p
}

func catch() {
	switch v := recover().(type) {
	case nil:
	case UnexpectedNodeError:
		println(v.Error())
	}
}

func assert(t *testing.T, msg string, cond bool) {
	if !cond {
		t.Error(msg)
		debug.PrintStack()
	}
}

func TestParser_ExpectStructType(t *testing.T) {
	p := newParser(`
struct {
	fieldA, fieldB TypeAlias
	fieldC TypeAlias
	Combination
}
`)
	p.Scan()
	typ := p.ExpectStructType()
	assert(t, "field gen decls number incorrect", len(typ.Fields) == 3)
}

func TestParser_ExpectGenDecl(t *testing.T) {
	p := newParser(`
ident, aa struct {
	Combination
	fieldA struct {
		fieldAA, fieldAB int
	}
	fieldB int
}
`)
	func() {
		defer func() {
			switch v := recover().(type) {
			case UnexpectedNodeError:
				println(v.Error())
				diagnosis.Print(&p.BufferScanner, v.Node)
			case nil:
				return
			default:
				panic(v)
			}
		}()

		p.Scan()
		genDecl := p.ExpectGenDecl()

		assert(t, "idents are incorrect", len(genDecl.Idents) == 2)
		assert(t, "ident name incorrect", genDecl.Idents[0].Literal == "ident")
		assert(t, "type name incorrect", len(genDecl.Type.(ast.StructType).Fields) == 3)
		assert(t, "nested fields are incorrect", len(genDecl.Type.(ast.StructType).Fields[1].Type.(ast.StructType).Fields) == 1)
	}()
}

func TestParser_ExpectFuncType(t *testing.T) {
	p := newParser(`
(paramA, paramB int, paramC int) (int, int, struct {})
`)
	p.Scan()
	typ := p.ExpectFuncType()
	assert(t, "params are incorrect", len(typ.Params) == 2)
	assert(t, "results are incorrect", len(typ.Results) == 3)
}

func TestParser_ExpectFuncDecl(t *testing.T) {
	p := newParser(`
fun Idents(paramA, paramB int, paramC string) (int, int, string) {
	return 0, 0, paramC
}
`)
	defer catch()

	p.Scan()
	funcDecl := p.ExpectFuncDecl()
	typ := funcDecl.Type
	assert(t, "function name incorrect", funcDecl.Ident.Literal == "Idents")
	assert(t, "paramB incorrect", typ.Params[0].Idents[1].Literal == "paramB")
	assert(t, "3rd result incorrect", typ.Results[2].(ast.TypeAlias).Literal == "string")
}

func TestParser_ExpectLeftAssociativeExpr(t *testing.T) {
	p := newParser(`
base.A.B + 1
`)
	p.Scan()
	expr := p.ExpectLeftAssociativeExpr()
	assert(t, "member incorrect", expr.(ast.MemberSelectExpr).Member.Literal == "B")
}

func TestParser_ExpectExpr(t *testing.T) {
	p := newParser(`
identA * identC + identB * identC * (identA + identB)
`)
	p.Scan()
	expr := p.ExpectExpr()
	println(expr.(ast.BinaryExpr).Operator.Literal)
}

func Test_ExpectBinaryExpr(t *testing.T) {
	p := newParser(`
a + a * b * c
`)
	p.Scan()
	expr := p.ExpectExpr()

	println(expr.GetPosRange().To.String())
}
