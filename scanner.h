#include <wchar.h>
typedef enum {
    T_IDENT, T_CONST, T_INT, T_VOID, T_RETURN, T_LET, T_IF, T_EOF,T_NUMBER,
    T_IMPORT,
    T_PLUS, T_MINUS, T_DIV, T_MUL,
    T_OPEN_PAREN, T_CLOSE_PAREN,
    T_OPEN_BRACE, T_CLOSE_BRACE,
    T_SEMICOLON,
    T_FN,
    T_ERR,
} TokenType;

typedef struct {
    TokenType type;
    const wchar_t *start;
    int length;
    int line;
} Token;

typedef struct
{
    const wchar_t *start;
    const wchar_t *current;
    int line;
} Scanner;


void initScanner(const wchar_t *source);
Token scanToken();
