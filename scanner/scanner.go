// Copyright 2023-2024 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package scanner

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"
)

const (
	_ = iota

	WORD
	INT
	CHAR
	STRING
	MARKS

	COMMENT
)

func IsMark(ch rune) bool {
	return unicode.IsPunct(ch) || unicode.IsSymbol(ch)
}

type Position struct {
	Offset, Line, Column int
}

func (p Position) String() string {
	return fmt.Sprint(p.Offset, ":", p.Line, ":", p.Column)
}

type BufferScanner struct {
	Position // Cursor
	Buffer   []rune

	Lines []string
}

func (bs *BufferScanner) FetchLine() string {
	begin := bs.Offset
	end := bs.Offset
	for end < len(bs.Buffer) && bs.Buffer[end] != '\n' {
		end++
	}
	return string(bs.Buffer[begin:end])
}

// PrintCursor print current line and the cursor position for debug use.
func (bs *BufferScanner) PrintCursor() {
	println(string(bs.Buffer[bs.Offset-bs.Column : bs.Offset]))
	for i := 0; i < bs.Column; i++ {
		print(" ")
	}
	println("^")
}

// Move returns current char and move cursor to the next.
// Move does not error when GetChar does not error.
func (bs *BufferScanner) Move() rune {
	ch := bs.GetChar()

	if ch == '\n' {
		bs.Column = 0
		bs.Line++
		bs.Lines = append(bs.Lines, bs.FetchLine())
	} else {
		bs.Column++
	}

	bs.Offset++

	return ch
}

// GetChar returns the char at the cursor.
func (bs *BufferScanner) GetChar() rune {
	if bs.Offset == len(bs.Buffer) {
		panic(EOFError{Pos: bs.Position})
	}
	return bs.Buffer[bs.Offset]
}

// Scanner is the token scanner.
type Scanner struct {
	BufferScanner

	Whitespaces map[rune]int
	Delimiters  map[rune]int
}

func (s *Scanner) GotoNextLine() {
	for {
		ch := s.GetChar()
		if ch == '\n' {
			return
		}
		s.Move()
	}
}

func (s *Scanner) SkipWhitespace() {
	for {
		ch := s.GetChar()
		if s.Whitespaces[ch] == 0 {
			return
		}
		s.Move()
	}
}

func (s *Scanner) ScanUnicodeCharHex(runesN int) rune {
	literal := make([]rune, runesN)
	for i := 0; i < runesN; i++ {
		ch := s.Move()
		literal[i] = ch
	}
	ch, err := strconv.ParseUint(string(literal), 16, runesN*4)
	if err != nil {
		switch {
		case errors.Is(err.(*strconv.NumError).Err, strconv.ErrRange):
			panic(FormatError{Pos: s.Position})
		case errors.Is(err.(*strconv.NumError).Err, strconv.ErrSyntax):
			panic(FormatError{Pos: s.Position})
		default:
			panic(err)
		}
	}
	return rune(ch)
}

// ScanEscapeChar returns the parsed char.
func (s *Scanner) ScanEscapeChar(quote rune) rune {
	ch := s.Move()
	switch ch {
	case quote:
		return quote
	case 'n':
		return '\n'
	case 't':
		return '\t'
	case 'r':
		return '\r'
	case '\\':
		return '\\'
	case 'x': // Hex 1-byte unicode, 2 runes in total.
		return s.ScanUnicodeCharHex(2)
	case 'u': // Hex 2-byte unicode, 4 runes in total.
		return s.ScanUnicodeCharHex(4)
	case 'U': // Hex 4-byte unicode, 8 runes in total.
		return s.ScanUnicodeCharHex(8)
	default:
		panic(UnknownEscapeCharError{Char: ch})
	}
}

// ScanQuotedString scans the string or char.
// PANIC: Non-closed string might cause panic due to EOFError.
func (s *Scanner) ScanQuotedString(quote rune) (int, []rune) {
	s.Move()
	var str []rune
	for {
		ch := s.Move()
		switch ch {
		case '\\':
			ch := s.ScanEscapeChar(quote)
			str = append(str, ch)
		case quote:
			return STRING, str
		default:
			str = append(str, ch)
		}
	}
}

// ScanQuotedChar scans char.
func (s *Scanner) ScanQuotedChar() (int, []rune) {
	_, quote := s.ScanQuotedString('\'')
	if len(quote) != 1 {
		panic(FormatError{Pos: s.Position})
	}
	return CHAR, quote
}

// ScanLineComment scans line comment.
func (s *Scanner) ScanLineComment() (int, []rune) {
	begin := s.Offset
	s.GotoNextLine()
	return COMMENT, s.Buffer[begin:s.Offset]
}

// ScanQuotedComment scans until "*/".
// Escape char does NOT affect.
func (s *Scanner) ScanQuotedComment() (int, []rune) {
	begin := s.Offset
	for {
		end := s.Offset
		ch := s.Move()
		if ch == '*' {
			ch := s.Move()
			if ch == '/' {
				return COMMENT, s.Buffer[begin:end]
			}
		}
	}
}

// ScanComment scans and distinguish line comment or quoted comment.
func (s *Scanner) ScanComment() (int, []rune) {
	ch := s.Move()
	switch ch {
	case '/':
		return s.ScanLineComment()
	case '*':
		return s.ScanQuotedComment()
	default:
		panic(FormatError{Pos: s.Position})
	}
}

func (s *Scanner) ScanDigit() (int, []rune) {
	ch := s.Move()

	digits := []rune{ch}

	// TODO

	return INT, digits
}

// ScanWord scans and accepts only letters, digits and underlines.
// No valid string found when returns empty []rune.
func (s *Scanner) ScanWord() (int, []rune) {
	var word []rune
	for {
		ch := s.GetChar()
		switch {
		case unicode.IsDigit(ch):
		case unicode.IsLetter(ch):
		case ch == '_':
		default: // Terminate
			if len(word) == 0 {
				panic(FormatError{Pos: s.Position})
			}
			return WORD, word
		}

		s.Move()
		word = append(word, ch)
	}
}

// ScanMarkSeq scans CONSEQUENT marks except/until delimiters.
func (s *Scanner) ScanMarkSeq() (int, []rune) {
	var seq []rune
	for {
		ch := s.GetChar()
		if IsMark(ch) && s.Delimiters[ch] == 0 {
			seq = append(seq, ch)
			s.Move()
		} else {
			if len(seq) == 0 {
				panic(FormatError{Pos: s.Position})
			}
			return MARKS, seq
		}
	}
}

// ScanToken decides the next way to scan by the cursor.
func (s *Scanner) ScanToken() (Position, int, string) {
	s.SkipWhitespace()

	begin := s.Position

	kind, lit := func() (int, []rune) {
		ch := s.GetChar()

		switch {
		case unicode.IsDigit(ch): // Digital literal value
			return s.ScanDigit()

		case unicode.IsLetter(ch) || ch == '_': // Keyword OR Idents
			return s.ScanWord()

		case ch == '"': // String
			return s.ScanQuotedString(ch)

		case ch == '\'': // Char
			return s.ScanQuotedChar()

		case s.Delimiters[ch] != 0: // Parentheses, brackets, braces, comma ...
			s.Move()
			return MARKS, []rune{ch}

		case ch == '/': // Comment
			return s.ScanComment()

		case IsMark(ch): // Operator
			return s.ScanMarkSeq()

		default:
			panic("impossible")
		}
	}()

	return begin, kind, string(lit)
}
