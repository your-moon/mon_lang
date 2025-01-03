#include <stdio.h>
FILE *outfile;

static char *reglist[4] = {"%r8", "%r9", "%r10", "%r11"};

static int
cgenadd(int r1, int r2)
{
    fprintf(outfile, "\taddq\t%s, %s\n", reglist[r2], reglist[r1]);
}