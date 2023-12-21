// Copyright 2023-2023 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package internal

import "strconv"

func Utoa[T uint | uint8 | uint16 | uint32 | uint64](i T, base int) string {
	return strconv.FormatUint(uint64(i), base)
}

func Itoa[T int | int8 | int16 | int32 | int64](i T, base int) string {
	return strconv.FormatInt(int64(i), base)
}
