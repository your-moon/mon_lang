#include "parser.h"
#include "defs.h"
#include <locale.h>
#include <wchar.h>
#include "tree.h"
int main(void)
{

    setlocale(LC_ALL, "");
    wchar_t source[] = L"фн майн()   -> тоо {буц 1;}\0";
    // wchar_t source[] = L"-1";
    // wprintf(L"CURRENT %lc\n", source[0]);

    initScanner(source);
    initParser(source);
    // struct ASTnode *ast = parse_decl();
    // pretty_print_ast(ast);

    // lex something that means source to token list
    for (;;)
    {
        Token token = scanToken();
        printToken(token);
        // wprintf(L"THIS IS Length %d\n", token.length);
        // wprintf(L"THIS IS VALUE %ls\n\n", token.value);
        if (token.type == T_ERR)
        {
            return -1;
        }
        if (token.type == T_EOF)
        {
            return 0;
            // + 1 < 0 => false, return 2
            // * 3 < 2 => false, return 4
            // * 3 < 4 => true, skip
            // +
        }
    }
    return 0;
}
