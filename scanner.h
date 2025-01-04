typedef enum {
    T_IDENT, T_CONST, T_INT, T_VOID, T_RETURN, T_LET, T_IF, T_EOF,T_NUMBER
} TokenType;

typedef struct {
    TokenType type;
    const char *start;
    int length;
    int line;
} Token;

typedef struct
{
    const char *start;
    const char *current;
    ;
    int line;
} Scanner;


void initScanner(const char *source);
Token scanToken();
