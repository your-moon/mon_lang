#include <wchar.h>
typedef enum
{
    T_IDENT,
    T_TRUE,
    T_FALSE,
    T_STRING,
    T_NUMBER,

    T_BINARY_OP,
    T_CONST,

    T_INT_TYPE,
    T_VOID_TYPE,

    T_RETURN,
    T_LET,
    T_IMPORT,
    T_IF,
    T_ELSE,

    T_EQUAL,
    T_COMMA,
    T_OPEN_PAREN,
    T_CLOSE_PAREN,

    T_OPEN_BRACE,
    T_CLOSE_BRACE,

    T_SEMICOLON,
    T_RIGHTARROW,

    T_FN,
    T_STRUCT,

    T_ERR,
    T_EOF,
    T_SOF,
} TokenType;

typedef struct
{
    TokenType type;
    const wchar_t *start;
    const wchar_t *value;
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
static Token fromEnum(TokenType type);
Token scanToken();
