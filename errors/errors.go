package errors

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/your-moon/mon_lang/lexer"
)

type CompilerError struct {
	Message string
	Line    int
	Span    lexer.Span
	Source  []int32
	Module  string
}

func New(message string, line int, span lexer.Span, source []int32, module string) *CompilerError {
	return &CompilerError{
		Message: message,
		Line:    line,
		Span:    span,
		Source:  source,
		Module:  module,
	}
}

func (e *CompilerError) Error() string {
	var buf bytes.Buffer

	lineStart, lineEnd := e.findLineBoundaries()
	lineContent := string(e.Source[lineStart:lineEnd])
	pointer := e.createErrorPointer(lineStart)

	fmt.Fprintf(&buf, "[%s] %d-р мөрөнд алдаа гарлаа:\n", e.Module, e.Line)
	fmt.Fprintf(&buf, "%s\n", lineContent)
	fmt.Fprintf(&buf, "%s\n", pointer)
	fmt.Fprintf(&buf, "Алдааны мессеж: %s\n", e.Message)

	return buf.String()
}

func (e *CompilerError) findLineBoundaries() (start, end int) {
	start = 0
	end = 0
	for i := 0; i < len(e.Source); i++ {
		if e.Source[i] == '\n' {
			if i < e.Span.Start {
				start = i + 1
			}
			if i >= e.Span.End {
				end = i
				break
			}
		}
	}
	if end == 0 {
		end = len(e.Source)
	}
	return start, end
}

func (e *CompilerError) createErrorPointer(lineStart int) string {
	return strings.Repeat(" ", e.Span.Start-lineStart) +
		strings.Repeat("^", e.Span.End-e.Span.Start)
}

const (
	ErrExpectedNextToken = "дараагийн тэмдэгт '%s' байх ёстой, та бичиглэлээ шалгана уу"
	ErrMissingIdentifier = "хувьсагчийн нэр заавал заагдах ёстой"
	ErrUnknownExpression = "илэрхийллийн төрөл тодорхойгүй байна: '%s'"
	ErrUnknownBinOp      = "үл мэдэгдэх үйлдлийн тэмдэгт: '%s'"
	ErrMissingBraceOpen  = "'{' хаалт шаардлагатай"
	ErrMissingBraceClose = "'}' хаалт шаардлагатай"
	ErrMissingSemicolon  = "';' тэмдэгт шаардлагатай"
	ErrMissingIs         = "'бо' гэсэн тэмдэгт шаардлагатай"
	ErrMissingColon      = "':' тэмдэгт шаардлагатай"
	ErrMissingParenOpen  = "'(' хаалт шаардлагатай"
	ErrMissingParenClose = "')' хаалт шаардлагатай"
	ErrMissingArrow      = "'->' тэмдэгт шаардлагатай"
	ErrMissingIntType    = "'тоо' төрөл шаардлагатай"

	ErrNotValidExpression         = "буруу илэрхийлэл байна"
	ErrDuplicateVariable          = "хувьсагч '%s' нь давхардсан байна"
	ErrDuplicateFnDecl            = "функц '%s' нь давхардсан байна"
	ErrInvalidAssignment          = "хувьсагчид утга оноох үед зүүн талд хувьсагч байх ёстой, олдсон: '%s'"
	ErrFnDeclCanNotBeInsideFnDecl = "функц дотор функц үүсгэж болохгүй: '%s'"
	ErrFnDeclCanNotBeInsideBlock  = "блок дотор функц үүсгэж болохгүй: '%s'"
	ErrUndeclaredVariable         = "хувьсагч '%s'-г зарлаагүй байна"
	ErrNotDeclaredFnCall          = "функц '%s'-г зарлаагүй байна"
)

func FormatError(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
