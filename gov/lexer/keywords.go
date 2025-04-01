package lexer

type Keyword string

const (
	KeywordPrint  Keyword = "хэвлэ"
	KeywordReturn Keyword = "буц"
	KeywordFn     Keyword = "фн"
	KeywordDecl   Keyword = "зарла"
	KeywordInt    Keyword = "тоо"
	KeywordVoid   Keyword = "хоосон"
)
