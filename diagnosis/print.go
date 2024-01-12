// Copyright 2023-2024 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package diagnosis

import (
	"cee/ast"
	"cee/scanner"
	"term"
)

func Print(bs *scanner.BufferScanner, node ast.Node) {
	var (
		posRange   = node.GetPosRange()
		begin, end = posRange.From.Column, posRange.To.Column
		line       = bs.Lines[posRange.From.Line]
	)

	println(line[0:begin], term.Red, line[begin:end], term.Reset, line[end:])
}
