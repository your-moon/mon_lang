#include <wchar.h>
#include "scanner.h"

typedef struct
{
    const wchar_t *source;
    Token current;
} Parser;


Parser initParser(const wchar_t *source);
Token parseExpression();
