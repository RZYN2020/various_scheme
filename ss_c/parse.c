#include "scheme.h"
#include <ctype.h>

static int is_delimiter(int c) {
    return isspace(c) || c == '(' || c == ')' || c == EOF;
}

static S_Object *read_token(FILE *stream) {
    int c;
    while ((c = getc(stream)) != EOF) {
        if (!isspace(c)) {
            ungetc(c, stream);
            break;
        }
    }
    if (c == EOF) return NULL;
    
    c = getc(stream);
    if (c == '(') {
        // 列表
        S_Object *list = s_nil();
        S_Object **tail = &list;
        while ((c = getc(stream)) != EOF && c != ')') {
            ungetc(c, stream);
            S_Object *expr = scheme_read(stream);
            if (!expr) {
                fprintf(stderr, "Error: expected ')'\n");
                exit(1);
            }
            S_Object *pair = s_pair(expr, s_nil());
            *tail = pair;
            tail = &pair->val.pair.cdr;
        }
        return list;
    } else if (c == ')') {
        fprintf(stderr, "Error: unexpected ')'\n");
        exit(1);
    } else if (isdigit(c) || (c == '-' && isdigit(getc(stream)))) {
        ungetc(c, stream);
        double num;
        fscanf(stream, "%lf", &num);
        return s_number(num);
    } else if (c == '#') {
        c = getc(stream);
        if (c == 't') return s_bool(1);
        if (c == 'f') return s_bool(0);
        fprintf(stderr, "Error: invalid boolean literal\n");
        exit(1);
    } else {
        // 符号
        char buf[256];
        int i = 0;
        buf[i++] = c;
        while (!is_delimiter(c = getc(stream))) {
            buf[i++] = c;
            if (i >= sizeof(buf) - 1) break;
        }
        buf[i] = '\0';
        ungetc(c, stream);
        return s_symbol(buf);
    }
}

// 主读取函数
S_Object *scheme_read(FILE *stream) {
    return read_token(stream);
}