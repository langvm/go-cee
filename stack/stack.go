// Copyright 2024 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package stack

func Pop[T any](arr []T) []T {
	return arr[:len(arr)-2]
}

func Top[T any](arr []T) T {
	return arr[len(arr)-1]
}
