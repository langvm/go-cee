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

func IsMark(ch rune) bool {
	return unicode.IsPunct(ch) || unicode.IsSymbol(ch)
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
func (bs *BufferScanner) Move() (rune, error) {
	ch, err := bs.GetChar()
	if err != nil {
		return 0, err
	}

	if ch == '\n' {
		bs.Column = 0
		bs.Line++
		bs.Lines = append(bs.Lines, bs.FetchLine())
	} else {
		bs.Column++
	}

	bs.Offset++

	return ch, nil
}

// GetChar returns the char at the cursor.
func (bs *BufferScanner) GetChar() (rune, error) {
	if bs.Offset == len(bs.Buffer) {
		return 0, EOFError{Pos: bs.Position}
	}
	return bs.Buffer[bs.Offset], nil
}

// Scanner is the token scanner.
type Scanner struct {
	BufferScanner

	Delimiters map[rune]int
}

func (s *Scanner) GotoNextLine() error {
	for {
		ch, err := s.GetChar()
		if err != nil {
			return err
		}
		if ch == '\n' {
			return nil
		}
		_, _ = s.Move()
	}
}

func (s *Scanner) SkipWhitespace() error {
	for {
		ch, err := s.GetChar()
		if err != nil {
			return err
		}
		switch ch {
		case ' ':
		case '\t':
		case '\r':
		default:
			return nil
		}
		_, _ = s.Move()
	}
}

func (s *Scanner) ScanUnicodeCharHex(runesN int) (rune, error) {
	literal := make([]rune, runesN)
	for i := 0; i < runesN; i++ {
		ch, err := s.Move()
		if err != nil {
			return 0, err
		}
		literal[i] = ch
	}
	ch, err := strconv.ParseUint(string(literal), 16, runesN*4)
	if err != nil {
		switch {
		case errors.Is(err.(*strconv.NumError).Err, strconv.ErrRange):
			return 0, FormatError{Pos: s.Position}
		case errors.Is(err.(*strconv.NumError).Err, strconv.ErrSyntax):
			return 0, FormatError{Pos: s.Position}
		default:
			return 0, err
		}
	}
	return rune(ch), nil
}

// ScanEscapeChar returns the parsed char.
func (s *Scanner) ScanEscapeChar(quote rune) (rune, error) {
	ch, _ := s.Move()
	switch ch {
	case quote:
		return quote, nil
	case 'n':
		return '\n', nil
	case 't':
		return '\t', nil
	case 'r':
		return '\r', nil
	case '\\':
		return '\\', nil
	case 'x': // Hex 1-byte unicode, 2 runes in total.
		return s.ScanUnicodeCharHex(2)
	case 'u': // Hex 2-byte unicode, 4 runes in total.
		return s.ScanUnicodeCharHex(4)
	case 'U': // Hex 4-byte unicode, 8 runes in total.
		return s.ScanUnicodeCharHex(8)
	default:
		return 0, UnknownEscapeCharError{Char: ch}
	}
}

// ScanQuotedString scans the string or char.
// PANIC: Non-closed string might cause panic due to EOFError.
func (s *Scanner) ScanQuotedString(quote rune) (int, []rune, error) {
	_, _ = s.Move()
	var str []rune
	for {
		ch, err := s.Move()
		if err != nil {
			return 0, nil, err
		}
		switch ch {
		case '\\':
			ch, err := s.ScanEscapeChar(quote)
			if err != nil {
				return 0, nil, err
			}
			str = append(str, ch)
		case quote:
			return STRING, str, nil
		default:
			str = append(str, ch)
		}
	}
}

// ScanQuotedChar scans char.
func (s *Scanner) ScanQuotedChar() (int, []rune, error) {
	_, quote, err := s.ScanQuotedString('\'')
	if err != nil {
		return 0, nil, err
	}
	if len(quote) != 1 {
		return 0, nil, FormatError{Pos: s.Position}
	}
	return CHAR, quote, nil
}

// ScanLineComment scans line comment.
func (s *Scanner) ScanLineComment() (int, []rune, error) {
	begin := s.Offset
	err := s.GotoNextLine()
	if err != nil {
		return 0, nil, err
	}
	return COMMENT, s.Buffer[begin:s.Offset], nil
}

// ScanQuotedComment scans until "*/".
// Escape char does NOT affect.
func (s *Scanner) ScanQuotedComment() (int, []rune, error) {
	begin := s.Offset
	for {
		end := s.Offset
		ch, err := s.Move()
		if err != nil {
			return 0, nil, err
		}
		if ch == '*' {
			ch, err := s.Move()
			if err != nil {
				return 0, nil, err
			}
			if ch == '/' {
				return COMMENT, s.Buffer[begin:end], nil
			}
		}
	}
}

// ScanComment scans and distinguish line comment or quoted comment.
func (s *Scanner) ScanComment() (int, []rune, error) {
	ch, err := s.Move()
	if err != nil {
		return 0, nil, err
	}
	switch ch {
	case '/':
		return s.ScanLineComment()
	case '*':
		return s.ScanQuotedComment()
	default:
		return 0, nil, FormatError{Pos: s.Position}
	}
}

func (s *Scanner) ScanDigit() (int, []rune, error) {
	ch, _ := s.Move()

	digits := []rune{ch}

	// TODO

	return INT, digits, nil
}

// ScanWord scans and accepts only letters, digits and underlines.
// No valid string found when returns empty []rune.
func (s *Scanner) ScanWord() (int, []rune, error) {
	var word []rune
	for {
		ch, err := s.GetChar()
		if err != nil {
			return 0, nil, err
		}
		switch {
		case unicode.IsDigit(ch):
		case unicode.IsLetter(ch):
		case ch == '_':
		default: // Terminate
			if len(word) == 0 {
				return 0, nil, FormatError{Pos: s.Position}
			}
			return WORD, word, nil
		}

		_, _ = s.Move()
		word = append(word, ch)
	}
}

// ScanMarkSeq scans CONSEQUENT marks except/until delimiters.
func (s *Scanner) ScanMarkSeq() (int, []rune, error) {
	var seq []rune
	for {
		ch, err := s.GetChar()
		if err != nil {
			return 0, nil, err
		}
		if IsMark(ch) && s.Delimiters[ch] == 0 {
			seq = append(seq, ch)
			_, _ = s.Move()
		} else {
			if len(seq) == 0 {
				return 0, nil, FormatError{Pos: s.Position}
			}
			return MARKS, seq, nil
		}
	}
}

// ScanToken decides the next way to scan by the cursor.
func (s *Scanner) ScanToken() (Position, int, string, error) {
	err := s.SkipWhitespace()
	if err != nil {
		return Position{}, 0, "", err
	}

	begin := s.Position

	kind, lit, err := func() (int, []rune, error) {
		ch, err := s.GetChar()
		if err != nil {
			return 0, nil, err
		}

		switch {
		case unicode.IsDigit(ch): // Digital literal value
			return s.ScanDigit()

		case unicode.IsLetter(ch) || ch == '_': // Keyword OR Idents
			return s.ScanWord()

		case ch == '"': // String
			return s.ScanQuotedString(ch)

		case ch == '\'': // Char
			return s.ScanQuotedChar()

		case s.Delimiters[ch] != 0:
			_, _ = s.Move()
			return MARKS, []rune{ch}, nil

		case ch == '/': // Comment
			return s.ScanComment()

		case IsMark(ch): // Operator
			return s.ScanMarkSeq()

		default:
			panic("impossible")
		}
	}()

	return begin, kind, string(lit), err
}
