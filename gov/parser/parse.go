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
	errors    []string
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

func (p *Parser) Errors() []string {
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

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.PeekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) ParseProgram() (*ASTProgram, error) {

	program := &ASTProgram{}
	program.Statements = []ASTStmt{}

	for p.Current.Type != lexer.EOF {
		stmt := p.ParseStmt()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
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
	case lexer.PRINT:
		return p.ParsePrint()
	default:
		return p.ParseExpressionStmt()
	}
}

func (p *Parser) ParseExpressionStmt() *ASTExpressionStmt {
	ast := &ASTExpressionStmt{
		Token: p.Current,
	}

	ast.Expression = p.ParseExpr()
	if base.Debug {
		fmt.Println(ast.PrintAST())
	}

	return ast
}

func (p *Parser) ParsePrint() *ASTPrintStmt {
	ast := &ASTPrintStmt{
		Token: p.Current,
	}

	value := p.ParseExpr()

	ast.Value = value

	return ast
}

func (p *Parser) ParseFN() *ASTFNStmt {
	p.NextToken()
	fnName := p.current()
	fmt.Println("FNNAME", fnName)
	ast := &ASTFNStmt{
		Token: fnName,
	}
	if !p.expect(lexer.OPEN_PAREN) {
		return nil
	}

	if !p.expect(lexer.CLOSE_PAREN) {
		return nil
	}

	return ast
}

func (p *Parser) ParseReturn() *ASTReturnStmt {
	ast := &ASTReturnStmt{
		Token: p.Current,
	}

	value := p.ParseExpr()

	ast.ReturnValue = value

	return ast
}

func (p *Parser) ParseExpr() ASTExpression {
	// p.advance()
	switch p.Current.Type {
	case lexer.PLUS:
		return p.ParseInFixExpr(lexer.PLUS)
	case lexer.NUMBER:
		return p.ParseIntLit()
	default:
		return nil
	}
}

func (p *Parser) ParseInFixExpr(op lexer.TokenType) *ASTInfixExpression {
	fmt.Println("Infix parsing", op)
	ast := &ASTInfixExpression{
		Token: p.Current,
		Op:    string(op),
	}

	left := p.ParseExpr()
	right := p.ParseExpr()
	ast.Left = left
	ast.Right = right

	return ast
}

func (p *Parser) ParseIntLit() ASTExpression {
	intVal, _ := strconv.ParseInt(*p.Current.Value, 0, 64)
	ast := &ASTIntLitExpression{Token: p.Current, Value: intVal}
	return ast
}
