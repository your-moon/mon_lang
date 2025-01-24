package lexer

import (
	"errors"
	"fmt"
)

const (
	IDENT int = iota
	FN
	OPEN_PAREN
	CLOSE_PAREN
	RIGHT_ARROW
	OPEN_BRACE
	CLOSE_BRACE
	PRINT
)

type Token struct {
	Type  int
	Value string
}

type Scanner struct {
	Source  []int32
	Current int32
	Cursor  uint
	Line    uint
}

func NewScanner(source []int32) Scanner {
	return Scanner{
		Line:   0,
		Cursor: 0,
		Source: source,
	}
}

func (s *Scanner) Next() int32 {
	if s.Cursor >= uint(len(s.Source)) {
		panic(
			fmt.Sprintf(
				"there is no char on source cursor: %d source len: %d",
				s.Cursor,
				len(s.Source),
			),
		)
	}
	c := s.Source[s.Cursor]
	s.Cursor++
	return c
}

func (s *Scanner) isAlpha(c int32) bool {
	//10 => \0
	//32 => space
	//unicode 2byte long

	if c >= 'а' && c <= 'я' || c >= 'А' && c <= 'Я' {
		return true
	}

	if c == 'ё' || c == 'ү' || c == 'е' || c == 'ө' || c == 'Ё' || c == 'Ү' || c == 'Е' ||
		c == 'Ө' {
		return true
	}

	if c >= 'a' && c <= 'z' {
		return true
	}

	return false
}
func (s *Scanner) current() int32 {
	return s.Current
}

func (s *Scanner) buildIdent() (Token, error) {
	for s.isAlpha(s.current()) {
		s.Next()
	}

	return Token{
		Type:  IDENT,
		Value: "",
	}, nil
}

func (s *Scanner) Scan() (Token, error) {
	c := s.Next()

	if s.isAlpha(c) {
		fmt.Printf("%c :IS ALPHA\n", c)
		return s.buildIdent()
	}

	switch c {
	case '(':
		fmt.Println("OPEN_PAREN")
	case ')':
		fmt.Println("CLOSE_PAREN")
	}

	return Token{}, errors.New("not implemented")
}
