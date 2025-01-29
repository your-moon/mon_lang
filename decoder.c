#include "decoder.h"
//leading bits determines the how long is byte is
//that means
//0xxxxxxx -> 1 byte
//110xxxxx -> 2 byte
//1110xxxx -> 2 byte
//11110xxx -> 2 byte
int decode_utf8(const char *p, uint32_t *rune) {
    unsigned char c = p[0];

    if (c < 0x80) {  // 1-byte ASCII
        *rune = c;
        return 1;
    } else if ((c & 0xE0) == 0xC0) {  // 2-byte UTF-8
        *rune = ((p[0] & 0x1F) << 6) | (p[1] & 0x3F);
        return 2;
    } else if ((c & 0xF0) == 0xE0) {  // 3-byte UTF-8
        *rune = ((p[0] & 0x0F) << 12) | ((p[1] & 0x3F) << 6) | (p[2] & 0x3F);
        return 3;
    } else if ((c & 0xF8) == 0xF0) {  // 4-byte UTF-8
        *rune = ((p[0] & 0x07) << 18) | ((p[1] & 0x3F) << 12) | ((p[2] & 0x3F) << 6) | (p[3] & 0x3F);
        return 4;
    }

    *rune = 0xFFFD;  // Invalid byte sequence
    return 1;
}
