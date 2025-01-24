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
	if err != nil {
		return lexer.Token{}, err
	}
	p.Current = token
	return prev, nil
}

func (p *Parser) ParseReturn() (ASTnode, error) {
	right, err := p.ParseExpr()
	if err != nil {
		return ASTnode{}, err
	}

	return NewUnaryNode(ASTreturn, &right, 0), nil

}
func (p *Parser) ParseStmt() (ASTnode, error) {
	_, err := p.advance()
	if err != nil {
		return ASTnode{}, err
	}

	switch p.Current.Type {
	case lexer.RETURN:
		return p.ParseReturn()
	default:
		return ASTnode{}, fmt.Errorf("not implemented stmt")
	}
}

func (p *Parser) ParseExpr() (ASTnode, error) {
	_, err := p.advance()
	if err != nil {
		return ASTnode{}, err
	}
	fmt.Println("current:", p.Current)

	switch p.Current.Type {
	case lexer.NUMBER:
		as_number, err := strconv.Atoi(*p.Current.Value)
		if err != nil {
			return ASTnode{}, fmt.Errorf("cant parse the number")
		}
		return NewLeafNode(as_number), nil
	default:
		return ASTnode{}, fmt.Errorf("not implemented expr")
	}
}
