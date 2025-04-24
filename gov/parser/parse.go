package parser

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/your-moon/mn_compiler_go_version/base"
	"github.com/your-moon/mn_compiler_go_version/lexer"
)

type ParseError struct {
	message string
	Line    int
	Span    lexer.Span
	Source  []int32
}

func (p *ParseError) Error() string {
	var out bytes.Buffer

	// Get the line content
	lineStart := 0
	lineEnd := 0
	for i := 0; i < len(p.Source); i++ {
		if p.Source[i] == '\n' {
			if i < p.Span.Start {
				lineStart = i + 1
			}
			if i >= p.Span.End {
				lineEnd = i
				break
			}
		}
	}
	if lineEnd == 0 {
		lineEnd = len(p.Source)
	}

	lineContent := string(p.Source[lineStart:lineEnd])

	// Create error pointer
	pointer := strings.Repeat(" ", p.Span.Start-lineStart) +
		strings.Repeat("^", p.Span.End-p.Span.Start)

	out.WriteString(fmt.Sprintf("%d-р мөрөнд алдаа гарлаа:\n", p.Line))
	out.WriteString(fmt.Sprintf("%s\n", lineContent))
	out.WriteString(fmt.Sprintf("%s\n", pointer))
	out.WriteString(fmt.Sprintf("Алдааны мессеж: %s\n", p.message))

	return out.String()
}

type Parser struct {
	Source      []int32
	Current     lexer.Token
	PeekToken   lexer.Token
	scanner     lexer.Scanner
	parseErrors []ParseError
}

func NewParser(source []int32) Parser {

	parser := Parser{
		Source:  source,
		scanner: lexer.NewScanner(source),
	}
	parser.NextToken()
	parser.NextToken()
	return parser
}

func (p *Parser) Errors() []ParseError {
	return p.parseErrors
}

func (p *Parser) NextToken() {
	p.Current = p.PeekToken
	p.PeekToken, _ = p.scanner.Scan()
	// if base.Debug {
	// 	fmt.Println("ADVANCING TOKEN:", p.PeekToken)
	// }
}

func (p *Parser) checkOptional(expected lexer.TokenType) bool {
	if p.PeekToken.Type == expected {
		p.NextToken()
		return true
	}
	return false
}

func (p *Parser) expect(expected lexer.TokenType) bool {
	if p.PeekToken.Type == expected {
		p.NextToken()
		return true
	}
	p.peekError(expected)
	return false
}

func (p *Parser) currIs(expected lexer.TokenType) bool {
	return p.Current.Type == expected
}

func (p *Parser) appendError(message string) {
	p.parseErrors = append(p.parseErrors, ParseError{
		message: message,
		Line:    p.Current.Line,
		Span:    p.Current.Span,
		Source:  p.Source,
	})
}

func (p *Parser) peekError(t lexer.TokenType) {
	p.parseErrors = append(p.parseErrors, ParseError{
		message: formatExpectedNextToken(t),
		Line:    p.Current.Line,
		Span:    p.Current.Span,
		Source:  p.Source,
	})
}

// <program> ::= <function>
func (p *Parser) ParseProgram() (*ASTProgram, error) {
	program := &ASTProgram{}

	for p.Current.Type != lexer.EOF {
		stmt := p.ParseFN()
		if stmt != nil {
			program.FnDef = *stmt
		}
		p.NextToken()
	}

	if len(p.parseErrors) > 0 {
		err := p.parseErrors[0]
		p.parseErrors = nil // Clear errors to prevent double printing
		return nil, &err
	}

	return program, nil
}

// <block-item> ::= <statement> | <declaration>
func (p *Parser) ParseBlockItem() BlockItem {
	switch p.Current.Type {
	//<declaration>
	case lexer.DECL:
		return p.ParseDecl()
	// <statement>
	default:
		return p.ParseStmt()
	}
}

func (p *Parser) ParseStmt() ASTStmt {
	switch p.Current.Type {
	case lexer.RETURN:
		return p.ParseReturn()
	case lexer.IF:
		return p.ParseIf()
	default:
		return p.ParseExpressionStmt()
	}
}

func (p *Parser) ParseExpressionStmt() *ExpressionStmt {
	ast := &ExpressionStmt{}

	ast.Expression = p.ParseExpr(Lowest)
	if base.Debug {
		fmt.Println(ast.PrintAST(0))
	}

	if !p.expect(lexer.SEMICOLON) {
		return nil
	}

	return ast
}

// <function> ::= "функц" <identifier> "(" "" ")" "->" "тоо" "{" { <block-item> } "}"
func (p *Parser) ParseFN() *FNDef {
	p.NextToken()
	fnName := p.Current
	ast := FNDef{
		Token: fnName,
	}

	if !p.expect(lexer.OPEN_PAREN) {
		return nil
	}

	if !p.expect(lexer.CLOSE_PAREN) {
		return nil
	}
	if !p.expect(lexer.RIGHT_ARROW) {
		return nil
	}
	if !p.expect(lexer.INT_TYPE) {
		return nil
	}
	if !p.expect(lexer.OPEN_BRACE) {
		return nil
	}

	ast.BlockItems = p.ParseBlockItems()

	return &ast
}

// <block-item> ::= <statement> | <declaration>
func (p *Parser) ParseBlockItems() []BlockItem {
	var stmtarr []BlockItem

	p.NextToken()

	for !p.currIs(lexer.CLOSE_BRACE) && !p.currIs(lexer.EOF) {
		stmt := p.ParseBlockItem()
		if stmt != nil {
			stmtarr = append(stmtarr, stmt)
		}
		p.NextToken()
	}

	return stmtarr
}

// <declaration> ::= "зарла" <identifier> [ ":" "тоo" ] "=" <exp> ";"
func (p *Parser) ParseDecl() *Decl {
	ast := &Decl{}

	p.NextToken() // consume 'зарла'

	if p.Current.Value == nil {
		p.appendError(ErrMissingIdentifier)
	}

	ident := p.Current

	//has [ ":" "тоо" ]
	if p.checkOptional(lexer.COLON) {
		if !p.expect(lexer.INT_TYPE) {
			p.appendError(ErrMissingIntType)
			return nil
		}
	}

	ast.Ident = *ident.Value

	if !p.checkOptional(lexer.ASSIGN) {
		if !p.expect(lexer.SEMICOLON) {
			p.appendError(ErrMissingSemicolon)
			return nil
		}
		return ast
	}

	p.NextToken()
	ast.Expr = p.ParseExpr(Lowest)

	if !p.expect(lexer.SEMICOLON) {
		p.appendError(ErrMissingSemicolon)
		return nil
	}

	return ast
}

func (p *Parser) ParseIf() *ASTIfStmt {
	ast := &ASTIfStmt{}

	p.NextToken() // consume 'хэрэв'
	cond := p.ParseExpr(Lowest)
	if !p.expect(lexer.IS) {
		p.appendError(ErrMissingIs)
		return nil
	}

	if !p.expect(lexer.OPEN_BRACE) {
		p.appendError(ErrMissingBraceOpen)
		return nil
	}

	p.NextToken()
	then := p.ParseStmt() // parse then
	if !p.expect(lexer.CLOSE_BRACE) {
		p.appendError(ErrMissingBraceClose)
		return nil
	}

	var klse ASTStmt
	if p.checkOptional(lexer.IFNOT) {
		p.NextToken() // consume 'бол'

		if !p.expect(lexer.OPEN_BRACE) {
			p.appendError(ErrMissingBraceOpen)
			return nil
		}
		p.NextToken()

		klse = p.ParseStmt() // parse else
		if !p.expect(lexer.CLOSE_BRACE) {
			p.appendError(ErrMissingBraceClose)
			return nil
		}
	}

	ast.Cond = cond
	ast.Then = then
	ast.Else = klse

	return ast
}

func (p *Parser) ParseReturn() *ASTReturnStmt {
	ast := &ASTReturnStmt{
		Token: p.Current,
	}

	p.NextToken() // consume 'буц'
	value := p.ParseExpr(Lowest)

	ast.ReturnValue = value

	if !p.expect(lexer.SEMICOLON) {
		p.appendError(ErrMissingSemicolon)
		return nil
	}

	return ast
}

// <factor> ::= <int> | <identifier> | <unop> <factor> | "(" <exp> ")"
func (p *Parser) ParseFactor() ASTExpression {
	switch p.Current.Type {
	case lexer.IDENT:
		return p.ParseIdent()
	case lexer.NUMBER:
		return p.ParseIntLit()
	case lexer.MINUS, lexer.TILDE, lexer.NOT:
		return p.ParseUnary(p.Current.Type)
	case lexer.OPEN_PAREN:
		return p.ParseGrouping()
	default:
		p.appendError(formatUnknownExpression(p.Current.Type))
		return nil
	}
}

func (p *Parser) IsInfixOp() bool {
	return p.PeekToken.Type == lexer.PLUS || p.PeekToken.Type == lexer.MINUS ||
		p.PeekToken.Type == lexer.MUL || p.PeekToken.Type == lexer.DIV ||
		p.PeekToken.Type == lexer.LESSTHAN || p.PeekToken.Type == lexer.LESSTHANEQUAL ||
		p.PeekToken.Type == lexer.GREATERTHAN || p.PeekToken.Type == lexer.GREATERTHANEQUAL ||
		p.PeekToken.Type == lexer.EQUALTO || p.PeekToken.Type == lexer.NOTEQUAL ||
		p.PeekToken.Type == lexer.LOGICAND || p.PeekToken.Type == lexer.LOGICOR || p.PeekToken.Type == lexer.ASSIGN ||
		p.PeekToken.Type == lexer.QUESTIONMARK || p.PeekToken.Type == lexer.COLON

}

const (
	_           int = iota
	Lowest          // 0
	Assign          // = (1)
	Conditional     // ? (3)
	LogicOr         // || (5)
	LogicAnd        // && (10)
	Equals          // == != (30)
	Compare         // < <= > >= (35)
	Sum             // + - (45)
	Product         // * / % (50)
	Prefix          // -X or !X
	Call            // myFunction(X)
)

var precedences = map[lexer.TokenType]int{
	lexer.PLUS:             Sum,
	lexer.MINUS:            Sum,
	lexer.DIV:              Product,
	lexer.MUL:              Product,
	lexer.LESSTHAN:         Compare,
	lexer.LESSTHANEQUAL:    Compare,
	lexer.GREATERTHAN:      Compare,
	lexer.GREATERTHANEQUAL: Compare,
	lexer.EQUALTO:          Equals,
	lexer.NOTEQUAL:         Equals,
	lexer.LOGICAND:         LogicAnd,
	lexer.LOGICOR:          LogicOr,
	lexer.QUESTIONMARK:     Conditional,
	lexer.ASSIGN:           Assign,
}

func (p *Parser) ParseExpr(minPrec int) ASTExpression {
	left := p.ParseFactor()

	for !p.currIs(lexer.SEMICOLON) && minPrec < p.peekPrecedence() {
		if !p.IsInfixOp() {
			return left
		}

		p.NextToken() // consume the operator
		op, _ := p.ParseBinOp(p.Current.Type)
		nextPrec := p.currPrecedence()

		if op == ASTBinOp(A_ASSIGN) {
			p.NextToken()
			right := p.ParseExpr(nextPrec)
			leftIdent, ok := left.(*ASTVar)
			if !ok {
				panic("left side of assign must be var")
			}
			left = &ASTAssignment{
				Left:  leftIdent,
				Right: right,
			}
		} else if op == ASTBinOp(A_QUESTIONMARK) {
			p.NextToken()
			middle := p.ParseExpr(Lowest)

			if !p.expect(lexer.COLON) {
				panic("expected colon after middle expression in ternary")
			}
			p.NextToken()
			right := p.ParseExpr(nextPrec)
			left = &ASTConditional{
				Cond: left,
				Then: middle,
				Else: right,
			}
		} else {
			p.NextToken()                  // consume the right operand
			right := p.ParseExpr(nextPrec) // Use nextPrec for proper precedence
			left = &ASTBinary{
				Left:  left,
				Right: right,
				Op:    op,
			}
		}
	}

	return left
}

func (p *Parser) ParseBinOp(op lexer.TokenType) (ASTBinOp, error) {
	switch p.Current.Type {
	case lexer.QUESTIONMARK:
		return ASTBinOp(A_QUESTIONMARK), nil
	case lexer.MINUS:
		return ASTBinOp(A_MINUS), nil
	case lexer.PLUS:
		return ASTBinOp(A_PLUS), nil
	case lexer.DIV:
		return ASTBinOp(A_DIV), nil
	case lexer.MUL:
		return ASTBinOp(A_MUL), nil
	case lexer.LOGICAND:
		return ASTBinOp(A_AND), nil
	case lexer.LOGICOR:
		return ASTBinOp(A_OR), nil
	case lexer.ASSIGN:
		return ASTBinOp(A_ASSIGN), nil
	case lexer.EQUALTO:
		return ASTBinOp(A_EQUALTO), nil
	case lexer.NOTEQUAL:
		return ASTBinOp(A_NOTEQUAL), nil
	case lexer.LESSTHAN:
		return ASTBinOp(A_LESSTHAN), nil
	case lexer.LESSTHANEQUAL:
		return ASTBinOp(A_LESSTHANEQUAL), nil
	case lexer.GREATERTHAN:
		return ASTBinOp(A_GREATERTHAN), nil
	case lexer.GREATERTHANEQUAL:
		return ASTBinOp(A_GREATERTHANEQUAL), nil
	default:
		return ASTBinOp(A_PLUS), fmt.Errorf(formatUnknownBinOp(p.Current.Type))
	}
}

func (p *Parser) ParseUnary(op lexer.TokenType) *ASTUnary {
	ast := &ASTUnary{
		Op: op,
	}

	p.NextToken()
	right := p.ParseExpr(Lowest)

	ast.Inner = right

	return ast
}
func (p *Parser) ParseGrouping() ASTExpression {
	p.NextToken()
	inner := p.ParseExpr(Lowest)
	p.expect(lexer.CLOSE_PAREN)
	return inner
}

func (p *Parser) ParseIdent() ASTExpression {
	ast := &ASTVar{Ident: *p.Current.Value}
	return ast
}

func (p *Parser) ParseIntLit() ASTExpression {
	intVal, _ := strconv.ParseInt(*p.Current.Value, 0, 64)
	ast := &ASTConstant{Token: p.Current, Value: intVal}
	return ast
}

func (p *Parser) currPrecedence() int {
	if p, ok := precedences[p.Current.Type]; ok {
		return p
	}
	return Lowest
}
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.PeekToken.Type]; ok {
		return p
	}
	return Lowest
}
