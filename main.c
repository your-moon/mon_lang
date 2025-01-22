#include "parser.h"
#include <locale.h>
#include <wchar.h>
int main(void)
{

    setlocale(LC_ALL, "");
    wchar_t source[] = L"танигч зарл (1 + 1)*2/3 123 \n фн оруулах хоосон хэрв\0";
    wprintf(L"CURRENT %lc\n", source[0]);
    initParser(source);
    Token token = parseExpression();
    wprintf(L"THIS IS TOKEN %d\n", token.type);
    wprintf(L"THIS IS Length %d\n\n", token.length);
    // lex something that means source to token list
    // for (;;)
    // {
    //     Token token = scanToken();
    //     wprintf(L"THIS IS TOKEN %d\n", token.type);
    //     wprintf(L"THIS IS Length %d\n\n", token.length);
    //     if (token.type == T_ERR)
    //     {
    //         return -1;
    //     }
    //     if (token.type == T_EOF)
    //     {
    //         return 0;
    //         // + 1 < 0 => false, return 2
    //         // * 3 < 2 => false, return 4
    //         // * 3 < 4 => true, skip
    //         // +
    //     }
    // }
    return 0;
}
