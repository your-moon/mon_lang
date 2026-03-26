#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <unistd.h>

// хэвлэ - print 64-bit integer
void khevle(long n) {
    printf("%ld", n);
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
