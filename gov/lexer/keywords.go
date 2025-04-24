package lexer

type Keyword string

const (
	KeywordPrint  Keyword = "хэвлэ"
	KeywordReturn Keyword = "буц"
	KeywordFn     Keyword = "функц"
	KeywordDecl   Keyword = "зарла"
	KeywordIf     Keyword = "хэрэв"
	KeywordIs     Keyword = "бол"
	KeywordNot    Keyword = "үгүй"
	KeywordInt    Keyword = "тоо"
	KeywordVoid   Keyword = "хоосон"
)
