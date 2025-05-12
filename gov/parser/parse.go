package parser

import (
	"fmt"
	"strconv"

	"github.com/your-moon/mn_compiler_go_version/errors"
	"github.com/your-moon/mn_compiler_go_version/lexer"
	"github.com/your-moon/mn_compiler_go_version/utfconvert"
)

type Parser struct {
	source      []int32
	current     lexer.Token
	peekToken   lexer.Token
	scanner     lexer.Scanner
	parseErrors []error
}

func NewParser(source []int32) *Parser {
	p := &Parser{
		source:  source,
		scanner: lexer.NewScanner(source),
	}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []error {
	return p.parseErrors
}

func (p *Parser) nextToken() {
	p.current = p.peekToken
	p.peekToken, _ = p.scanner.Scan()
}

func (p *Parser) checkOptional(expected lexer.TokenType) bool {
	if p.peekToken.Type == expected {
		p.nextToken()
		return true
	}
	return false
}

func (p *Parser) expect(expected lexer.TokenType) bool {
	if p.peekToken.Type == expected {
		p.nextToken()
		return true
	}
	p.peekError(expected)
	return false
}

func (p *Parser) peekIs(expected lexer.TokenType) bool {
	return p.peekToken.Type == expected
}

func (p *Parser) appendError(message string) {
	err := errors.New(message, p.current.Line, p.current.Span, p.source, "Синтакс шинжилгээ")
	p.parseErrors = append(p.parseErrors, err)
}

func (p *Parser) peekError(t lexer.TokenType) {
	message := formatExpectedNextToken(t)
	err := errors.New(message, p.current.Line, p.current.Span, p.source, "Синтакс шинжилгээ")
	p.parseErrors = append(p.parseErrors, err)
}

func (p *Parser) parseImport() *ASTImport {
	ast := &ASTImport{
		Token: p.current,
	}

	if !p.expect(lexer.IDENT) {
		p.appendError(ErrMissingIdentifier)
		return nil
	}

	ast.Ident = *p.current.Value

	for !p.peekIs(lexer.SEMICOLON) {
		if p.peekIs(lexer.DOT) {
			p.nextToken()
			if !p.expect(lexer.IDENT) {
				p.appendError(ErrMissingIdentifier)
				return nil
			}
			ast.SubImports = append(ast.SubImports, *p.current.Value)
		}
	}
	return ast
}

func (p *Parser) ParseProgram() (*ASTProgram, error) {
	program := &ASTProgram{}

	for p.current.Type != lexer.EOF {
		switch p.current.Type {
		case lexer.IMPORT:
			stmt := p.parseImport()
			if stmt != nil {
				program.Decls = append(program.Decls, stmt)
			}
			p.nextToken()
		case lexer.PUBLIC:
			p.nextToken()
			stmt := p.parseFnDecl(true)
			if stmt != nil {
				program.Decls = append(program.Decls, stmt)
			}
		case lexer.FN:
			stmt := p.parseFnDecl(false)
			if stmt != nil {
				program.Decls = append(program.Decls, stmt)
			}
		}
		p.nextToken()
	}

	if len(p.parseErrors) > 0 {
		err := p.parseErrors[0]
		p.parseErrors = nil
		return nil, err
	}

	return program, nil
}

func (p *Parser) parseBlockItem() BlockItem {
	switch p.peekToken.Type {
	case lexer.DECL:
		return p.parseVarDecl()
	case lexer.FN:
		p.appendError("функц дотор функц үүсгэж болохгүй")
		return nil
	default:
		return p.parseStmt()
	}
}

func (p *Parser) parseStmt() ASTStmt {
	switch p.peekToken.Type {
	case lexer.BREAK:
		ast := &ASTBreakStmt{
			Token: p.peekToken,
		}
		p.nextToken()
		p.expect(lexer.SEMICOLON)
		return ast
	case lexer.CONTINUE:
		ast := &ASTContinueStmt{
			Token: p.peekToken,
		}
		p.nextToken()
		p.expect(lexer.SEMICOLON)
		return ast
	case lexer.WHILE:
		return p.parseWhile()
	case lexer.LOOP:
		return p.parseLoop()
	case lexer.RETURN:
		return p.parseReturn()
	case lexer.IF:
		return p.parseIf()
	case lexer.OPEN_BRACE:
		return p.parseCompoundStmt()
	default:
		return p.parseExpressionStmt()
	}
}

func (p *Parser) parseExpressionStmt() *ExpressionStmt {
	ast := &ExpressionStmt{
		Expression: p.parseExpr(Lowest),
	}

	if !p.expect(lexer.SEMICOLON) {
		p.appendError(ErrMissingSemicolon)
		return nil
	}

	return ast
}

func (p *Parser) parseBlock() *ASTBlock {
	if !p.expect(lexer.OPEN_BRACE) {
		p.appendError(ErrMissingBraceOpen)
		return nil
	}

	items := p.parseBlockItems()

	if !p.expect(lexer.CLOSE_BRACE) {
		p.appendError(ErrMissingBraceClose)
		return nil
	}

	return &ASTBlock{
		BlockItems: items,
	}
}

func (p *Parser) parseCompoundStmt() *ASTCompoundStmt {
	block := p.parseBlock()
	if block == nil {
		return nil
	}

	return &ASTCompoundStmt{
		Block: *block,
	}
}

func (p *Parser) parseBlockItems() []BlockItem {
	var items []BlockItem

	// block duustal davtna
	for !p.peekIs(lexer.CLOSE_BRACE) {
		stmt := p.parseBlockItem()
		items = append(items, stmt)
	}

	return items
}

// <function> ::= "функц" <identifier> "(" [ <params> ] ")" "->" "тоо" "{" { <block-item> } "}"
func (p *Parser) parseFnDecl(isPublic bool) *FnDecl {
	ast := &FnDecl{
		Token:    p.current,
		IsPublic: isPublic,
	}

	if p.peekToken.Value == nil {
		p.appendError(ErrMissingIdentifier)
		return nil
	}

	ast.Ident = utfconvert.UtfConvert(*p.peekToken.Value)
	p.nextToken()

	if !p.expect(lexer.OPEN_PAREN) {
		p.appendError(errors.ErrMissingParenOpen)
		return nil
	}

	params, err := p.parseParams()
	if err != nil {
		p.appendError(err.Error())
		return nil
	}
	if !p.expect(lexer.CLOSE_PAREN) {
		p.appendError(errors.ErrMissingParenClose)
		return nil
	}

	ast.Params = params

	if !p.expect(lexer.RIGHT_ARROW) {
		p.appendError(errors.ErrMissingArrow)
		return nil
	}

	returnType, err := p.parseType()
	if err != nil {
		p.appendError(err.Error())
		return nil
	}
	p.nextToken() // consume type
	ast.ReturnType = returnType

	if p.peekIs(lexer.OPEN_BRACE) {
		block := p.parseBlock()
		if block != nil {
			ast.Body = block
		}
	}

	return ast
}

func (p *Parser) parseParams() ([]Param, error) {
	params := []Param{}

	for !p.peekIs(lexer.CLOSE_PAREN) {
		ident := p.peekToken.Value
		if !p.expect(lexer.IDENT) {
			p.appendError(ErrMissingIdentifier)
			return nil, errors.New(ErrMissingIdentifier, p.current.Line, p.current.Span, p.source, "Синтакс шинжилгээ")
		}

		params = append(params, Param{
			Token: p.peekToken,
			Ident: *ident,
			Type:  p.peekToken.Type,
		})
		p.nextToken()

		if p.peekIs(lexer.COMMA) {
			p.nextToken()
		}
	}

	return params, nil
}

func (p *Parser) parseType() (lexer.TokenType, error) {
	if !p.peekIs(lexer.INT_TYPE) && !p.peekIs(lexer.VOID) {
		return lexer.ERROR, errors.New(ErrMissingIntType, p.current.Line, p.current.Span, p.source, "Синтакс шинжилгээ")
	}

	return p.peekToken.Type, nil
}

func (p *Parser) parseVarDecl() *VarDecl {
	ast := &VarDecl{}

	p.nextToken() // consume 'зарла'

	ast.Token = p.current

	if p.peekToken.Value == nil {
		p.appendError(ErrMissingIdentifier)
		return nil
	}

	ast.Ident = *p.peekToken.Value
	p.nextToken()

	if p.checkOptional(lexer.COLON) {
		if !p.expect(lexer.INT_TYPE) {
			p.appendError(ErrMissingIntType)
			return nil
		}
	}

	if p.checkOptional(lexer.ASSIGN) {
		ast.Expr = p.parseExpr(Lowest)
	}

	if !p.expect(lexer.SEMICOLON) {
		p.appendError(ErrMissingSemicolon)
		return nil
	}

	return ast
}

func (p *Parser) parseIf() *ASTIfStmt {
	ast := &ASTIfStmt{}

	p.nextToken() // consume 'хэрэв'
	cond := p.parseExpr(Lowest)

	if !p.expect(lexer.IS) {
		p.appendError(ErrMissingIs)
		return nil
	}

	block := p.parseStmt()
	ast.Cond = cond
	ast.Then = block

	if p.checkOptional(lexer.IFNOT) {
		if !p.expect(lexer.IS) {
			p.appendError(ErrMissingIs)
			return nil
		}
		ast.Else = p.parseStmt()
	}

	return ast
}

func (p *Parser) parseWhile() *ASTWhile {
	ast := &ASTWhile{
		Token: p.peekToken,
	}

	p.nextToken() // consume 'давтах'
	// if dont have cond
	if p.peekIs(lexer.OPEN_BRACE) {
		block := p.parseBlock()
		ast.Body = *block
	} else {
		ast.Cond = p.parseExpr(Lowest)
		if ast.Cond == nil {
			return nil
		}

		if !p.expect(lexer.IS) {
			p.appendError(ErrMissingIs)
			return nil
		}

		ast.Body = *p.parseBlock()
	}

	return ast
}

func (p *Parser) parseLoop() *ASTLoop {
	ast := &ASTLoop{
		Token: p.peekToken,
	}

	p.nextToken() // consume 'давт'

	// Parse loop variable
	if !p.peekIs(lexer.IDENT) {
		p.appendError("давт түлхүүр үгний араас заавал хувьсагч байна")
		return nil
	}
	ast.Var = p.parseIdent().(*ASTVar)

	// Parse assignment
	if !p.expect(lexer.IS) {
		p.appendError("хувьсагчийн араас заавал 'бол' үг байна")
		return nil
	}

	// Parse start value
	ast.Expr = p.parseExpr(Lowest)

	// Parse 'to' keyword
	if !p.expect(lexer.UNTIL) {
		p.appendError("'хүртэл' түлхүүр үгийг оруулж өгнө үү")
		return nil
	}

	// Parse loop body as a block
	block := p.parseBlock()
	if block == nil {
		return nil
	}
	ast.Body = *block

	return ast
}

func (p *Parser) parseReturn() *ASTReturnStmt {
	ast := &ASTReturnStmt{
		Token: p.current,
	}

	p.nextToken() // consume 'буц'
	ast.ReturnValue = p.parseExpr(Lowest)

	if !p.expect(lexer.SEMICOLON) {
		p.appendError(ErrMissingSemicolon)
		return nil
	}

	return ast
}

func (p *Parser) parseFactor() ASTExpression {
	next := p.peekToken
	switch next.Type {
	case lexer.IDENT:
		return p.parseIdent()
	case lexer.NUMBER:
		return p.parseIntLit()
	case lexer.MINUS, lexer.TILDE, lexer.NOT:
		return p.parseUnary(next.Type)
	case lexer.OPEN_PAREN:
		return p.parseGrouping()
	default:
		p.appendError(formatUnknownExpression(p.peekToken.Type))
		return nil
	}
}

func (p *Parser) isInfixOp() bool {
	return p.peekToken.Type == lexer.PLUS || p.peekToken.Type == lexer.MINUS ||
		p.peekToken.Type == lexer.MUL || p.peekToken.Type == lexer.DIV ||
		p.peekToken.Type == lexer.LESSTHAN || p.peekToken.Type == lexer.LESSTHANEQUAL ||
		p.peekToken.Type == lexer.GREATERTHAN || p.peekToken.Type == lexer.GREATERTHANEQUAL ||
		p.peekToken.Type == lexer.EQUALTO || p.peekToken.Type == lexer.NOTEQUAL ||
		p.peekToken.Type == lexer.LOGICAND || p.peekToken.Type == lexer.LOGICOR || p.peekToken.Type == lexer.ASSIGN ||
		p.peekToken.Type == lexer.QUESTIONMARK || p.peekToken.Type == lexer.DOTDOT
}

const (
	_           int = iota
	Lowest          // 0
	Assign          // = (1)
	Conditional     // ? : (3)
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

func (p *Parser) parseExpr(minPrec int) ASTExpression {
	left := p.parseFactor()
	if left == nil {
		return nil
	}

	for {
		prec := p.peekPrecedence()
		if prec < minPrec || !p.isInfixOp() {
			break
		}

		op, _ := p.parseInfixOp(p.peekToken.Type)
		p.nextToken() // consume operator

		if op == ASTBinOp(A_DOTDOT) {
			right := p.parseExpr(Assign)
			left = &ASTRangeExpr{
				Token: p.current,
				Start: left,
				End:   right,
			}
			// range baival zaaval tsaash yvahgui
			return left
		}
		if op == ASTBinOp(A_ASSIGN) {
			right := p.parseExpr(Assign)
			leftIdent, ok := left.(*ASTVar)
			if !ok {
				panic("left side of assign must be var")
			}
			left = &ASTAssignment{
				Token: p.current,
				Left:  leftIdent,
				Right: right,
			}
		} else if op == ASTBinOp(A_QUESTIONMARK) {
			middle := p.parseExpr(Lowest)

			if p.peekToken.Type != lexer.COLON {
				p.appendError("'?' түлхүүрийн араас заавал ':' түлхүүр байна")
				return nil
			}
			p.nextToken() // consume colon

			right := p.parseExpr(Lowest)
			left = &ASTConditional{
				Token: p.current,
				Cond:  left,
				Then:  middle,
				Else:  right,
			}
		} else {
			// baruun tal ni urgelj iluu operator-iin precedence-tei baina
			right := p.parseExpr(prec + 1)
			if right == nil {
				return nil
			}
			left = &ASTBinary{
				Token: p.current,
				Left:  left,
				Right: right,
				Op:    op,
			}
		}
	}

	return left
}

func (p *Parser) ParseBinOp() ASTBinOp {
	next := p.peekToken
	p.nextToken()
	switch next.Type {
	case lexer.PLUS:
		return ASTBinOp(A_PLUS)
	case lexer.MINUS:
		return ASTBinOp(A_MINUS)
	case lexer.MUL:
		return ASTBinOp(A_MUL)
	case lexer.DIV:
		return ASTBinOp(A_DIV)
	case lexer.LOGICAND:
		return ASTBinOp(A_AND)
	case lexer.LOGICOR:
		return ASTBinOp(A_OR)
	case lexer.EQUALTO:
		return ASTBinOp(A_EQUALTO)
	case lexer.NOTEQUAL:
		return ASTBinOp(A_NOTEQUAL)
	case lexer.LESSTHAN:
		return ASTBinOp(A_LESSTHAN)
	case lexer.LESSTHANEQUAL:
		return ASTBinOp(A_LESSTHANEQUAL)
	case lexer.GREATERTHAN:
		return ASTBinOp(A_GREATERTHAN)
	case lexer.GREATERTHANEQUAL:
		return ASTBinOp(A_GREATERTHANEQUAL)
	default:
		panic(fmt.Sprintf("unknown bin op: %v", p.peekToken.Type))
	}
}

func (p *Parser) parseInfixOp(op lexer.TokenType) (ASTBinOp, error) {
	switch op {
	case lexer.DOTDOT:
		return ASTBinOp(A_DOTDOT), nil
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
		return ASTBinOp(A_PLUS), fmt.Errorf(formatUnknownBinOp(p.current.Type))
	}
}

func (p *Parser) parseUnary(op lexer.TokenType) *ASTUnary {
	p.nextToken() // consume the operator
	inner := p.parseExpr(Prefix)
	return &ASTUnary{
		Token: p.current,
		Op:    op,
		Inner: inner,
	}
}

func (p *Parser) parseGrouping() ASTExpression {
	p.nextToken()
	inner := p.parseExpr(Lowest)
	p.expect(lexer.CLOSE_PAREN)
	return inner
}

func (p *Parser) parseArgList() []ASTExpression {
	args := []ASTExpression{}

	for !p.peekIs(lexer.CLOSE_PAREN) {
		args = append(args, p.parseExpr(Lowest))
		if p.peekIs(lexer.COMMA) {
			p.nextToken()
		}
	}

	return args
}

func (p *Parser) parseFnCall() ASTExpression {
	cur := p.current
	p.nextToken() // consume '('

	//parse args
	args := p.parseArgList()

	p.expect(lexer.CLOSE_PAREN)

	if cur.Value == nil {
		p.appendError(ErrMissingIdentifier)
		return nil
	}

	ident := utfconvert.UtfConvert(*cur.Value)

	return &ASTFnCall{
		Token: cur,
		Ident: ident,
		Args:  args,
	}
}
func (p *Parser) parseIdent() ASTExpression {
	next := p.peekToken
	p.nextToken()

	if p.peekIs(lexer.OPEN_PAREN) {
		return p.parseFnCall()
	}

	return &ASTVar{
		Token: next,
		Ident: *next.Value,
	}
}

func (p *Parser) parseIntLit() ASTExpression {
	next := p.peekToken
	p.nextToken()
	intVal, _ := strconv.ParseInt(*next.Value, 0, 64)
	return &ASTConstant{Token: next, Value: intVal}
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}
	return Lowest
}

func (p *Parser) parseOptionalExpr() ASTExpression {
	currentToken := p.current
	peekToken := p.peekToken

	expr := p.parseExpr(Lowest)

	if expr == nil {
		p.current = currentToken
		p.peekToken = peekToken
		return nil
	}

	return expr
}

func (p *Parser) parseOptionalExprWithDelimiter(delimiter lexer.TokenType) ASTExpression {
	if p.peekIs(delimiter) {
		p.nextToken() // consume the delimiter
		return nil
	}

	expr := p.parseExpr(Lowest)
	if expr == nil {
		return nil
	}

	if !p.expect(delimiter) {
		return nil
	}

	return expr
}
