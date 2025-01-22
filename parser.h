#include <wchar.h>
#include "scanner.h"

typedef struct
{
    const wchar_t *source;
    Token current;
} Parser;


Parser initParser(const wchar_t *source);
Token parse_expression();
struct ASTnode *parse_exp();
struct ASTnode *parse_stmt();
struct ASTnode *parse_decl();
