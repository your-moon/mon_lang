package parser

import (
	"fmt"
	"strconv"

	"github.com/your-moon/mn_compiler_go_version/base"
	"github.com/your-moon/mn_compiler_go_version/lexer"
)

type Parser struct {
	Source    []int32
	Current   lexer.Token
	PeekToken lexer.Token
	scanner   lexer.Scanner
	errors    []error
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

func (p *Parser) Errors() []error {
	return p.errors
}

func (p *Parser) NextToken() {
	p.Current = p.PeekToken
	p.PeekToken, _ = p.scanner.Scan()
	if base.Debug {
		fmt.Println("ADVANCING TOKEN:", p.PeekToken)
	}
}

func (p *Parser) current() lexer.Token {
	return p.Current
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

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Errorf("expected next token to be %s, got %s instead", t, p.PeekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) ParseProgram() (*ASTProgram, error) {

	program := &ASTProgram{}

	for p.Current.Type != lexer.EOF {
		stmt := p.ParseFN()
		if stmt != nil {
			program.FnDef = *stmt
		}
		p.NextToken()
	}

	return program, nil
}

func (p *Parser) ParseStmt() ASTStmt {
	switch p.Current.Type {
	case lexer.FN:
		stmt := p.ParseFN()
		if stmt == nil {
			return nil
		}
		return stmt
	case lexer.RETURN:
		return p.ParseReturn()
	default:
		return p.ParseExpressionStmt()
	}
}

func (p *Parser) ParseExpressionStmt() *ASTExpressionStmt {
	ast := &ASTExpressionStmt{
		Token: p.Current,
	}

	ast.Expression = p.ParseExpr(Lowest)
	if base.Debug {
		fmt.Println(ast.PrintAST(0))
	}

	if !p.expect(lexer.SEMICOLON) {
		return nil
	}

	return ast
}

func (p *Parser) ParseFN() *ASTFNStmt {
	p.NextToken()
	fnName := p.Current
	fmt.Println("FNNAME", fnName)
	ast := ASTFNStmt{
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

	ast.Stmt = p.ParseBlockStmt()[0]

	return &ast
}

func (p *Parser) ParseBlockStmt() []ASTStmt {
	var stmtarr []ASTStmt

	p.NextToken()

	for !p.currIs(lexer.CLOSE_BRACE) && !p.currIs(lexer.EOF) {
		stmt := p.ParseStmt()
		if stmt != nil {
			stmtarr = append(stmtarr, stmt)
		}
		p.NextToken()
	}

	return stmtarr
}

func (p *Parser) ParseReturn() *ASTReturnStmt {
	ast := &ASTReturnStmt{
		Token: p.Current,
	}

	p.NextToken() // consume 'буц'
	value := p.ParseExpr(Lowest)

	ast.ReturnValue = value

	if !p.expect(lexer.SEMICOLON) {
		return nil
	}

	return ast
}

func (p *Parser) ParseFactor() ASTExpression {
	switch p.Current.Type {
	case lexer.NUMBER:
		return p.ParseIntLit()
	case lexer.MINUS, lexer.TILDE, lexer.NOT:
		return p.ParseUnary(p.Current.Type)
	case lexer.OPEN_PAREN:
		return p.ParseGrouping()
	default:
		panic(fmt.Sprintf("dont know this expr %s", p.Current.Type))
	}
}

func (p *Parser) IsInfixOp() bool {
	return p.PeekToken.Type == lexer.PLUS || p.PeekToken.Type == lexer.MINUS ||
		p.PeekToken.Type == lexer.MUL || p.PeekToken.Type == lexer.DIV ||
		p.PeekToken.Type == lexer.LESSTHAN || p.PeekToken.Type == lexer.LESSTHANEQUAL ||
		p.PeekToken.Type == lexer.GREATERTHAN || p.PeekToken.Type == lexer.GREATERTHANEQUAL ||
		p.PeekToken.Type == lexer.EQUALTO || p.PeekToken.Type == lexer.NOTEQUAL ||
		p.PeekToken.Type == lexer.LOGICAND || p.PeekToken.Type == lexer.LOGICOR
}

func (p *Parser) ParseExpr(prec int) ASTExpression {
	left := p.ParseFactor()
	for !p.currIs(lexer.SEMICOLON) && prec < p.peekPrecedence() {
		if !p.IsInfixOp() {
			return left
		}

		p.NextToken() // consume the operator
		op, _ := p.ParseBinOp(p.Current.Type)
		currPrec := p.currPrecedence()
		p.NextToken() // consume the right operand
		right := p.ParseExpr(currPrec)

		left = &ASTBinary{
			Left:  left,
			Right: right,
			Op:    op,
		}
	}

	return left
}

func (p *Parser) ParseBinOp(op lexer.TokenType) (ASTBinOp, error) {
	switch p.Current.Type {
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
		return ASTBinOp(A_PLUS), fmt.Errorf("unknown bin op: %s", p.Current.Type)
	}
}

func (p *Parser) ParseUnary(op lexer.TokenType) *ASTUnary {
	ast := &ASTUnary{
		Op: op,
	}

	right := p.ParseFactor()

	ast.Inner = right

	return ast
}
func (p *Parser) ParseGrouping() ASTExpression {
	p.NextToken()
	inner := p.ParseExpr(Lowest)
	p.expect(lexer.CLOSE_PAREN)
	return inner
}

func (p *Parser) ParseInFixExpr(op lexer.TokenType) *ASTInfixExpression {
	panic("not implemented")
}

func (p *Parser) ParseIntLit() ASTExpression {
	intVal, _ := strconv.ParseInt(*p.Current.Value, 0, 64)
	ast := &ASTIntLitExpression{Token: p.Current, Value: intVal}
	return ast
}

const (
	_ int = iota
	Lowest
	Logic         // && ||
	Equals        // =
	LessOrGreater // < or >
	Sum           // +
	Product       // *
	Prefix        // -X or !X
	Call          // myFunction(X)
	Index         // array[index]
)

var precedences = map[lexer.TokenType]int{
	lexer.PLUS:             Sum,
	lexer.MINUS:            Sum,
	lexer.DIV:              Product,
	lexer.MUL:              Product,
	lexer.LESSTHAN:         LessOrGreater,
	lexer.LESSTHANEQUAL:    LessOrGreater,
	lexer.GREATERTHAN:      LessOrGreater,
	lexer.GREATERTHANEQUAL: LessOrGreater,
	lexer.EQUALTO:          Equals,
	lexer.NOTEQUAL:         Equals,
	lexer.LOGICAND:         Logic,
	lexer.LOGICOR:          Logic,
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
