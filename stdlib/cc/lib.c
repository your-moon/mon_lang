#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <unistd.h>

// хэвлэ - print 64-bit integer
void khevle(long n) {
    printf("%ld", n);
}

// эхэвлэ - print unsigned 64-bit integer
void ekhevle(unsigned long n) {
    printf("%lu", n);
}

// тэмдэгтХэвлэх - print a Unicode codepoint as UTF-8
void temdegtKhevlekh(int cp) {
    unsigned c = (unsigned)cp;
    char buf[4];
    int n;
    if (c < 0x80) { buf[0] = c; n = 1; }
    else if (c < 0x800) { buf[0] = 0xC0 | (c >> 6); buf[1] = 0x80 | (c & 0x3F); n = 2; }
    else if (c < 0x10000) { buf[0] = 0xE0 | (c >> 12); buf[1] = 0x80 | ((c >> 6) & 0x3F); buf[2] = 0x80 | (c & 0x3F); n = 3; }
    else { buf[0] = 0xF0 | (c >> 18); buf[1] = 0x80 | ((c >> 12) & 0x3F); buf[2] = 0x80 | ((c >> 6) & 0x3F); buf[3] = 0x80 | (c & 0x3F); n = 4; }
    fwrite(buf, 1, n, stdout);
}

// мөр_хэвлэх - print string
void mqr_khevlekh(const char *s) {
    printf("%s", s);
}

// унш - read 64-bit integer
long unsh(void) {
    long n;
    scanf("%ld", &n);
    return n;
}

// унш32 - read 32-bit integer
int unsh32(void) {
    int n;
    scanf("%d", &n);
    return n;
}

// санамсаргүйТоо - random number (1 to n)
int sanamsargwyToo(int n) {
    static int seeded = 0;
    if (!seeded) {
        srand((unsigned int)time(NULL));
        seeded = 1;
    }
    return (rand() % n) + 1;
}

// одоо - current timestamp
long odoo(void) {
    return (long)time(NULL);
}

// чөлөөлөх - free heap memory
void chqlqqlqkh(void *p) {
    free(p);
}

// хүлээх - sleep milliseconds
void khwleekh(int ms) {
    usleep(ms * 1000);
}

// дэлгэцЦэвэрлэх - clear screen (ANSI escape)
void delgetsTseverlekh(void) {
    printf("\033[H\033[2J");
    fflush(stdout);
}
