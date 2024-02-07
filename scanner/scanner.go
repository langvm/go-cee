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

	IDENT
	MARK

	INT
	FLOAT

	CHAR
	STRING

	COMMENT
)

const (
	_ = iota

	INT_HEX
	INT_DEC
	INT_OCT
	INT_BIN
)

const (
	_ = iota

	COMMENT_LINE
	COMMENT_QUOTED
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

// Move moves the cursor and return the next char.
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

	Whitespaces map[rune]int
	Delimiters  map[rune]int
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
		_, err = s.Move()
		if err != nil {
			return err
		}
	}
}

func (s *Scanner) SkipWhitespace() error {
	for {
		ch, err := s.GetChar()
		if err != nil {
			return err
		}
		if s.Whitespaces[ch] == 0 {
			return nil
		}
		_, err = s.Move()
		if err != nil {
			return err
		}
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
	ch, err := s.Move()
	if err != nil {
		return 0, err
	}
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
func (s *Scanner) ScanQuotedString(quote rune) ([]rune, error) {
	_, err := s.Move()
	if err != nil {
		return nil, err
	}

	var str []rune
	for {
		ch, err := s.Move()
		if err != nil {
			return nil, err
		}
		switch ch {
		case '\\':
			ch, err := s.ScanEscapeChar(quote)
			if err != nil {
				return nil, err
			}
			str = append(str, ch)
		case quote:
			return str, nil
		default:
			str = append(str, ch)
		}
	}
}

// ScanQuotedChar scans char.
func (s *Scanner) ScanQuotedChar() ([]rune, error) {
	quote, err := s.ScanQuotedString('\'')
	if err != nil {
		return nil, err
	}
	if len(quote) != 1 {
		return nil, FormatError{Pos: s.Position}
	}
	return quote, nil
}

// ScanLineComment scans line comment.
func (s *Scanner) ScanLineComment() ([]rune, error) {
	begin := s.Offset
	err := s.GotoNextLine()
	if err != nil {
		return nil, err
	}
	return s.Buffer[begin:s.Offset], nil
}

// ScanQuotedComment scans until "*/".
// Escape char does NOT affect.
func (s *Scanner) ScanQuotedComment() ([]rune, error) {
	begin := s.Offset
	for {
		end := s.Offset
		ch, err := s.Move()
		if err != nil {
			return nil, err
		}
		if ch == '*' {
			ch, err := s.Move()
			if err != nil {
				return nil, err
			}
			if ch == '/' {
				return s.Buffer[begin:end], nil
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

		lit, err := s.ScanLineComment()
		return COMMENT_LINE, lit, err
	case '*':
		lit, err := s.ScanQuotedComment()
		return COMMENT_QUOTED, lit, err
	default:
		return 0, nil, FormatError{Pos: s.Position}
	}
}

func (s *Scanner) ScanWhile(cond func() bool) ([]rune, error) {
	var seq []rune

	for cond() {
		ch, err := s.GetChar()
		if err != nil {
			return nil, err
		}

		seq = append(seq, ch)
		_, err = s.Move()
		if err != nil {
			return nil, err
		}
	}

	if len(seq) == 0 {
		return nil, FormatError{Pos: s.Position}
	}

	return seq, nil
}

func (s *Scanner) ScanBinDigit() ([]rune, error) {
	cond := func() bool {
		ch, err := s.GetChar()
		if err != nil {
			return false
		}
		return ch == '0' || ch == '1'
	}
	return s.ScanWhile(cond)
}

func (s *Scanner) ScanOctDigit() ([]rune, error) {
	cond := func() bool {
		ch, err := s.GetChar()
		if err != nil {
			return false
		}
		return '0' <= ch && ch <= '7'
	}
	return s.ScanWhile(cond)
}

func (s *Scanner) ScanHexDigit() ([]rune, error) {
	cond := func() bool {
		ch, err := s.GetChar()
		if err != nil {
			return false
		}
		return '0' <= ch && ch <= '9' || 'a' <= ch && ch <= 'f'
	}
	return s.ScanWhile(cond)
}

func (s *Scanner) ScanDigit() (int, []rune, error) {
	ch, err := s.Move()
	if err != nil {
		return 0, nil, err
	}
	digits := []rune{ch}

	if ch == '0' {
		ch, err := s.Move()
		if err != nil {
			return 0, nil, err
		}
		switch ch {
		case 'x':
			lit, err := s.ScanHexDigit()
			return INT_HEX, lit, err
		case 'o':
			lit, err := s.ScanOctDigit()
			return INT_OCT, lit, err
		case 'b':
			lit, err := s.ScanBinDigit()
			return INT_BIN, lit, err
		default:
			return 0, nil, FormatError{Pos: s.Position}
		}
	}

	for unicode.IsDigit(ch) {
		ch, err := s.Move()
		if err != nil {
			return 0, nil, err
		}
		digits = append(digits, ch)
	}

	return 0, digits, nil
}

// ScanIdent scans and accepts only letters, digits and underlines.
// No valid string found when returns empty []rune.
func (s *Scanner) ScanIdent() ([]rune, error) {
	var word []rune
	for {
		ch, err := s.GetChar()
		if err != nil {
			return nil, err
		}
		switch {
		case unicode.IsDigit(ch):
		case unicode.IsLetter(ch):
		case ch == '_':
		default: // Terminate
			if len(word) == 0 {
				return nil, FormatError{Pos: s.Position}
			}
			return word, nil
		}

		_, err = s.Move()
		if err != nil {
			return nil, err
		}

		word = append(word, ch)
	}
}

// ScanMarkSeq scans CONSEQUENT marks except/until delimiters.
func (s *Scanner) ScanMarkSeq() ([]rune, error) {
	var seq []rune
	for {
		ch, err := s.GetChar()
		if err != nil {
			return nil, err
		}
		if IsMark(ch) && s.Delimiters[ch] == 0 {
			seq = append(seq, ch)
			_, err := s.Move()
			if err != nil {
				return nil, err
			}
		} else {
			if len(seq) == 0 {
				panic(FormatError{Pos: s.Position})
			}
			return seq, nil
		}
	}
}

// ScanToken decides the next way to scan by the cursor.
func (s *Scanner) ScanToken() (Position, int, int, []rune, error) {
	err := s.SkipWhitespace()
	if err != nil {
		return Position{}, 0, 0, nil, err
	}

	begin := s.Position

	ch, err := s.GetChar()
	if err != nil {
		return begin, 0, 0, nil, err
	}

	switch {
	case unicode.IsDigit(ch): // Digital literal value
		format, lit, err := s.ScanDigit()
		return begin, INT, format, lit, err

	case unicode.IsLetter(ch) || ch == '_': // Keyword OR Idents
		lit, err := s.ScanIdent()
		return begin, IDENT, 0, lit, err

	case ch == '"': // String
		lit, err := s.ScanQuotedString(ch)
		return begin, STRING, 0, lit, err

	case ch == '\'': // Char
		lit, err := s.ScanQuotedChar()
		return begin, CHAR, 0, lit, err

	case s.Delimiters[ch] != 0: // Parentheses, brackets, braces, comma ...
		ch, err := s.Move()
		return begin, MARK, 0, []rune{ch}, err

	case ch == '/': // Comment
		format, lit, err := s.ScanComment()
		return begin, COMMENT, format, lit, err

	case IsMark(ch): // Operator
		lit, err := s.ScanMarkSeq()
		return begin, MARK, 0, lit, err

	default:
		panic("impossible")
	}
}
