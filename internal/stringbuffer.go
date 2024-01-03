// Copyright 2023-2024 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package internal

import "fmt"

type StringBuffer struct {
	String string
}

func (b *StringBuffer) Print(a ...any)                 { b.String += fmt.Sprint(a...) }
func (b *StringBuffer) Println(a ...any)               { b.String += fmt.Sprintln(a...) }
func (b *StringBuffer) Printf(format string, a ...any) { b.String += fmt.Sprintf(format, a...) }
