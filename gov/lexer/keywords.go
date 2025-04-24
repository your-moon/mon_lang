package lexer

type Keyword string

const (
	KeywordPrint  Keyword = "хэвлэ"
	KeywordReturn Keyword = "буц"
	KeywordFn     Keyword = "функц"
	KeywordDecl   Keyword = "зарла"
	KeywordInt    Keyword = "тоо"
	KeywordVoid   Keyword = "хоосон"
)
