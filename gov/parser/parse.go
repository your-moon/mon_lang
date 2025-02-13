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

func (p *Parser) currIs(expected lexer.TokenType) bool {
	return p.Current.Type == expected
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
	fnName := p.Current
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
	if !p.expect(lexer.RIGHT_ARROW) {
		return nil
	}
	if !p.expect(lexer.INT_TYPE) {
		return nil
	}
	if !p.expect(lexer.OPEN_BRACE) {
		return nil
	}

	ast.Block = p.ParseBlockStmt()

	return ast
}

func (p *Parser) ParseBlockStmt() *ASTBlockStmt {
	ast := &ASTBlockStmt{
		Token:      p.Current,
		Statements: []ASTStmt{},
	}
	p.NextToken()

	for !p.currIs(lexer.CLOSE_BRACE) && !p.currIs(lexer.EOF) {
		stmt := p.ParseStmt()
		if stmt != nil {
			ast.Statements = append(ast.Statements, stmt)
		}
		p.NextToken()
	}

	return ast
}

func (p *Parser) ParseReturn() *ASTReturnStmt {
	ast := &ASTReturnStmt{
		Token: p.Current,
	}

	p.NextToken()
	value := p.ParseExpr()

	ast.ReturnValue = value

	if !p.expect(lexer.SEMICOLON) {
		return nil
	}

	return ast
}

func (p *Parser) ParseExpr() ASTExpression {
	switch p.Current.Type {
	case lexer.NUMBER:
		return p.ParseIntLit()
	default:
		panic(fmt.Sprintf("dont know this expr %s", p.Current.Type))
		return nil
	}
}

func (p *Parser) ParseInFixExpr(op lexer.TokenType) *ASTInfixExpression {
	panic("not implemented")
}

func (p *Parser) ParseIntLit() ASTExpression {
	intVal, _ := strconv.ParseInt(*p.Current.Value, 0, 64)
	ast := &ASTIntLitExpression{Token: p.Current, Value: intVal}
	return ast
}
