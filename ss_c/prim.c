#include "scheme.h"

static void check_arg_count(S_Object *args, int min, int max) {
    int count = 0;
    S_Object *p = args;
    while (p->type != S_NIL) {
        count++;
        p = p->val.pair.cdr;
    }
    if (count < min || (max != -1 && count > max)) {
        fprintf(stderr, "Error: incorrect number of arguments\n");
        exit(1);
    }
}

S_Object *prim_add(S_Object *args) {
    double sum = 0.0;
    S_Object *p = args;
    while (p->type != S_NIL) {
        S_Object *arg = p->val.pair.car;
        if (arg->type != S_NUMBER) {
            fprintf(stderr, "Error: '+' requires numbers\n");
            exit(1);
        }
        sum += arg->val.num_val;
        p = p->val.pair.cdr;
    }
    return s_number(sum);
}

S_Object *prim_sub(S_Object *args) {
    check_arg_count(args, 1, 2);
    S_Object *first = args->val.pair.car;
    if (first->type != S_NUMBER) {
        fprintf(stderr, "Error: '-' requires numbers\n");
        exit(1);
    }
    if (args->val.pair.cdr->type == S_NIL) { // Unary minus
        return s_number(-first->val.num_val);
    }
    S_Object *second = args->val.pair.cdr->val.pair.car;
    if (second->type != S_NUMBER) {
        fprintf(stderr, "Error: '-' requires numbers\n");
        exit(1);
    }
    return s_number(first->val.num_val - second->val.num_val);
}

S_Object *prim_mul(S_Object *args) {
    double product = 1.0;
    S_Object *p = args;
    while (p->type != S_NIL) {
        S_Object *arg = p->val.pair.car;
        if (arg->type != S_NUMBER) {
            fprintf(stderr, "Error: '*' requires numbers\n");
            exit(1);
        }
        product *= arg->val.num_val;
        p = p->val.pair.cdr;
    }
    return s_number(product);
}

S_Object *prim_div(S_Object *args) {
    check_arg_count(args, 1, 2);
    S_Object *first = args->val.pair.car;
    if (first->type != S_NUMBER) {
        fprintf(stderr, "Error: '/' requires numbers\n");
        exit(1);
    }
    if (args->val.pair.cdr->type == S_NIL) { // Unary division (reciprocal)
        if (first->val.num_val == 0.0) {
            fprintf(stderr, "Error: division by zero\n");
            exit(1);
        }
        return s_number(1.0 / first->val.num_val);
    }
    S_Object *second = args->val.pair.cdr->val.pair.car;
    if (second->type != S_NUMBER) {
        fprintf(stderr, "Error: '/' requires numbers\n");
        exit(1);
    }
    if (second->val.num_val == 0.0) {
        fprintf(stderr, "Error: division by zero\n");
        exit(1);
    }
    return s_number(first->val.num_val / second->val.num_val);
}

S_Object *prim_eq(S_Object *args) {
    check_arg_count(args, 2, 2);
    S_Object *first = args->val.pair.car;
    S_Object *second = args->val.pair.cdr->val.pair.car;
    if (first->type != second->type) {
        return s_bool(0);
    }
    if (first->type == S_NUMBER) {
        return s_bool(first->val.num_val == second->val.num_val);
    } else if (first->type == S_BOOL) {
        return s_bool(first->val.bool_val == second->val.bool_val);
    } else {
        return s_bool(0);
    }
}

S_Object *prim_lt(S_Object *args) {
    check_arg_count(args, 2, 2);
    S_Object *first = args->val.pair.car;
    S_Object *second = args->val.pair.cdr->val.pair.car;
    if (first->type != S_NUMBER || second->type != S_NUMBER) {
        fprintf(stderr, "Error: '<' requires numbers\n");
        exit(1);
    }
    return s_bool(first->val.num_val < second->val.num_val);
}

S_Object *prim_gt(S_Object *args) {
    check_arg_count(args, 2, 2);
    S_Object *first = args->val.pair.car;
    S_Object *second = args->val.pair.cdr->val.pair.car;
    if (first->type != S_NUMBER || second->type != S_NUMBER) {
        fprintf(stderr, "Error: '>' requires numbers\n");
        exit(1);
    }
    return s_bool(first->val.num_val > second->val.num_val);
}

S_Object *prim_not(S_Object *args) {
    check_arg_count(args, 1, 1);
    S_Object *arg = args->val.pair.car;
    if (arg->type != S_BOOL) {
        fprintf(stderr, "Error: 'not' requires a boolean\n");
        exit(1);
    }
    return s_bool(!arg->val.bool_val);
}

void init_primitives(S_Env *env) {
    s_env_bind(env, "+", s_proc(prim_add));
    s_env_bind(env, "-", s_proc(prim_sub));
    s_env_bind(env, "*", s_proc(prim_mul));
    s_env_bind(env, "/", s_proc(prim_div));
    s_env_bind(env, "=", s_proc(prim_eq));
    s_env_bind(env, "<", s_proc(prim_lt));
    s_env_bind(env, ">", s_proc(prim_gt));
    s_env_bind(env, "not", s_proc(prim_not));
}