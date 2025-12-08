package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type Env struct {
	parent   *Env
	bindings map[string]Val
}

func NewEnv(parent *Env) *Env {
	return &Env{parent, make(map[string]Val)}
}

func (env *Env) Define(name string, value Val) {
	env.bindings[name] = value
}

func (env *Env) Resolve(name string) Val {
	val, ok := env.bindings[name]
	if ok {
		return val
	}
	if env.parent != nil {
		return env.parent.Resolve(name)
	}
	return nil
}

type Val interface {
	repr() string
}

type SNumber float64
type SBool bool
type SLambda func([]Val, *Env) (Val, error)

func (n SNumber) repr() string {
	return fmt.Sprintf("%f", n)
}

func (b SBool) repr() string {
	return fmt.Sprintf("%t", b)
}

func (s SLambda) repr() string {
	return "<#lambda>"
}

type Expr interface {
	evaluate(env *Env) (Val, error)
}

func (n SNumber) evaluate(env *Env) (Val, error) {
	return n, nil
}

func (b SBool) evaluate(env *Env) (Val, error) {
	return b, nil
}

type SSym string

func (n SSym) evaluate(env *Env) (Val, error) {
	return env.Resolve(string(n)), nil
}

type SList []Expr

func (n SList) evaluate(env *Env) (Val, error) {
	if len(n) == 0 {
		return nil, errors.New("empty list")
	}

	fun_sym, ok := n[0].(SSym)
	var fun SLambda
	if ok {
		if string(fun_sym) == "define" {
			if len(n) != 3 {
				return nil, errors.New("invalid list")
			}
			sym, ok := n[1].(SSym)
			if !ok {
				return nil, errors.New("invalid list")
			}
			val, err := n[2].evaluate(env)
			if err != nil {
				return nil, err
			}
			env.Define(string(sym), val)
			return val, nil
		}
		fun_special, ok := SpecialForms[string(fun_sym)]
		if ok {
			return fun_special(n[1:], env)
		} else {
			fun, ok = env.Resolve(string((fun_sym))).(SLambda)
			if fun == nil {
				return nil, errors.New("unknown function")
			}
			if !ok {
				return nil, errors.New("unknown function")
			}
		}
	} else {
		funV, err := n[0].evaluate(env)
		if err != nil {
			return nil, err
		}
		fun, ok = funV.(SLambda)
		if !ok {
			return nil, errors.New("first element is not a function")
		}
	}

	args := make([]Val, len(n)-1)
	for i, e := range n[1:] {
		val, err := e.evaluate(env)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}

	return fun(args, env)
}

func readExpr(code string) (Expr, string, error) {
	code = strings.TrimSpace(code)
	if len(code) == 0 {
		return nil, "", errors.New("unexpected EOF")
	}

	first := code[0]
	if first == '(' {
		remain := code[1:]
		list := SList{}

		for {
			remain = strings.TrimSpace(remain)
			if len(remain) == 0 {
				return nil, "", errors.New("unbalanced parenthesis: missing ')'")
			}

			if remain[0] == ')' {
				return list, remain[1:], nil
			}

			expr, remain_next, err := readExpr(remain)
			if err != nil {
				return nil, "", err
			}
			list = append(list, expr)
			remain = remain_next
		}
	}

	nextParamsIdx := strings.IndexFunc(code, func(r rune) bool {
		return unicode.IsSpace(r) || r == ')' || r == '('
	})

	var token string
	var nextRemain string

	if nextParamsIdx == -1 {
		token = code
		nextRemain = ""
	} else {
		token = code[:nextParamsIdx]
		nextRemain = code[nextParamsIdx:]
	}

	if token == "#t" {
		return SBool(true), nextRemain, nil
	} else if token == "#f" {
		return SBool(false), nextRemain, nil
	} else if num, err := strconv.ParseFloat(token, 64); err == nil {
		return SNumber(num), nextRemain, nil
	} else {
		return SSym(token), nextRemain, nil
	}
}

var SpecialForms map[string]func(args []Expr, env *Env) (Val, error)

func init() {
	SpecialForms = map[string]func(args []Expr, env *Env) (Val, error){}
	SpecialForms["if"] = func(args []Expr, env *Env) (Val, error) {
		if len(args) != 3 {
			return nil, errors.New("if() requires 3 arguments")
		}
		condition, err := args[0].evaluate(env)
		if err != nil {
			return nil, err
		}
		conditionBool, ok := condition.(SBool)
		if !ok {
			return nil, errors.New("if() requires bool condition")
		}
		if conditionBool {
			return args[1].evaluate(env)
		} else {
			return args[2].evaluate(env)
		}
	}
	SpecialForms["and"] = func(args []Expr, env *Env) (Val, error) {
		if len(args) != 2 {
			return nil, errors.New("and() requires 2 arguments")
		}
		condition1, err := args[0].evaluate(env)
		if err != nil {
			return nil, err
		}
		condition1Bool, ok := condition1.(SBool)
		if !ok {
			return nil, errors.New("and() requires bool condition")
		}
		if !condition1Bool {
			return condition1, nil
		}
		condition2, err := args[1].evaluate(env)
		if err != nil {
			return nil, err
		}
		_, ok = condition2.(SBool)
		if !ok {
			return nil, errors.New("and() requires bool condition")
		}
		return condition2, nil
	}
	SpecialForms["or"] = func(args []Expr, env *Env) (Val, error) {
		if len(args) != 2 {
			return nil, errors.New("or() requires 2 arguments")
		}
		condition1, err := args[0].evaluate(env)
		if err != nil {
			return nil, err
		}
		condition1Bool, ok := condition1.(SBool)
		if !ok {
			return nil, errors.New("or() requires bool condition")
		}
		if condition1Bool {
			return condition1, nil
		}
		condition2, err := args[1].evaluate(env)
		if err != nil {
			return nil, err
		}
		_, ok = condition2.(SBool)
		if !ok {
			return nil, errors.New("or() requires bool condition")
		}
		return condition2, nil
	}
	SpecialForms["lambda"] = func(args []Expr, env *Env) (Val, error) {
		if len(args) != 2 {
			return nil, errors.New("lambda() requires 2 arguments")
		}
		parametersSlist, ok := args[0].(SList)
		if !ok {
			return nil, errors.New("lambda() parameter requires list argument")
		}
		parameters := []SSym{}
		for _, para := range parametersSlist {
			sym, ok := para.(SSym)
			if !ok {
				return nil, errors.New("lambda() parameter requires symbol argument")
			}
			parameters = append(parameters, sym)
		}
		body := args[1]
		return SLambda(func(args []Val, env *Env) (Val, error) {
			if len(args) != len(parameters) {
				//fmt.Println("len of args:", len(args))
				//fmt.Print("len of parameters:", len(parameters))
				//for i, para := range parameters {
				//fmt.Println("args[", i, "]", para)
				//}
				return nil, errors.New(" argument length mismatch")
			}
			lambdaEnv := NewEnv(env)
			for i, arg := range args {
				lambdaEnv.Define(string(parameters[i]), arg)
			}
			return body.evaluate(lambdaEnv)
		}), nil
	}
}

func main() {
	global_env := NewEnv(nil)
	global_env.Define("+", SLambda(func(args []Val, env *Env) (Val, error) {
		res := float64(0)
		for _, arg := range args {
			num, ok := arg.(SNumber)
			if !ok {
				return nil, errors.New("lambda() parameter requires numeric argument")
			}
			res += float64(num)
		}
		return SNumber(res), nil
	}))
	global_env.Define("*", SLambda(func(args []Val, env *Env) (Val, error) {
		res := float64(1)
		for _, arg := range args {
			num, ok := arg.(SNumber)
			if !ok {
				return nil, errors.New("lambda() parameter requires numeric argument")
			}
			res *= float64(num)
		}
		return SNumber(res), nil
	}))
	global_env.Define("-", SLambda(func(args []Val, env *Env) (Val, error) {
		if len(args) != 2 {
			return nil, errors.New("-() requires 2 arguments")
		}
		num1, ok := args[0].(SNumber)
		if !ok {
			return nil, errors.New("lambda() parameter requires numeric argument")
		}
		num2, ok := args[1].(SNumber)
		if !ok {
			return nil, errors.New("lambda() parameter requires numeric argument")
		}
		return SNumber(num1 * num2), nil
	}))
	global_env.Define("/", SLambda(func(args []Val, env *Env) (Val, error) {
		if len(args) != 2 {
			return nil, errors.New("/() requires 2 arguments")
		}
		num1, ok := args[0].(SNumber)
		if !ok {
			return nil, errors.New("lambda() parameter requires numeric argument")
		}
		num2, ok := args[1].(SNumber)
		if !ok {
			return nil, errors.New("lambda() parameter requires numeric argument")
		}
		return SNumber(num1 * num2), nil
	}))
	global_env.Define("=", SLambda(func(args []Val, env *Env) (Val, error) {
		if len(args) != 2 {
			return nil, errors.New("=() requires 2 arguments")
		}
		if args[0] == args[1] {
			return SBool(true), nil
		} else {
			return SBool(false), nil
		}
	}))
	global_env.Define(">", SLambda(func(args []Val, env *Env) (Val, error) {
		if len(args) != 2 {
			return nil, errors.New(">() requires 2 arguments")
		}
		num1, ok := args[0].(SNumber)
		if !ok {
			return nil, errors.New("lambda() parameter requires numeric argument")
		}
		num2, ok := args[1].(SNumber)
		if !ok {
			return nil, errors.New("lambda() parameter requires numeric argument")
		}
		return SBool(num1 > num2), nil
	}))
	global_env.Define("<", SLambda(func(args []Val, env *Env) (Val, error) {
		if len(args) != 2 {
			return nil, errors.New("<() requires 2 arguments")
		}
		num1, ok := args[0].(SNumber)
		if !ok {
			return nil, errors.New("lambda() parameter requires numeric argument")
		}
		num2, ok := args[1].(SNumber)
		if !ok {
			return nil, errors.New("lambda() parameter requires numeric argument")
		}
		return SBool(num1 < num2), nil
	}))

	global_env.Define("<=", SLambda(func(args []Val, env *Env) (Val, error) {
		if len(args) != 2 {
			return nil, errors.New("<=() requires 2 arguments")
		}
		num1, ok := args[0].(SNumber)
		if !ok {
			return nil, errors.New("lambda() parameter requires numeric argument")
		}
		num2, ok := args[1].(SNumber)
		if !ok {
			return nil, errors.New("lambda() parameter requires numeric argument")
		}
		return SBool(num1 <= num2), nil
	}))
	global_env.Define(">=", SLambda(func(args []Val, env *Env) (Val, error) {
		if len(args) != 2 {
			return nil, errors.New(">=() requires 2 arguments")
		}
		num1, ok := args[0].(SNumber)
		if !ok {
			return nil, errors.New("lambda() parameter requires numeric argument")
		}
		num2, ok := args[1].(SNumber)
		if !ok {
			return nil, errors.New("lambda() parameter requires numeric argument")
		}
		return SBool(num1 >= num2), nil
	}))
	global_env.Define("not", SLambda(func(args []Val, env *Env) (Val, error) {
		if len(args) != 1 {
			return nil, errors.New("not() requires 1 argument")
		}
		bool_, ok := args[0].(SBool)
		if !ok {
			return nil, errors.New("not() requires true argument")
		}
		return bool_, nil
	}))

	if len(os.Args) > 1 {
		// file mode
		filename := os.Args[1]
		fileMode(filename, global_env)
	} else {
		// repl mode
		fmt.Println("Welcome to the Scheme REPL (Go version)")
		replMode(global_env)
	}
}

func fileMode(testFile string, env *Env) {
	// 1. 读取整个文件
	bytes, err := os.ReadFile(testFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	code := string(bytes)
	remain := strings.TrimSpace(code)

	// 2. 正则去除注释 (; 到行尾)，对应 TS 的 .replace(/;.*$/gm, "")
	// (?m) 开启多行模式
	re := regexp.MustCompile(`(?m);.*$`)
	remain = re.ReplaceAllString(remain, "")
	remain = strings.TrimSpace(remain)

	// fmt.Printf("Running file: %s\n", testFile)

	// 3. 循环消费字符串
	for len(strings.TrimSpace(remain)) > 0 {
		var expr Expr
		var err error

		expr, remain, err = readExpr(remain)

		if err != nil {
			fmt.Printf("Syntax Error: %v\n", err)
			break
		}

		res, err := expr.evaluate(env)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Println(res.repr())
		}
	}
}

// 核心逻辑：REPL 模式
func replMode(env *Env) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("scheme> ")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			fmt.Print("scheme> ")
			continue
		}

		// 解析当前行
		expr, remain, err := readExpr(line)
		if err != nil {
			fmt.Printf("Syntax Error: %v\n", err)
			fmt.Print("scheme> ")
			continue
		}

		if len(strings.TrimSpace(remain)) != 0 {
			fmt.Println("Only one expression per line!")
		}

		res, err := expr.evaluate(env)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Println(res.repr())
		}

		fmt.Print("scheme> ")
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	fmt.Println("Exiting...")
}
