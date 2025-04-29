package lexer

type Keyword string

const (
	KeywordBreak    Keyword = "зогс"
	KeywordContinue Keyword = "үргэлжлүүл"
	KeywordLoop     Keyword = "давт"
	KeywordUntil    Keyword = "хүртэл"
	KeywordWhile    Keyword = "давтах"
	KeywordFrom     Keyword = "-с"

	KeywordReturn Keyword = "буц"
	KeywordFn     Keyword = "функц"
	KeywordDecl   Keyword = "зарла"
	KeywordIf     Keyword = "хэрэв"
	KeywordIs     Keyword = "бол"
	KeywordNot    Keyword = "үгүй"
	//type
	KeywordInt  Keyword = "тоо"
	KeywordVoid Keyword = "хоосон"
)
