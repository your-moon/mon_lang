package lexer

type TokenType string

const (
	PLUS    TokenType = "PLUS"
	MINUS   TokenType = "MINUS"
	MUL     TokenType = "MUL"
	DIV     TokenType = "DIV"
	PERCENT TokenType = "PERCENT"

	GREATERTHAN      TokenType = "GREATERTHAN"      // >
	GREATERTHANEQUAL TokenType = "GREATERTHANEQUAL" //>=
	LESSTHAN         TokenType = "LESSTHAN"         // <
	LESSTHANEQUAL    TokenType = "LESSTHANEQUAL"    // <=
	ASSIGN           TokenType = "EQUAL"            // =
	NOTEQUAL         TokenType = "NOTEQUAL"         // !=
	EQUALTO          TokenType = "EQUALTO"          // ==

	LOGICAND TokenType = "LOGICAND" // &&
	LOGICOR  TokenType = "LOGICOR"  // ||
	NOT      TokenType = "NOT"      // !

	RETURN TokenType = "RETURN"
	PRINT  TokenType = "PRINT"

	IDENT       TokenType = "IDENT"
	NUMBER      TokenType = "NUMBER"
	FN          TokenType = "FN"
	OPEN_PAREN  TokenType = "OPEN_PAREN"
	CLOSE_PAREN TokenType = "CLOSE_PAREN"
	RIGHT_ARROW TokenType = "RIGHT_ARROW"
	OPEN_BRACE  TokenType = "OPEN_BRACE"
	CLOSE_BRACE TokenType = "CLOSE_BRACE"
	SEMICOLON   TokenType = "SEMICOLON"
	TILDE       TokenType = "TILDE"
	EOF         TokenType = "EOF"

	INT_TYPE TokenType = "INT_TYPE"
	VOID     TokenType = "VOID"
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
