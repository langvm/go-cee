// Copyright 2023-2023 LangVM Project
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0
// that can be found in the LICENSE file and https://mozilla.org/MPL/2.0/.

package scanner

import (
	"cee/token"
	"errors"
	"strconv"
	"unicode"
)

type BufferScanner struct {
	token.Position // Cursor
	Buffer         []rune

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
// Move does not error if GetChar does not error.
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

func (bs *BufferScanner) GetChar() (rune, error) {
	if bs.Offset == len(bs.Buffer) {
		return 0, EOFError{Pos: bs.Position}
	}
	return bs.Buffer[bs.Offset], nil
}

type Scanner struct {
	BufferScanner
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
func (s *Scanner) ScanQuotedString(quote rune) ([]rune, error) {
	_, _ = s.Move()
	str := []rune{quote}
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
			str = append(str, quote)
			return str, nil
		default:
			str = append(str, ch)
		}
	}
}

// ScanQuotedChar scans char.
func (s *Scanner) ScanQuotedChar() (rune, error) {
	quote, err := s.ScanQuotedString('\'')
	if err != nil {
		return 0, err
	}
	if len(quote) != 1 {
		return 0, FormatError{Pos: s.Position}
	}
	return quote[0], nil
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
func (s *Scanner) ScanComment() ([]rune, error) {
	ch, err := s.Move()
	if err != nil {
		return nil, err
	}
	switch ch {
	case '/':
		return s.ScanLineComment()
	case '*':
		return s.ScanQuotedComment()
	default:
		return nil, FormatError{Pos: s.Position}
	}
}

func (s *Scanner) ScanDigit() ([]rune, error) {
	ch, _ := s.Move()

	digits := []rune{ch}

	// TODO

	return digits, nil
}

// ScanWord scans and accepts only letters, digits and underlines.
// No valid string found when returns empty []rune.
func (s *Scanner) ScanWord() (str []rune, err error) {
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
			if len(str) == 0 {
				return nil, FormatError{Pos: s.Position}
			}
			return str, nil
		}

		_, _ = s.Move()
		str = append(str, ch)
	}
}

// ScanMarkSeq scans CONSEQUENT marks except/until delimiters.
func (s *Scanner) ScanMarkSeq() (str []rune, err error) {
	for {
		ch, err := s.GetChar()
		if err != nil {
			return nil, err
		}
		if IsMark(ch) && !token.IsDelimiter(ch) {
			str = append(str, ch)
			_, _ = s.Move()
		} else {
			if len(str) == 0 {
				return nil, FormatError{Pos: s.Position}
			}
			return str, nil
		}
	}
}

// ScanToken decides the next way to scan by the cursor.
func (s *Scanner) ScanToken() (int, []rune, error) {
	err := s.SkipWhitespace()
	if err != nil {
		return 0, nil, err
	}

	ch, err := s.GetChar()
	if err != nil {
		return 0, nil, err
	}

	switch {
	case unicode.IsDigit(ch): // Digital literal value
		digit, err := s.ScanDigit()
		if err != nil {
			return 0, nil, err
		}
		return token.INT, digit, nil

	case unicode.IsLetter(ch) || ch == '_': // Keyword OR Ident
		tok, err := s.ScanWord()

		switch err := err.(type) {
		case nil:
		case FormatError:
		default:
			return 0, nil, err
		}

		if keyword := token.KeywordEnums[string(tok)]; keyword != 0 {
			return keyword, tok, nil
		}

		return token.IDENT, tok, nil

	case ch == '"': // String
		tok, err := s.ScanQuotedString(ch)
		return token.STRING, tok, err

	case ch == '\'': // Char
		tok, err := s.ScanQuotedChar()
		return token.CHAR, []rune{tok}, err

	case ch == '/': // Comment
		_, err := s.ScanComment() // TODO
		if err != nil {
			return 0, nil, err
		}
		return s.ScanToken()

	case token.IsDelimiter(ch):
		_, _ = s.Move()
		return token.Delimiters[ch], []rune{ch}, nil

	case IsMark(ch): // Operator
		tok, err := s.ScanMarkSeq()

		switch err := err.(type) {
		case nil:
		case FormatError:
		default:
			return 0, nil, err
		}

		if keyword := token.KeywordEnums[string(tok)]; keyword != 0 {
			return keyword, tok, nil
		}

		return 0, nil, UnknownOperatorError{Pos: s.Position}
	default:
		panic("impossible")
	}
}
