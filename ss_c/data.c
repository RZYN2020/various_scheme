#include "scheme.h"

// 简单的引用计数，避免频繁的malloc/free
void s_inc_ref(S_Object *obj) {
    if (obj) obj->ref_count++;
}

void s_dec_ref(S_Object *obj) {
    if (!obj) return;
    obj->ref_count--;
    if (obj->ref_count <= 0) {
        // 递归释放
        if (obj->type == S_SYMBOL && obj->val.sym_val) {
            free(obj->val.sym_val);
        } else if (obj->type == S_PAIR) {
            s_dec_ref(obj->val.pair.car);
            s_dec_ref(obj->val.pair.cdr);
        } else if (obj->type == S_CLOSURE) {
            s_dec_ref(obj->val.closure.params);
            s_dec_ref(obj->val.closure.body);
            // 环境不在这里释放，由其自身的生命周期管理
        }
        free(obj);
    }
}

// 创建 Scheme 对象
S_Object *s_number(double num) {
    S_Object *obj = malloc(sizeof(S_Object));
    obj->type = S_NUMBER;
    obj->val.num_val = num;
    obj->ref_count = 0;
    return obj;
}

S_Object *s_bool(int b) {
    S_Object *obj = malloc(sizeof(S_Object));
    obj->type = S_BOOL;
    obj->val.bool_val = b;
    obj->ref_count = 0;
    return obj;
}

S_Object *s_symbol(const char *sym) {
    S_Object *obj = malloc(sizeof(S_Object));
    obj->type = S_SYMBOL;
    obj->val.sym_val = strdup(sym);
    obj->ref_count = 0;
    return obj;
}

S_Object *s_pair(S_Object *car, S_Object *cdr) {
    S_Object *obj = malloc(sizeof(S_Object));
    obj->type = S_PAIR;
    obj->val.pair.car = car;
    obj->val.pair.cdr = cdr;
    s_inc_ref(car);
    s_inc_ref(cdr);
    obj->ref_count = 0;
    return obj;
}

S_Object *s_proc(S_Object* (*proc)(S_Object*)) {
    S_Object *obj = malloc(sizeof(S_Object));
    obj->type = S_PROC;
    obj->val.prim_proc = proc;
    obj->ref_count = 0;
    return obj;
}

S_Object *s_closure(S_Object *params, S_Object *body, S_Object *env) {
    S_Object *obj = malloc(sizeof(S_Object));
    obj->type = S_CLOSURE;
    obj->val.closure.params = params;
    obj->val.closure.body = body;
    obj->val.closure.env = env; // 不增加引用，环境由外部管理
    s_inc_ref(params);
    s_inc_ref(body);
    obj->ref_count = 0;
    return obj;
}

S_Object *s_nil() {
    static S_Object nil_obj = { .type = S_NIL };
    return &nil_obj;
}

// 环境管理
S_Env *s_env_new(S_Env *parent) {
    S_Env *env = malloc(sizeof(S_Env));
    env->bindings = s_nil();
    env->parent = parent;
    return env;
}

S_Object *s_env_find(S_Env *env, const char *sym) {
    S_Object *p = env->bindings;
    while (p->type != S_NIL) {
        S_Object *binding = p->val.pair.car;
        if (binding->type == S_PAIR &&
            binding->val.pair.car->type == S_SYMBOL &&
            strcmp(binding->val.pair.car->val.sym_val, sym) == 0) {
            return binding->val.pair.cdr;
        }
        p = p->val.pair.cdr;
    }
    if (env->parent) {
        return s_env_find(env->parent, sym);
    }
    fprintf(stderr, "Error: unbound variable '%s'\n", sym);
    exit(1);
}

void s_env_bind(S_Env *env, const char *sym, S_Object *val) {
    S_Object *p = env->bindings;
    while (p->type != S_NIL) {
        S_Object *binding = p->val.pair.car;
        if (binding->type == S_PAIR &&
            binding->val.pair.car->type == S_SYMBOL &&
            strcmp(binding->val.pair.car->val.sym_val, sym) == 0) {
            s_dec_ref(binding->val.pair.cdr);
            binding->val.pair.cdr = val;
            s_inc_ref(val);
            return;
        }
        p = p->val.pair.cdr;
    }
    S_Object *new_binding = s_pair(s_symbol(sym), val);
    env->bindings = s_pair(new_binding, env->bindings);
    s_inc_ref(new_binding);
    s_inc_ref(val);
}