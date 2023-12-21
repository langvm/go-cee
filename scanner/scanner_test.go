// Copyright 2023-2023 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package scanner

import (
	"cee/token"
	"testing"
)

var src = []rune(`
package main
var i = len("String for testing."+"")
i++
if i != 0b01 | 0b01 && i == 1 {
	println("String for testing.\nChinese letter: \u554a")
}
`)

func TestScanner_ScanToken(t *testing.T) {
	var s = Scanner{BufferScanner{
		Position: token.Position{},
		Buffer:   src,
	}}

	for {
		var _, tok, err = s.ScanToken()
		switch err := err.(type) {
		case nil:
		case EOFError:
			println("EOF")
			return
		default:
			println(err.Error())
			return
		}
		if tok[0] != '\n' {
			println(string(tok))
		}
	}
}
