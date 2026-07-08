/*
 * mon_lang - lexer
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package lexer

type Keyword string

const (
	KeywordExtern Keyword = "extern"
	KeywordStatic Keyword = "статик"

	KeywordImport   Keyword = "ашигла"
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
	KeywordUInt   Keyword = "этоо"   // unsigned 32-bit (эерэг тоо)
	KeywordULong  Keyword = "этоо64" // unsigned 64-bit
	KeywordChar   Keyword = "тэмдэгт" // Unicode codepoint, alias of тоо
	KeywordStruct Keyword = "бүтэц"
	KeywordImpl   Keyword = "хэрэгжүүл"
	KeywordMatch  Keyword = "тааруул"
	KeywordAs     Keyword = "гэж"
	KeywordSelf   Keyword = "өөрөө"
	KeywordVoid   Keyword = "хоосон"
	KeywordString Keyword = "мөр"
	KeywordNew    Keyword = "шинэ"
	KeywordElse   Keyword = "эсвэл"
)
