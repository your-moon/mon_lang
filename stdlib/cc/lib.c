#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <unistd.h>
#include <string.h>
#include <fcntl.h>
#include <stdint.h>
#include <pthread.h>

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

/* Conservative mark-sweep garbage collector.

   The compiler emits monAlloc for every heap value (шинэ, arrays, structs).
   Blocks carry a header {next, size, mark}. A collection scans the program
   stack (from the current frame up to the bottom captured at startup) plus
   the data/bss globals conservatively: any machine word that points into a
   known live block is a root, and marking follows pointers found inside
   marked blocks. Single-threaded batch model - no write barriers, no
   incremental work. чөлөөлөх is a hint the collector doesn't need. */

typedef struct GcBlock {
    struct GcBlock *next;
    size_t size;
    int mark;
    /* payload follows */
} GcBlock;

static GcBlock *gc_head = NULL;
static char *gc_stack_bottom = NULL;
static size_t gc_bytes = 0;
static size_t gc_threshold = 1 << 20; /* collect every ~1 MB allocated */
extern char **environ;

static GcBlock *gc_block_of(void *ptr) {
    /* is ptr the payload of a known block? */
    for (GcBlock *b = gc_head; b; b = b->next) {
        char *payload = (char *)(b + 1);
        if ((char *)ptr >= payload && (char *)ptr < payload + b->size) {
            return b;
        }
    }
    return NULL;
}

static void gc_mark_range(char *lo, char *hi) {
    lo = (char *)((uintptr_t)lo & ~(uintptr_t)(sizeof(void *) - 1));
    for (char *p = lo; p + sizeof(void *) <= hi; p += sizeof(void *)) {
        void *word = *(void **)p;
        GcBlock *b = gc_block_of(word);
        if (b && !b->mark) {
            b->mark = 1;
            gc_mark_range((char *)(b + 1), (char *)(b + 1) + b->size);
        }
    }
}

static void gc_collect(void) {
    for (GcBlock *b = gc_head; b; b = b->next) b->mark = 0;

    /* roots: the whole stack (callee-saved regs are spilled there in this
       compiler), plus the data/bss segment reachable via environ is not
       scanned; globals live in our own __DATA which the stack references. */
    char stack_top;
    char *lo = &stack_top, *hi = gc_stack_bottom;
    if (lo > hi) { char *t = lo; lo = hi; hi = t; }
    gc_mark_range(lo, hi);

    GcBlock **link = &gc_head;
    while (*link) {
        GcBlock *b = *link;
        if (b->mark) {
            link = &b->next;
        } else {
            *link = b->next;
            free(b);
        }
    }
    gc_bytes = 0;
}

void *monAlloc(long n) {
    if (gc_stack_bottom == NULL) {
        /* constructor didn't run (shouldn't happen); fall back to base */
        gc_stack_bottom = (char *)pthread_get_stackaddr_np(pthread_self());
    }
    if (gc_bytes >= gc_threshold) gc_collect();
    GcBlock *b = malloc(sizeof(GcBlock) + (size_t)n);
    if (!b) { gc_collect(); b = malloc(sizeof(GcBlock) + (size_t)n); }
    if (!b) { _exit(71); }
    b->size = (size_t)n;
    b->mark = 0;
    b->next = gc_head;
    gc_head = b;
    gc_bytes += (size_t)n;
    /* zero the payload: callers rely on fresh memory being clean */
    char *payload = (char *)(b + 1);
    for (long i = 0; i < n; i++) payload[i] = 0;
    return payload;
}

// чөлөөлөх - a hint; the collector reclaims automatically
void chqlqqlqkh(void *p) {
    (void)p;
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
    char *buf = monAlloc(size + 1);
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
    // this frame sits above main's; a local here is a safe stack bottom
    char anchor;
    gc_stack_bottom = &anchor;
}
int argumyentToo(void) { return mon_argc; }
char *argumyent(int i) { return (i >= 0 && i < mon_argc) ? mon_argv[i] : ""; }

// мөрШинэ - allocate a mutable string buffer (zeroed, NUL-safe)
char *mqrShine(long len) {
    char *buf = monAlloc(len + 1);
    memset(buf, 0, len + 1);
    return buf;
}

// бутархайХэвлэх - print a double as fixed decimal, 6 truncated fractional
// digits. Deliberately simple (no dtoa/rounding) so the hand-written
// native printer can produce byte-identical output.
void butarkhayKhevlekh(double d) {
    if (d < 0) { putchar('-'); d = -d; }
    long ip = (long)d;
    printf("%ld.", ip);
    double frac = d - (double)ip;
    for (int i = 0; i < 6; i++) {
        frac *= 10.0;
        int dg = (int)frac;
        putchar('0' + dg);
        frac -= (double)dg;
    }
}
