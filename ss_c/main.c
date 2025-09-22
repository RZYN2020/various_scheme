#include "scheme.h"

// 全局环境
S_Env *global_env;

// 前向声明
void init_primitives(S_Env *env);

void repl() {
    printf("Simple Scheme REPL (press Ctrl+D to exit)\n");
    while (1) {
        printf("> ");
        S_Object *expr = scheme_read(stdin);
        if (!expr) break;
        
        S_Object *result = scheme_eval(expr, global_env);
        scheme_print(result);
        printf("\n");
        
        s_dec_ref(expr);
        s_dec_ref(result);
    }
}

void process_file(const char *filename) {
    FILE *fp = fopen(filename, "r");
    if (!fp) {
        perror("fopen");
        exit(1);
    }
    
    S_Object *expr;
    while ((expr = scheme_read(fp)) != NULL) {
        S_Object *result = scheme_eval(expr, global_env);
        scheme_print(result);
        printf("\n");
        s_dec_ref(expr);
        s_dec_ref(result);
    }
    
    fclose(fp);
}

void scheme_print(S_Object *obj) {
    if (!obj) return;
    
    switch (obj->type) {
        case S_NUMBER:
            printf("%g", obj->val.num_val);
            break;
        case S_BOOL:
            printf("#%c", obj->val.bool_val ? 't' : 'f');
            break;
        case S_SYMBOL:
            printf("%s", obj->val.sym_val);
            break;
        case S_PAIR:
            printf("(");
            scheme_print(obj->val.pair.car);
            S_Object *cdr = obj->val.pair.cdr;
            while (cdr->type == S_PAIR) {
                printf(" ");
                scheme_print(cdr->val.pair.car);
                cdr = cdr->val.pair.cdr;
            }
            if (cdr->type != S_NIL) {
                printf(" . ");
                scheme_print(cdr);
            }
            printf(")");
            break;
        case S_NIL:
            printf("()");
            break;
        case S_PROC:
            printf("#<procedure>");
            break;
        case S_CLOSURE:
            printf("#<closure>");
            break;
    }
}

int main(int argc, char **argv) {
    global_env = s_env_new(NULL);
    init_primitives(global_env);
    
    if (argc == 1) {
        repl();
    } else if (argc == 2) {
        process_file(argv[1]);
    } else {
        fprintf(stderr, "Usage: %s [file.ss]\n", argv[0]);
        exit(1);
    }
    
    // 简单的清理
    // s_dec_ref(global_env->bindings);
    free(global_env);
    
    return 0;
}