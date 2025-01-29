#include "parser.h"
#include "defs.h"
#include <stdint.h>
#include <wchar.h>
#include "tree.h"
#include "decoder.h"

#define MAX_RUNES 256

int main() {
    const char source[] = "А Б В Г Д Е Ё Ж З И Й К Л М Н О Ө П Р С Т У Ф Х Ц Ч Ш Щ Ь Ы Э Ю Я а б в г д е ё ж з и й к л м н о ө п р с т у ф х ц ч ш щ ь ы э ю я";
    for (int i = 0; source[i] != '\0'; i++) {
        printf("Character: '%c', Hex: 0x%02X\n", source[i], (unsigned char)source[i]);
    }

    const char *p = source;
    uint32_t runes[MAX_RUNES];
    int rune_count = 0;

    while (*p && rune_count < MAX_RUNES) {
        uint32_t rune;
        int width = decode_utf8(p, &runes[rune_count]);
        printf("Rune: U+%04X, Width: %d bytes\n", runes[rune_count], width);
        if (runes[rune_count] == L'А') {
            printf("MATCH FOUND\n");
        }
        p += width;  // Move forward by decoded rune width
        rune_count++;
    }

    return 0;
}

// int main(void)
// {
//
//     // setlocale(LC_ALL, "");
//     char source[] = "фн майн()   -> тоо {буц 1;}\0";
//
//     for (int i = 0; source[i] != '\0'; i++) {
//         printf("%02X ", (unsigned char)source[i]);
//     }
//     // wchar_t source[] = L"-1";
//     // wprintf(L"CURRENT %lc\n", source[0]);
//
//     // initScanner(source);
//     // initParser(source);
//     // // struct ASTnode *ast = parse_decl();
//     // // pretty_print_ast(ast);
//     //
//     // // lex something that means source to token list
//     // for (;;)
//     // {
//     //     Token token = scanToken();
//     //     printToken(token);
//     //     // wprintf(L"THIS IS Length %d\n", token.length);
//     //     // wprintf(L"THIS IS VALUE %ls\n\n", token.value);
//     //     if (token.type == T_ERR)
//     //     {
//     //         return -1;
//     //     }
//     //     if (token.type == T_EOF)
//     //     {
//     //         return 0;
//     //         // + 1 < 0 => false, return 2
//     //         // * 3 < 2 => false, return 4
//     //         // * 3 < 4 => true, skip
//     //         // +
//     //     }
//     // }
//
//     // return 0;
// }
