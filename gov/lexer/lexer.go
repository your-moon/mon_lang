package lexer

import (
	"fmt"
)

type Scanner struct {
	Source []int32

	// asdf
	//^
	Cursor uint

	// asdf
	// ^
	Start uint

	// asdf
	//    ^
	Current int32

	Line   uint
	Column uint
}

func NewScanner(source []int32) Scanner {
	return Scanner{
		Line:    1,
		Cursor:  0,
		Start:   0,
		Current: source[0],
		Source:  source,
	}
}

func (s *Scanner) IsWhiteSpace() bool {
	if s.Peek() == 32 {
		return true
	}
	return false
}

func (s *Scanner) IsTab() bool {
	if s.Peek() == 9 {
		return true
	}
	return false
}

func (s *Scanner) isDigit(c int32) bool {
	return c >= '0' && c <= '9'
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
func (s *Scanner) Next() int32 {
	if s.Cursor >= uint(len(s.Source)) {
		s.Current = 0
		return 0
	}
	prev := s.Source[s.Cursor]
	s.Cursor++
	s.Column++
	if s.Cursor < uint(len(s.Source)) {
		s.Current = s.Source[s.Cursor]
	} else {
		s.Current = 0
	}
	return prev
}
func (s *Scanner) Peek() int32 {
	return s.Current
}

func (s *Scanner) BuildToken(ttype TokenType) Token {
	if ttype == IDENT || ttype == NUMBER {
		str := string(s.Source[s.Start:s.Cursor])
		return BuildToken(ttype, &str, int(s.Line), int(s.Start), int(s.Cursor))
	}
	return BuildToken(ttype, nil, int(s.Line), int(s.Start), int(s.Cursor))
}

func (s *Scanner) ToKeyword() (Token, bool) {
	str := string(s.Source[s.Start:s.Cursor])
	if str == string(KeywordNot) {
		return s.BuildToken(IFNOT), true
	}
	if str == string(KeywordIs) {
		return s.BuildToken(IS), true
	}
	if str == string(KeywordIf) {
		return s.BuildToken(IF), true
	}
	if str == string(KeywordFn) {
		return s.BuildToken(FN), true
	}

	if str == string(KeywordFn) {
		return s.BuildToken(FN), true
	}
	if str == string(KeywordDecl) {
		return s.BuildToken(DECL), true
	}
	if str == string(KeywordInt) {
		return s.BuildToken(INT_TYPE), true
	}
	if str == string(KeywordVoid) {
		return s.BuildToken(VOID), true
	}
	if str == string(KeywordReturn) {
		return s.BuildToken(RETURN), true
	}

	if str == string(KeywordPrint) {
		return s.BuildToken(PRINT), true
	}

	return Token{}, false
}

func (s *Scanner) BuildNumber() (Token, error) {
	for s.isDigit(s.Peek()) {
		s.Next()

	}

	return s.BuildToken(NUMBER), nil
}
func (s *Scanner) BuildIdent() (Token, error) {
	for s.isAlpha(s.Peek()) {
		s.Next()
	}
	token, isKeyword := s.ToKeyword()
	if isKeyword {
		return token, nil
	}

	return s.BuildToken(IDENT), nil
}

func (s *Scanner) isAtEnd() bool {
	if s.Current == 0 {
		return true
	}
	return false
}

func (s *Scanner) IsLine() bool {
	if s.Peek() == 10 {
		return true
	}
	return false
}

func (s *Scanner) Skip() {
	for {
		if s.IsLine() {
			s.Next()
			s.Line++
			s.Column = 0
			continue
		}

		if s.IsWhiteSpace() {
			s.Next()
			continue
		}

		if s.IsTab() {
			s.Next()
			continue
		}

		if s.Peek() == '/' && s.Cursor+1 < uint(len(s.Source)) && s.Source[s.Cursor+1] == '/' {
			s.Next()
			s.Next()

			for !s.IsLine() && !s.isAtEnd() {
				s.Next()
			}
			continue
		}

		break
	}
}
func (s *Scanner) Scan() (Token, error) {
	s.Skip() //skip whitespace and incr line
	if s.isAtEnd() {
		return s.BuildToken(EOF), nil
	}

	s.Start = s.Cursor
	c := s.Next()

	for s.isAlpha(c) {
		return s.BuildIdent()
	}
	for s.isDigit(c) {
		return s.BuildNumber()
	}

	switch c {
	case '+':
		return s.BuildToken(PLUS), nil
	case '-':
		if s.Peek() == '>' {
			s.Next()
			return s.BuildToken(RIGHT_ARROW), nil
		}
		return s.BuildToken(MINUS), nil
	case '*':
		return s.BuildToken(MUL), nil
	case '/':
		return s.BuildToken(DIV), nil
	case '%':
		return s.BuildToken(PERCENT), nil
	case '(':
		return s.BuildToken(OPEN_PAREN), nil
	case ')':
		return s.BuildToken(CLOSE_PAREN), nil
	case '{':
		return s.BuildToken(OPEN_BRACE), nil
	case '}':
		return s.BuildToken(CLOSE_BRACE), nil
	case ':':
		return s.BuildToken(COLON), nil
	case ';':
		return s.BuildToken(SEMICOLON), nil
	case '~':
		return s.BuildToken(TILDE), nil
	case '?':
		return s.BuildToken(QUESTIONMARK), nil
	case '!':
		if s.Peek() == '=' {
			s.Next()
			return s.BuildToken(NOTEQUAL), nil
		}
		return s.BuildToken(NOT), nil
	case '&':
		if s.Peek() == '&' {
			s.Next()
			return s.BuildToken(LOGICAND), nil
		}
		return Token{}, fmt.Errorf(
			"not implemented: got [%c] and line [%d] where [%d]",
			c,
			s.Line,
			s.Column,
		)
	case '|':
		if s.Peek() == '|' {
			s.Next()
			return s.BuildToken(LOGICOR), nil
		}
		return Token{}, fmt.Errorf(
			"not implemented: got [%c] and line [%d] where [%d]",
			c,
			s.Line,
			s.Column,
		)
	case '=':
		if s.Peek() == '=' {
			s.Next()
			return s.BuildToken(EQUALTO), nil
		}
		return s.BuildToken(ASSIGN), nil
	case '>':
		if s.Peek() == '=' {
			s.Next()
			return s.BuildToken(GREATERTHANEQUAL), nil
		}
		return s.BuildToken(GREATERTHAN), nil
	case '<':
		if s.Peek() == '=' {
			s.Next()
			return s.BuildToken(LESSTHANEQUAL), nil
		}
		return s.BuildToken(LESSTHAN), nil
	}

	return Token{}, fmt.Errorf(
		"not implemented: got [%c] and line [%d] where [%d]",
		c,
		s.Line,
		s.Column,
	)
}
