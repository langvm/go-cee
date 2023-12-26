package diagnosis

import (
	"cee/ast"
	"cee/scanner"
	"term"
)

func Print(bs *scanner.BufferScanner, node ast.Node) {
	var (
		from, to   = node.Pos(), node.End()
		begin, end = from.Column, to.Column
		line       = bs.Lines[from.Line]
	)

	println(line[0:begin], term.Red, line[begin:end], term.Reset, line[end:])
}
