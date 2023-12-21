// Copyright 2023-2023 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package token

import "fmt"

type Position struct {
	Offset, Line, Column int
}

func (p Position) String() string {
	return fmt.Sprint(p.Offset, ":", p.Line, ":", p.Column)
}
