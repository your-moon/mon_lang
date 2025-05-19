package lexer

type Keyword string

const (
	KeywordExtern Keyword = "extern"
	KeywordStatic Keyword = "статик"

	KeywordImport   Keyword = "импорт"
	KeywordPublic   Keyword = "тунх"
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
	KeywordInt    Keyword = "тоо"
	KeywordLong   Keyword = "тоо64"
	KeywordVoid   Keyword = "хоосон"
	KeywordString Keyword = "мөр"
)
