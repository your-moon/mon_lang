#define DEBUG_SCANNER 0
#ifdef DEBUG_SCANNER
#include <stdio.h>
#endif
#include <stdbool.h>

typedef struct
{
    const char *start;
    const char *current;
    ;
    int line;
} Scanner;

Scanner scanner;

void initScanner(const char *source)
{
    if (DEBUG_SCANNER)
        printf("Initializing scanner\n");
    scanner.start = source;
    scanner.current = source;
    scanner.line = 1;
}

static bool isAtEnd()
{
    return *scanner.current == '\0';
}
