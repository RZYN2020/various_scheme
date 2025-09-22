#include "scheme.h"

// 辅助函数：将列表转换为C数组
static S_Object **list_to_array(S_Object *list, int *len) {
    int count = 0;
    S_Object *p = list;
    while (p->type != S_NIL) {
        count++;
        p = p->val.pair.cdr;
    }
    *len = count;
    S_Object **arr = malloc(count * sizeof(S_Object*));
    p = list;
    int i = 0;
    while (p->type != S_NIL) {
        arr[i++] = p->val.pair.car;
        p = p->val.pair.cdr;
    }
    return arr;
}

S_Object *scheme_eval(S_Object *expr, S_Env *env) {
    if (!expr) return s_nil();
    
    switch (expr->type) {
        case S_NUMBER:
        case S_BOOL:
        case S_PROC:
        case S_CLOSURE:
            return expr;
            
        case S_SYMBOL:
            return s_env_find(env, expr->val.sym_val);
            
        case S_PAIR: {
            S_Object *proc = expr->val.pair.car;
            S_Object *args = expr->val.pair.cdr;
            
            if (proc->type == S_SYMBOL) {
                if (strcmp(proc->val.sym_val, "if") == 0) {
                    S_Object *test = args->val.pair.car;
                    S_Object *conseq = args->val.pair.cdr->val.pair.car;
                    S_Object *alt = args->val.pair.cdr->val.pair.cdr->val.pair.car;
                    
                    if (scheme_eval(test, env)->val.bool_val != 0) {
                        return scheme_eval(conseq, env);
                    } else {
                        return scheme_eval(alt, env);
                    }
                }
                if (strcmp(proc->val.sym_val, "define") == 0) {
                    S_Object *sym = args->val.pair.car;
                    S_Object *val_expr = args->val.pair.cdr->val.pair.car;
                    S_Object *val = scheme_eval(val_expr, env);
                    s_env_bind(env, sym->val.sym_val, val);
                    return s_nil();
                }
                if (strcmp(proc->val.sym_val, "lambda") == 0) {
                    S_Object *params = args->val.pair.car;
                    S_Object *body = args->val.pair.cdr->val.pair.car;
                    return s_closure(params, body, env);
                }
                if (strcmp(proc->val.sym_val, "and") == 0) {
                    S_Object *p = args;
                    while (p->type != S_NIL) {
                        S_Object *res = scheme_eval(p->val.pair.car, env);
                        if (res->type == S_BOOL && res->val.bool_val == 0) {
                            return s_bool(0);
                        }
                        p = p->val.pair.cdr;
                    }
                    return s_bool(1);
                }
                if (strcmp(proc->val.sym_val, "or") == 0) {
                    S_Object *p = args;
                    while (p->type != S_NIL) {
                        S_Object *res = scheme_eval(p->val.pair.car, env);
                        if (res->type == S_BOOL && res->val.bool_val != 0) {
                            return s_bool(1);
                        }
                        p = p->val.pair.cdr;
                    }
                    return s_bool(0);
                }
            }

            // 函数应用
            S_Object *proc_obj = scheme_eval(proc, env);
            if (proc_obj->type != S_PROC && proc_obj->type != S_CLOSURE) {
                fprintf(stderr, "Error: not a procedure\n");
                exit(1);
            }

            S_Object *evaled_args = s_nil();
            S_Object **tail = &evaled_args;
            S_Object *p = args;
            while (p->type != S_NIL) {
                S_Object *arg = scheme_eval(p->val.pair.car, env);
                S_Object *pair = s_pair(arg, s_nil());
                *tail = pair;
                tail = &pair->val.pair.cdr;
                s_dec_ref(arg);
                p = p->val.pair.cdr;
            }
            
            S_Object *result;
            if (proc_obj->type == S_PROC) {
                result = proc_obj->val.prim_proc(evaled_args);
            } else { // S_CLOSURE
                S_Env *new_env = s_env_new(proc_obj->val.closure.env);
                
                S_Object *p_params = proc_obj->val.closure.params;
                S_Object *p_args = evaled_args;
                
                while (p_params->type != S_NIL) {
                    S_Object *param_sym = p_params->val.pair.car;
                    S_Object *arg_val = p_args->val.pair.car;
                    s_env_bind(new_env, param_sym->val.sym_val, arg_val);
                    p_params = p_params->val.pair.cdr;
                    p_args = p_args->val.pair.cdr;
                }
                result = scheme_eval(proc_obj->val.closure.body, new_env);
            }
            s_dec_ref(evaled_args);
            return result;
        }
        default:
            fprintf(stderr, "Error: invalid expression\n");
            exit(1);
    }
}