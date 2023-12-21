// Copyright 2023-2023 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package scanner

import (
	"cee/token"
	"fmt"
)

type EOFError struct {
	Pos token.Position
}

func (e EOFError) Error() string {
	return "EOF"
}

type UnknownOperatorError struct {
	Pos token.Position
}

func (e UnknownOperatorError) Error() string {
	return fmt.Sprintln(e.Pos.String(), "unknown operator")
}

type UnknownEscapeCharError struct {
	Pos token.Position

	Char rune
}

func (e UnknownEscapeCharError) Error() string {
	return fmt.Sprintln(e.Pos.String(), "unknown escape char:", e.Char)
}

type NonClosedQuoteError struct {
	Pos token.Position
}

func (e NonClosedQuoteError) Error() string {
	return fmt.Sprintln(e.Pos.String(), "the string is not closed")
}

type FormatError struct {
	Pos token.Position
}

func (e FormatError) Error() string {
	return fmt.Sprintln(e.Pos.String(), "format error")
}
