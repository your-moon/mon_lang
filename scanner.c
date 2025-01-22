#define DEBUG_SCANNER 0
#ifdef DEBUG_SCANNER
#endif
#include <string.h>
#include <stdbool.h>
#include "scanner.h"

Scanner scanner;

void initScanner(const wchar_t *source)
{
    if (DEBUG_SCANNER)
        wprintf(L"Initializing scanner\n");
    // this one points to start of literals
    // Hello
    // ^ address
    scanner.start = source;
    scanner.current = source;
    scanner.line = 1;
}

static Token fromEnum(TokenType type)
{
    Token token;
    token.type = type;
    token.line = scanner.line;
    token.start = scanner.start;
    token.length = (scanner.current - scanner.start);

    wchar_t valbuild[token.length];

    for (int i=0; i<token.length; i++) {
        valbuild[i] = scanner.start[i];
    }

    token.value = valbuild;
    return token;
}

static wchar_t peek()
{
    return *scanner.current;
}
static wchar_t nextTimes(int time)
{
    scanner.current += time;
    return scanner.current[-1];
}

static wchar_t next()
{
    scanner.current++;
    return scanner.current[-1];
}

static void skipWhitespace()
{
    wchar_t c = peek();
    if (c == ' ')
    {
        next();
    }
}

static void trimWhitespace()
{
    next();
}

static bool isAtEnd()
{
    return *scanner.current == '\0';
}

static bool isDigit(wchar_t c)
{
    return c >= '0' && c <= '9';
}

bool checkKeyword(int length, const wchar_t *rest, TokenType tokenType)
{
    if (memcmp(scanner.start, rest, length) == 0)
    {
        return true;
    }
    return false;
}

static bool isAlpha(wchar_t c)
{
    if (c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z')
    {
        return true;
    }
    if (c >= L'а' && c <= L'я' || c >= L'А' && c <= L'Я')
    {
        return true;
    }

    if (c == L'ё' || c == L'ү' || c == L'е')
        return true;

    return false;
}

Token scanKeyword()
{
    if (checkKeyword(3, L"буц", T_TRUE))
    {
        nextTimes(2);
        return fromEnum(T_RETURN);
    }
    if (checkKeyword(4, L"үнэн", T_TRUE))
    {
        nextTimes(3);
        return fromEnum(T_TRUE);
    }
    if (checkKeyword(5, L"худал", T_FALSE))
    {
        nextTimes(4);
        return fromEnum(T_FALSE);
    }
    if (checkKeyword(4, L"зарл", T_LET))
    {
        nextTimes(3);
        return fromEnum(T_LET);
    }
    if (checkKeyword(2, L"фн", T_FN))
    {
        nextTimes(1);
        return fromEnum(T_FN);
    }
    if (checkKeyword(7, L"оруулах", T_IMPORT))
    {
        nextTimes(6);
        return fromEnum(T_IMPORT);
    }
    if (checkKeyword(6, L"хоосон", T_VOID_TYPE))
    {
        nextTimes(5);
        return fromEnum(T_VOID_TYPE);
    }
    if (checkKeyword(3, L"тоо", T_INT_TYPE))
    {
        nextTimes(2);
        return fromEnum(T_INT_TYPE);
    }
    if (checkKeyword(4, L"хэрв", T_IF))
    {
        nextTimes(3);
        return fromEnum(T_IF);
    }
    return fromEnum(T_ERR);
}

Token scanIdent()
{
    Token token = scanKeyword();

    if (token.type != T_ERR)
    {
        return token;
    }

    while (isAlpha(peek()) || isDigit(peek()) || peek() == '_')
    {
        next();
    }

    return fromEnum(T_IDENT);
}

// algorithm
//   123123
//   ^     ^ -> whitespace occurs that means halt and gen token
//   T_NUMBER
Token buildNumber()
{
    while (isDigit(peek()))
        next();
    return fromEnum(T_NUMBER);
}

static void advanceLine()
{
    wchar_t c = peek();
    if (c == '\r' || c == '\n')
    {
        scanner.line++;
        next();
    }
}
static Token scanString()
{
    while (peek() != '"' && !isAtEnd())
    {
        if (peek() == '\n')
        {
            scanner.line++;
        }
        next();
    }

    if (isAtEnd())
    {
        return fromEnum(T_ERR);
    }

    next();
    return fromEnum(T_STRING);
}

// check current wchar_t and find longest matching token
Token scanToken()
{
    advanceLine();
    skipWhitespace();
    scanner.start = scanner.current;

    if (isAtEnd())
        return fromEnum(T_EOF);

    wchar_t c = next();
    // wprintf(L"CURRENT CHAR %lc\n", c);
    // wprintf(L"START FROM %ls\n", scanner.start);
    // wprintf(L"ISALPHA %ld\n", isAlpha(c));

    while (isAlpha(c))
    {
        const wchar_t *word = scanner.start;
        wprintf(L"\n%ls\n", word);
        return scanIdent();
    }

    // find the longest matching token
    //  for now at least find digits and operators
    while (isDigit(c))
        return buildNumber();

    switch (c)
    {
    case '+':
    case '*':
    case '/':
        return fromEnum(T_BINARY_OP);
        break;
    case '-':
        if (peek() == '>')
        {
            next();
            return fromEnum(T_RIGHTARROW);
        }
        return fromEnum(T_BINARY_OP);
        break;
    case '=':
        if (peek() == '=')
        {
            next();
            return fromEnum(T_EQUAL);
        }
        break;
    case '(':
        return fromEnum(T_OPEN_PAREN);
    case ')':
        return fromEnum(T_CLOSE_PAREN);
    case '{':
        return fromEnum(T_OPEN_BRACE);
    case '}':
        return fromEnum(T_CLOSE_BRACE);
    case ';':
        return fromEnum(T_SEMICOLON);
    case ',':
        return fromEnum(T_COMMA);
    case '>':
    case '<':
        return fromEnum(T_BINARY_OP);
    case '"':
        return scanString();
    }

    return fromEnum(T_ERR);
}
