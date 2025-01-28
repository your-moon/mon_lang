package parser

import (
	"fmt"
	"strconv"

	"github.com/your-moon/mn_compiler_go_version/lexer"
)

type Parser struct {
	Source  []int32
	Current lexer.Token
	scanner lexer.Scanner
}

func NewParser(source []int32) Parser {
	return Parser{
		Source:  source,
		scanner: lexer.NewScanner(source),
	}
}

func (p *Parser) current() lexer.Token {
	return p.Current
}

func (p *Parser) advance() (lexer.Token, error) {
	prev := p.Current
	token, err := p.scanner.Scan()
	fmt.Println("ADVANCING TOKEN:", token.Type)
	if err != nil {
		return lexer.Token{}, err
	}
	p.Current = token
	return prev, nil
}

func (p *Parser) ParseStmt() ASTStmt {
	p.advance()
	// if err != nil {
	// 	return nil, err
	// }

	switch p.Current.Type {
	case lexer.RETURN:
		return p.ParseReturn()
	case lexer.PRINT:
		return p.ParsePrint()
	default:
		return nil
	}
}

func (p *Parser) ParsePrint() *ASTPrintStmt {
	ast := &ASTPrintStmt{
		Token: p.Current,
	}

	value := p.ParseExpr()

	ast.Value = value

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
	p.advance()
	switch p.Current.Type {
	case lexer.NUMBER:
		return p.ParseIntLit()
	default:
		return nil
		// return ASTNode{}, fmt.Errorf("not implemented expr")
	}
}

func (p *Parser) ParseIntLit() ASTExpression {
	intVal, _ := strconv.ParseInt(*p.Current.Value, 0, 64)
	ast := &ASTIntLitExpression{Token: p.Current, Value: intVal}
	return ast
}
