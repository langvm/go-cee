// Copyright 2024 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package diagnosis

import (
	"cee/ast"
	. "cee/locale"
	"fmt"
)

type SyntaxError struct {
	Kind  int
	Error any
}

const (
	_ = iota

	UnexpectedNode
)

type UnexpectedNodeError struct {
	Have ast.Node
	Want ast.Kind
}

func (e UnexpectedNodeError) Error() string {
	from := e.Have.GetPosRange().From

	if tok, ok := e.Have.(ast.Token); ok {
		return fmt.Sprint(from.String(), Tr(" syntax error: unexpected token: "), tok.Literal)
	}
	return fmt.Sprint(from.String(), Tr(" syntax error: unexpected node"))
}
