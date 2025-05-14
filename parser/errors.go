package parser

import (
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/lexer"
)

// Error messages in Mongolian
const (
	// Token related errors
	ErrExpectedNextToken = "дараагийн тэмдэгт '%s' байх ёстой, та бичиглэлээ шалгана уу"
	ErrMissingIdentifier = "хувьсагчийн нэр заавал заагдах ёстой"
	ErrUnknownExpression = "илэрхийллийн төрөл тодорхойгүй байна: '%s'"
	ErrUnknownBinOp      = "үл мэдэгдэх үйлдлийн тэмдэгт: '%s'"

	// Syntax errors
	ErrMissingBraceOpen  = "'{' хаалт шаардлагатай"
	ErrMissingBraceClose = "'}' хаалт шаардлагатай"
	ErrMissingSemicolon  = "';' тэмдэгт шаардлагатай"
	ErrMissingIs         = "'бо' гэсэн тэмдэгт шаардлагатай"
	ErrMissingColon      = "':' тэмдэгт шаардлагатай"
	ErrMissingParenOpen  = "'(' хаалт шаардлагатай"
	ErrMissingParenClose = "')' хаалт шаардлагатай"
	ErrMissingArrow      = "'->' тэмдэгт шаардлагатай"
	ErrMissingIntType    = "'тоо' төрөл шаардлагатай"
)

// Token type translations to Mongolian
var tokenTranslations = map[lexer.TokenType]string{
	lexer.IS:               "бол",
	lexer.IDENT:            "",
	lexer.NUMBER:           "тоо",
	lexer.PLUS:             "+",
	lexer.MINUS:            "-",
	lexer.MUL:              "*",
	lexer.DIV:              "/",
	lexer.LOGICAND:         "&&",
	lexer.LOGICOR:          "||",
	lexer.ASSIGN:           "=",
	lexer.EQUALTO:          "==",
	lexer.NOTEQUAL:         "!=",
	lexer.LESSTHAN:         "<",
	lexer.LESSTHANEQUAL:    "<=",
	lexer.GREATERTHAN:      ">",
	lexer.GREATERTHANEQUAL: ">=",
	lexer.QUESTIONMARK:     "?",
	lexer.COLON:            ":",
	lexer.SEMICOLON:        ";",
	lexer.OPEN_BRACE:       "{",
	lexer.CLOSE_BRACE:      "}",
	lexer.OPEN_PAREN:       "(",
	lexer.CLOSE_PAREN:      ")",
	lexer.RIGHT_ARROW:      "->",
	lexer.INT_TYPE:         "тоо",
	lexer.IF:               "хэрэв",
	lexer.IFNOT:            "бол",
	lexer.RETURN:           "буц",
	lexer.DECL:             "зарла",
}

// Error message formatters
func formatExpectedNextToken(expected lexer.TokenType) string {
	expectedStr := getTokenTranslation(expected)
	return formatError(ErrExpectedNextToken, expectedStr)
}

func formatUnknownExpression(expr lexer.TokenType) string {
	return formatError(ErrUnknownExpression, getTokenTranslation(expr))
}

func formatUnknownBinOp(op lexer.TokenType) string {
	return formatError(ErrUnknownBinOp, getTokenTranslation(op))
}

// Helper function to get token translation
func getTokenTranslation(t lexer.TokenType) string {
	if translation, ok := tokenTranslations[t]; ok {
		return translation
	}
	return string(t)
}

// Helper function to format error messages
func formatError(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
