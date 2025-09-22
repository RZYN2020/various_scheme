#ifndef SCHEME_H
#define SCHEME_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Scheme 对象类型
enum {
    S_NUMBER,
    S_BOOL,
    S_SYMBOL,
    S_PAIR,
    S_NIL,
    S_PROC,
    S_CLOSURE
};

// Scheme 对象结构体
typedef struct S_Object {
    int type;
    union {
        double num_val;
        int bool_val;
        char *sym_val;
        struct {
            struct S_Object *car;
            struct S_Object *cdr;
        } pair;
        struct {
            struct S_Object *params;
            struct S_Object *body;
            struct S_Object *env;
        } closure;
        struct S_Object* (*prim_proc)(struct S_Object* args);
    } val;
    // 简单的引用计数（可选）
    int ref_count;
} S_Object;

// 环境
typedef struct S_Env {
    S_Object *bindings;
    struct S_Env *parent;
} S_Env;

// 全局环境
extern S_Env *global_env;

// 数据类型函数
S_Object *s_number(double num);
S_Object *s_bool(int b);
S_Object *s_symbol(const char *sym);
S_Object *s_pair(S_Object *car, S_Object *cdr);
S_Object *s_proc(S_Object* (*proc)(S_Object*));
S_Object *s_closure(S_Object *params, S_Object *body, S_Object *env);
S_Object *s_nil();

// 内存管理
void s_free(S_Object *obj);
void s_inc_ref(S_Object *obj);
void s_dec_ref(S_Object *obj);

// 解析器
S_Object *scheme_parse(const char *str);
S_Object *scheme_read(FILE *stream);

// 求值器
S_Object *scheme_eval(S_Object *expr, S_Env *env);

// 打印
void scheme_print(S_Object *obj);

// 环境操作
S_Env *s_env_new(S_Env *parent);
S_Object *s_env_find(S_Env *env, const char *sym);
void s_env_bind(S_Env *env, const char *sym, S_Object *val);

#endif // SCHEME_H