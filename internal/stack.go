// Copyright 2023-2023 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package internal

type Stack[T any] struct {
	E   []T
	Top int
}

func NewStack[T any]() Stack[T] {
	return Stack[T]{
		E:   make([]T, 0),
		Top: 0,
	}
}

func (s *Stack[T]) Push(e T) {
	if s.Top > cap(s.E) {
		s.E = append(s.E, e)
	} else {
		s.E[s.Top] = e
	}
	s.Top++
}

func (s *Stack[T]) Pop() T {
	s.Top--
	return s.E[s.Top]
}
