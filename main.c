#include "scanner.h"
#include <stdio.h>
int main(void)
{

    char source[25] = "1+1 1233 3444-43\0";
    initScanner(source);
    // lex something that means source to token list

    for (;;)
    {
        Token token = scanToken();
        printf("THIS IS TOKEN %d\n", token.type);
        printf("THIS IS Length %d\n\n", token.length);
        if (token.type == T_EOF)
        {
            return -1;
        }
    }
    return 0;
}
