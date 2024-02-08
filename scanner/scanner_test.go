// Copyright 2023-2024 LangVM Project
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
if i != 127 | 0b01 && i == '1' {
	println("String for testing.\nChinese letter: \u554a")
}`)

func TestScanner_ScanToken(t *testing.T) {
	var s = Scanner{
		BufferScanner: BufferScanner{
			Position: Position{},
			Buffer:   src,
		},
		Delimiters:  token.Delimiters,
		Whitespaces: token.Whitespaces,
	}

	func() {
		for {
			_, _, _, lit, err := s.ScanToken()
			switch v := err.(type) {
			case nil:
			case EOFError:
				println("EOF")
				return
			case error:
				println(v.Error())
				return
			default:
				panic(v)
			}
			println(string(lit))
		}
	}()
}
