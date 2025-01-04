#define DEBUG_SCANNER 1
#ifdef DEBUG_SCANNER
#include <stdio.h>
#endif
#include <stdbool.h>
#include "scanner.h"


Scanner scanner;

void initScanner(const char *source)
{
    if (DEBUG_SCANNER)
        printf("Initializing scanner\n");
    // this one points to start of literals
    // Hello
    // ^ address
    scanner.start = source;
    scanner.current = source;
    scanner.line = 1;
}

static Token fromEnum(TokenType type) {
    Token token;
    token.type = type;
    token.line = scanner.line;
    token.start = scanner.start;
    token.length = (scanner.current - scanner.start);
    return token;
}


static bool isWhitespace(char c) {
    return c == ' ';
}


static char peek() {
    return *scanner.current;
}

// Дараагийн утга ийг шилжүүлж өмнөх утгийг авна.
static char next() {
    scanner.current++;
    return scanner.current[-1];
}

static void trimWhitespace() {
    next();
}

static bool isAtEnd() {
  return *scanner.current == '\0';
}

static bool isDigit(char c) {
    return c >= '0' && c <= '9';
}
//algorithm
//bunch of literals
//we go through scan
//  abc defg
//  ^ ^ ^  ^
// skip whitespace

Token buildNumber() {
    while (isDigit(peek())) next();
    return fromEnum(T_NUMBER);
}

Token scanToken() {
    while(true) {
        if (isAtEnd()) return fromEnum(T_EOF);
        if (isWhitespace(peek())) trimWhitespace();

        //TODO: find the longest matching token
        // for now at least find digits and operators
        while (isDigit(peek())) return buildNumber();
    }
}
