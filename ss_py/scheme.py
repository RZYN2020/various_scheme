import sys
import re
from typing import List, Union, Dict, Any, Callable, Optional

SchemeSymbol = str
SchemeProc = Callable[..., Any]
SchemeValue = Union[float, bool, SchemeProc]
Expression = Union[float, bool, SchemeSymbol, List['Expression']]

class Env:
    def __init__(self, bindings: Dict[str, SchemeValue], parent: Optional['Env'] = None):
        self.bindings = bindings
        self.parent = parent

    def find(self, var: SchemeSymbol) -> Dict[str, SchemeValue]:
        if var in self.bindings:
            return self.bindings
        if self.parent:
            return self.parent.find(var)
        raise NameError(f"Variable '{var}' not found")

    def __setitem__(self, key: SchemeSymbol, value: SchemeValue) -> None:
        self.bindings[key] = value

    def __getitem__(self, key: SchemeSymbol) -> SchemeValue:
        return self.find(key)[key]

    def __contains__(self, key: SchemeSymbol) -> bool:
        try:
            self.find(key)
            return True
        except NameError:
            return False

def tokenize(s: str) -> List[str]:
    return re.sub(r'([()])', r' \1 ', s).split()

def parse(tokens: List[str]) -> 'Expression':
    if not tokens:
        raise SyntaxError("Unexpected EOF while parsing")
    
    token = tokens.pop(0)
    if token == '(':
        L: List['Expression'] = []
        while tokens and tokens[0] != ')':
            L.append(parse(tokens))
        
        if not tokens or tokens[0] != ')':
            raise SyntaxError("Expected ')'")
        tokens.pop(0)
        return L
    
    if token == ')':
        raise SyntaxError("Unexpected ')'")
    
    return atom(token)


def atom(token: str) -> Union[float, bool, SchemeSymbol]:
    try:
        return float(token)
    except ValueError:
        if token == '#t':
            return True
        if token == '#f':
            return False
        return token

def _define(expr: List[Expression], env: Env) -> None:
    if len(expr) != 3 or not isinstance(expr[1], str):
        raise SyntaxError("Malformed define expression")
    _, var, value = expr
    if not isinstance(var, SchemeSymbol):
        raise SyntaxError("First argument to define must be a symbol")
    env[var] = evaluate(value, env)
    return None

def _if(expr: List[Expression], env: Env) -> SchemeValue:
    if len(expr) != 4:
        raise SyntaxError("Malformed if expression")
    _, test, consequent, alternative = expr
    if evaluate(test, env) is not False:
        return evaluate(consequent, env)
    return evaluate(alternative, env)

def _lambda(expr: List[Expression], env: Env) -> SchemeProc:
    if len(expr) != 3 or not isinstance(expr[1], list) or not all(isinstance(p, str) for p in expr[1]):
        raise SyntaxError("Malformed lambda expression")
    _, params, body = expr
    if not isinstance(params, list):
        raise SyntaxError("Lambda parameters must be a list")
    if not all(isinstance(p, SchemeSymbol) for p in params):
        raise SyntaxError("Lambda parameters must be symbols")
    return lambda *args: evaluate(body, Env(dict(zip(params, args)), env)) #type: ignore

def _and(expr: List[Expression], env: Env) -> bool:
    if len(expr) == 1:
        return True 
    
    val: Any = True
    for arg in expr[1:]:
        val = evaluate(arg, env)
        if val is False:
            return False
    return val

def _or(expr: List[Any], env: Env) -> bool:
    if len(expr) == 1:
        return False
        
    val: Any = False
    for arg in expr[1:]:
        val = evaluate(arg, env)
        if val is not False:
            return val
    return val

special_forms: Dict[str, Callable[[List[Expression], Env], SchemeValue]] = {
    'if': _if,
    'lambda': _lambda,
    'and': _and,
    'or': _or,
}

def evaluate(expr: Expression, env: Env) -> SchemeValue:    
    if isinstance(expr, (float, bool)):
        return expr
    if isinstance(expr, str):
        return env.find(expr)[expr]
    
    if not isinstance(expr, list) or not expr:
        raise SyntaxError(f"Malformed expression: {expr}")

    form = expr[0]

    if isinstance(form, str) and form == 'define':
         return _define(expr, env)  # type: ignore
    if isinstance(form, str) and form in special_forms:
         return special_forms[form](expr, env)
    else:
        proc = evaluate(form, env)
        if not callable(proc):
            raise TypeError(f"Procedure is not callable: {form}")
        args = [evaluate(arg, env) for arg in expr[1:]]
        return proc(*args)

def create_global_env() -> Env:
    return Env({
        '+': lambda *args: sum(args),
        '-': lambda x, *ys: x - sum(ys) if ys else -x,
        '*': lambda *args: (args[0] if len(args) == 1 else args[0] * args[1]),
        '/': lambda x, y: x / y,
        '=': lambda x, y: x == y,
        '<=': lambda x, y: x <= y,
        '>=': lambda x, y: x >= y,
        '<': lambda x, y: x < y,
        '>': lambda x, y: x > y,
        'not': lambda x: not x,
    })

def sprint(value: SchemeValue) -> None:
    if isinstance(value, bool):
        print("#t" if value else "#f")
    elif isinstance(value, float):
        if value.is_integer():
            print(f"{int(value)}.0")
        else:
            print(value)
    else:
        print(value)


def main() -> None:
    global_env = create_global_env()

    if len(sys.argv) > 1:
        file_mode(sys.argv[1], global_env)
    else:
        repl_mode(global_env)

def file_mode(filename: str, env: Env) -> None:
    try:
        with open(filename, 'r', encoding='utf-8') as f:
            source_buffer = ""
            for line in f:
                stripped_line = line.strip()
                if not stripped_line or stripped_line.startswith(';'):
                    continue
                source_buffer += " " + stripped_line

                open_parens = source_buffer.count('(')
                close_parens = source_buffer.count(')')

                if (open_parens > 0 and open_parens == close_parens) or (open_parens == 0 and close_parens == 0 and not source_buffer.isspace()):
                    try:
                        result = evaluate(parse(tokenize(source_buffer)), env)
                        if result is not None:
                            sprint(result)
                    except (SyntaxError, NameError, TypeError, IndexError) as e:
                        print(f"Error: {e}")
                    finally:
                        source_buffer = ""

    except FileNotFoundError:
        print(f"Error: File '{filename}' not found.")

def repl_mode(env: Env) -> None:
    while True:
        try:
            expr = input("scheme> ")
            if not expr.strip():
                continue
            if expr.strip() == "exit":
                break
            
            result = evaluate(parse(tokenize(expr)), env)
            if result is not None:
                sprint(result)
        except (SyntaxError, NameError, TypeError, IndexError) as e:
            print(f"Error: {e}")
        except EOFError:
            print("\nExiting...")
            break

if __name__ == "__main__":
    main()