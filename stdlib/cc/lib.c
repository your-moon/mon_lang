#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <unistd.h>
#include <string.h>
#include <fcntl.h>
#include <stdint.h>

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

// мөрУрт - string length in bytes
long mqrUrt(const char *s) { return (long)strlen(s); }

// байт - byte at index
int bayt(const char *s, long i) { return (unsigned char)s[i]; }

// байтТавих - store byte at index
void baytTavikh(char *s, long i, int b) { s[i] = (char)b; }

// файлУншихБүтэн - read a whole file into a NUL-terminated buffer
char *faylUnshikhBwten(const char *path) {
    int fd = open(path, O_RDONLY);
    if (fd < 0) return "";
    long size = lseek(fd, 0, SEEK_END);
    lseek(fd, 0, SEEK_SET);
    char *buf = malloc(size + 1);
    long n = read(fd, buf, size);
    if (n < 0) n = 0;
    buf[n] = 0;
    close(fd);
    return buf;
}

// файлБичих - write a string to a file, returns 0 on success
int faylBichikh(const char *path, const char *content) {
    int fd = open(path, O_WRONLY | O_CREAT | O_TRUNC, 0644);
    if (fd < 0) return 1;
    long len = (long)strlen(content);
    long n = write(fd, content, len);
    close(fd);
    return n == len ? 0 : 1;
}

// аргументТоо / аргумент - argc/argv access
static int mon_argc;
static char **mon_argv;
__attribute__((constructor)) static void mon_capture_args(int argc, char **argv) {
    mon_argc = argc;
    mon_argv = argv;
}
int argumyentToo(void) { return mon_argc; }
char *argumyent(int i) { return (i >= 0 && i < mon_argc) ? mon_argv[i] : ""; }

// мөрШинэ - allocate a mutable string buffer (zeroed, NUL-safe)
char *mqrShine(long len) {
    char *buf = malloc(len + 1);
    memset(buf, 0, len + 1);
    return buf;
}
