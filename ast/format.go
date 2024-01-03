// Copyright 2023-2024 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package ast

import . "cee/internal"

func (t *Token) Print() string {
	return t.Literal
}

func (t *FuncType) Print() string {
	var b StringBuffer
	b.Print("(")
	for _, param := range t.Params {
		for _, ident := range param.Idents {
			b.Println(ident.Literal)
		}
	}
	b.Print(")")

	return b.String
}

func (t *FuncDecl) Print() string {
	var b StringBuffer
	b.Print("fun")

	return b.String
}
