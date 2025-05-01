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
	ASSIGN           TokenType = "ASSIGN"           // =
	NOTEQUAL         TokenType = "NOTEQUAL"         // !=
	EQUALTO          TokenType = "EQUALTO"          // ==

	COMMA  TokenType = "COMMA"  // ,
	DOT    TokenType = "DOT"    // .
	DOTDOT TokenType = "DOTDOT" // ..

	QUESTIONMARK TokenType = "QUESTIONMARK" // ?
	IF           TokenType = "IF"
	IS           TokenType = "IS"
	IFNOT        TokenType = "IFNOT" //үгүй

	WHILE TokenType = "WHILE"
	LOOP  TokenType = "LOOP"
	UNTIL TokenType = "UNTIL"
	FROM  TokenType = "FROM" // -с
	TO    TokenType = "TO"   // хүртэл

	BREAK    TokenType = "BREAK"
	CONTINUE TokenType = "CONTINUE"

	LOGICAND TokenType = "LOGICAND" // &&
	LOGICOR  TokenType = "LOGICOR"  // ||
	NOT      TokenType = "NOT"      // !

	RETURN TokenType = "RETURN"
	PRINT  TokenType = "PRINT"

	IDENT       TokenType = "IDENT"
	NUMBER      TokenType = "NUMBER"
	FN          TokenType = "FN"
	DECL        TokenType = "DECL"
	OPEN_PAREN  TokenType = "OPEN_PAREN"
	CLOSE_PAREN TokenType = "CLOSE_PAREN"
	RIGHT_ARROW TokenType = "RIGHT_ARROW"
	OPEN_BRACE  TokenType = "OPEN_BRACE"
	CLOSE_BRACE TokenType = "CLOSE_BRACE"
	SEMICOLON   TokenType = "SEMICOLON"
	COLON       TokenType = "COLON"
	TILDE       TokenType = "TILDE"
	EOF         TokenType = "EOF"

	INT_TYPE TokenType = "INT_TYPE"
	VOID     TokenType = "VOID"
	ERROR    TokenType = "ERROR"
)

type Token struct {
	Type  TokenType
	Value *string
	Line  int
	Span  Span
}

type Span struct {
	Start int
	End   int
}

func BuildToken(ttype TokenType, value *string, line int, start int, end int) Token {
	return Token{
		Type:  ttype,
		Value: value,
		Line:  line,
		Span:  Span{Start: start, End: end},
	}
}
