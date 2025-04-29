package lexer

type Keyword string

const (
	KeywordLoop   Keyword = "давтах"
	KeywordHurtel Keyword = "хүртэл"
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
