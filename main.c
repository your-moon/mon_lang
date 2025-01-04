#include "scanner.h"
#include <stdio.h>
int main(void)
{

    char source[10] = "123 456\0";
    initScanner(source);
    //lex something that means source to token list

    Token token = scanToken();
    printf("THIS IS TOKEN %d\n", token.type);

    Token stoken = scanToken();
    printf("THIS IS 2TOKEN %d\n", stoken.type);
    return 0;
}
