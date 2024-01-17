// Copyright 2023-2024 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package parser

import (
	"cee/ast"
	"fmt"
	. "locale"
)

type UnexpectedNodeError struct {
	Node   ast.Node
	Expect []int
}

func (e UnexpectedNodeError) Error() string {
	from := e.Node.GetPosRange().From

	if tok, ok := e.Node.(ast.Token); ok {
		return fmt.Sprint(from.String(), Tr(" syntax error: unexpected token: "), tok.Literal)
	}
	return fmt.Sprint(from.String(), Tr(" syntax error: unexpected node"))
}

type UnknownOperatorError struct {
	ast.Token
}

func (e UnknownOperatorError) Error() string {
	return fmt.Sprintln("unknown operator:", e.Token.Literal)
}
