#include "parser.h"
#include "defs.h"
#include "tree.h"
#include <assert.h>
#include <stdlib.h>

Parser parser;

Parser initParser(const wchar_t *source)
{
    Parser parser;
    parser.source = source;
    return parser;
}

Token advance()
{
    Token prev = parser.current;
    parser.current = scanToken();
    return prev;
}

Token current()
{
    return parser.current;
}

typedef enum
{
    PREC_NONE,
    PREC_ASSIGNMENT, // =
    PREC_OR,         // or
    PREC_AND,        // and
    PREC_EQUALITY,   // == !=
    PREC_COMPARISON, // < > <= >=
    PREC_TERM,       // + -
    PREC_FACTOR,     // * /
    PREC_UNARY,      // ! -
    PREC_CALL,       // . ()
    PREC_PRIMARY
} Precedence;

typedef Token (*ParseFn)();

typedef struct
{
    ParseFn prefix;
    ParseFn infix;
    Precedence precedence;
} ParseRule;


ParseRule rules[] = {
    [T_NUMBER] = {NULL, NULL, PREC_NONE},
    [T_OPEN_PAREN] = {NULL, NULL, PREC_NONE},
    [T_BINARY_OP] = {NULL, NULL, PREC_TERM},
    [T_EOF] = {NULL, NULL, PREC_NONE}};

ParseRule *getRule(TokenType type)
{
    return &rules[type];
}

Token parsePrecedence(Precedence precedence)
{
    Token token = advance();
    ParseFn prefixRule = getRule(token.type)->prefix;
    if (prefixRule == NULL)
    {
        return token;
    }

    // Token left = prefixRule();
    //
    // while (getRule(peek().type)->precedence >= precedence)
    // {
    //     token = advanceAndGetPrev();
    //     ParseFn infixRule = getRule(token.type)->infix;
    //     left = infixRule(left);
    // }

    return token;
    // return left;
}

long toInt(const wchar_t *value) {
    wchar_t *endptr;

    long num = wcstol(value, &endptr, 10); // Base 10 conversion

    if (*endptr != L'\0') {
        wprintf(L"Conversion error, non-numeric characters found: %ls\n", endptr);
        exit(1);
    } else {
        wprintf(L"Converted number: %ld\n", num); // Output: 123
    }
    return num;
}

Token parse_expression()
{
    return parsePrecedence(PREC_ASSIGNMENT);
}

struct ASTnode *return_stmt() {
    struct ASTnode *tree;

    tree = parse_exp();

    advance();
    Token token = current();
    assert(token.type == T_SEMICOLON);

    tree = mkastunary(T_RETURN, tree, 0);
    return tree;
}

void match(int token_type) {
    advance();
    Token token = current();

    if (token.type != token_type) {
        fprintf(stderr, "Expected token type %d but got %d\n", token_type, token.type);
        assert(token.type == token_type);  // Trigger the assertion failure
    }
}



struct ASTnode *parse_fn() {
    struct ASTnode *tree;

    match(T_IDENT);
    match(T_OPEN_PAREN);
    match(T_CLOSE_PAREN);

    advance();
    Token token = current();
    if (token.type == T_RIGHTARROW) {
        match(T_INT_TYPE);
    }

    match(T_OPEN_BRACE);
    tree = parse_stmt();
    match(T_CLOSE_BRACE);

    tree = mkastunary(T_FN, tree, 0);

    return tree;

}


struct ASTnode *parse_decl() {
    struct ASTnode *tree;

    advance();
    Token token = current();
    wprintf(L"THIS IS TOKEN_VALUE %ls\n\n", token.value);
    wprintf(L"THIS IS TOKEN %d\n\n", token.type);
    switch (token.type) {
        case T_FN:
            tree = parse_fn();
            break;
        default:
            printf("We dont have decl stmt here: the token is %d\n", token.type);
            exit(1);
    };
    return tree;

}



struct ASTnode *parse_stmt() {
    struct ASTnode *n;

    advance();
    Token token = current();
    wprintf(L"THIS IS TOKEN_VALUE %ls\n\n", token.value);
    wprintf(L"THIS IS TOKEN %d\n\n", token.type);
    switch (token.type) {
        case T_SEMICOLON:
            //stop the stmt ended;
            break;
        case T_RETURN:
            n = return_stmt();
            break;
        default:
            printf("I DONT KNOW THIS STMT the token is %d\n", token.type);
            exit(1);
    };
    return n;
}

struct ASTnode *parse_exp() {
    struct ASTnode *n;

    advance();
    Token token = current();
    switch (token.type) {
        case T_NUMBER:
            n = mkastleaf((int)toInt(token.value));
            break;
        default:
            printf("I DONT KNOW THIS EXPR the token is %d\n", token.type);
            exit(1);
    };
    return n;
}

// ASTnode parseStatement()
// {
//     Token token = advanceAndGetPrev();
//     assert(token.type == T_RETURN);
//     // size_t return_val = parseExp();
//     token = advanceAndGetPrev();
//     assert(token.type == T_SEMICOLON);
//     // return
