// Copyright 2023-2023 LangVM Project
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
	return fmt.Sprint(e.Node.Pos().String(), Tr(" syntax error: unexpected "))
}
