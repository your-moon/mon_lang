#include "scanner.h"
#include <locale.h>
#include <wchar.h>
int main(void)
{

    setlocale(LC_ALL, "");
    wchar_t source[] = L"танигч зарл 1 + 1 123 \n фн оруулах хоосон хэрв\0";
    wprintf(L"CURRENT %lc\n", source[0]);
    initScanner(source);
    // lex something that means source to token list
    for (;;)
    {
        Token token = scanToken();
        wprintf(L"THIS IS TOKEN %d\n", token.type);
        wprintf(L"THIS IS Length %d\n\n", token.length);
        if (token.type == T_ERR)
        {
            return -1;
        }
        if (token.type == T_EOF)
        {
            return 0;
        }
    }
    return 0;
}
