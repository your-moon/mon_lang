package lexer

type TokenType string

const (
	PLUS  TokenType = "PLUS"
	MINUS TokenType = "MINUS"
	MUL   TokenType = "MUL"
	DIV   TokenType = "DIV"

	RETURN TokenType = "RETURN"
	PRINT  TokenType = "PRINT"

	IDENT       TokenType = "IDENT"
	NUMBER      TokenType = "NUMBER"
	FN          TokenType = "FN"
	OPEN_PAREN  TokenType = "OPEN_PAREN"
	CLOSE_PAREN TokenType = "CLOSE_BRACE"
	RIGHT_ARROW TokenType = "RIGHT_ARROW"
	OPEN_BRACE  TokenType = "OPEN_BRACE"
	CLOSE_BRACE TokenType = "CLOSE_BRACE"
	EOF         TokenType = "EOF"
)

type Token struct {
	Type  TokenType
	Value *string
}

func BuildToken(ttype TokenType, value *string) Token {
	return Token{
		Type:  ttype,
		Value: value,
	}
}
