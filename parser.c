#include "parser.h"

Parser parser;

Parser initParser(const wchar_t *source)
{
    Parser parser;
    parser.source = source;
    return parser;
}

Token advanceAndGetPrev()
{
    Token prev = parser.current;
    parser.current = scanToken();
    return prev;
}

Token peek()
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
    Token token = advanceAndGetPrev();
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

    // return left;
}

Token parseExpression()
{
    initScanner(parser.source);
    return parsePrecedence(PREC_ASSIGNMENT);
}
