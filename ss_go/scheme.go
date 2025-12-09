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

// ==========================================
// 1. 类型定义与环境 (Types & Environment)
// ==========================================

type Env struct {
	parent *Env
	store  map[string]Val
}

func NewEnv(parent *Env) *Env {
	return &Env{parent: parent, store: make(map[string]Val)}
}

func (e *Env) Get(name string) (Val, bool) {
	if v, ok := e.store[name]; ok {
		return v, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

func (e *Env) Set(name string, val Val) {
	e.store[name] = val
}

type Val interface {
	String() string // Go 惯用 String() 而非 repr()
}

type (
	Number  float64
	Bool    bool
	Symbol  string
	List    []Expr
	Builtin func([]Val) (Val, error) // 简化原生函数签名
	Lambda  struct {
		params []string
		body   Expr
		env    *Env
	}
)

func (n Number) String() string { return strconv.FormatFloat(float64(n), 'f', -1, 64) }
func (b Bool) String() string {
	if b {
		return "#t"
	}
	return "#f"
}
func (s Symbol) String() string { return string(s) }
func (l List) String() string {
	parts := make([]string, len(l))
	for i, e := range l {
		parts[i] = fmt.Sprintf("%v", e) // 简化打印
	}
	return "(" + strings.Join(parts, " ") + ")"
}
func (b Builtin) String() string { return "<builtin>" }
func (l Lambda) String() string  { return "<lambda>" }

// ==========================================
// 2. 表达式与求值 (Expression & Evaluation)
// ==========================================

type Expr interface {
	Eval(env *Env) (Val, error)
}

// 原子类型自求值
func (n Number) Eval(env *Env) (Val, error) { return n, nil }
func (b Bool) Eval(env *Env) (Val, error)   { return b, nil }
func (s Symbol) Eval(env *Env) (Val, error) {
	if v, ok := env.Get(string(s)); ok {
		return v, nil
	}
	return nil, fmt.Errorf("undefined symbol: %s", s)
}

// 列表求值（核心逻辑）
func (l List) Eval(env *Env) (Val, error) {
	if len(l) == 0 {
		return nil, fmt.Errorf("cannot evaluate empty list")
	}

	// 1. 检查是否为 Special Form (if, define, lambda, etc.)
	// 这里为了简化，假设 Special Form 的第一个元素必须是 Symbol
	if headSym, ok := l[0].(Symbol); ok {
		if handler, ok := specialForms[string(headSym)]; ok {
			return handler(l[1:], env)
		}
	}

	// 2. 函数调用：先求值第一个元素
	fnVal, err := l[0].Eval(env)
	if err != nil {
		return nil, err
	}

	// 3. 求值所有参数
	args := make([]Val, 0, len(l)-1)
	for _, expr := range l[1:] {
		arg, err := expr.Eval(env)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	// 4. 执行函数
	switch fn := fnVal.(type) {
	case Builtin:
		return fn(args)
	case Lambda:
		if len(args) != len(fn.params) {
			return nil, fmt.Errorf("arity mismatch: expected %d args, got %d", len(fn.params), len(args))
		}
		// 创建闭包环境
		newEnv := NewEnv(fn.env)
		for i, param := range fn.params {
			newEnv.Set(param, args[i])
		}
		return fn.body.Eval(newEnv)
	default:
		return nil, fmt.Errorf("not a function: %s", l[0])
	}
}

// ==========================================
// 3. Special Forms & Builtins (统一管理)
// ==========================================

// SpecialForm 处理器接收未求值的参数 AST
type SpecialForm func(args []Expr, env *Env) (Val, error)

var specialForms = map[string]SpecialForm{
	"define": func(args []Expr, env *Env) (Val, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("define takes 2 args")
		}
		sym, ok := args[0].(Symbol)
		if !ok {
			return nil, fmt.Errorf("define first arg must be symbol")
		}
		val, err := args[1].Eval(env)
		if err != nil {
			return nil, err
		}
		env.Set(string(sym), val)
		return nil, nil // define 返回 nil
	},
	"if": func(args []Expr, env *Env) (Val, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("if takes 3 args")
		}
		cond, err := args[0].Eval(env)
		if err != nil {
			return nil, err
		}
		// 只有 #f 是 false，其他都是 true
		if b, ok := cond.(Bool); ok && !bool(b) {
			return args[2].Eval(env)
		}
		return args[1].Eval(env)
	},
	"lambda": func(args []Expr, env *Env) (Val, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("lambda takes 2 args")
		}
		paramsNode, ok := args[0].(List)
		if !ok {
			return nil, fmt.Errorf("lambda params must be a list")
		}
		params := make([]string, len(paramsNode))
		for i, p := range paramsNode {
			sym, ok := p.(Symbol)
			if !ok {
				return nil, fmt.Errorf("param must be symbol")
			}
			params[i] = string(sym)
		}
		return Lambda{params: params, body: args[1], env: env}, nil
	},
	"and": func(args []Expr, env *Env) (Val, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("and takes 2 args")
		}
		cond1, err := args[0].Eval(env)
		if err != nil {
			return nil, err
		}
		if cond1b, ok := cond1.(Bool); ok {
			if !cond1b {
				return cond1, nil
			}
			cond2, err := args[1].Eval(env)
			if err != nil {
				return nil, err
			}
			if cond2b, ok := cond2.(Bool); ok {
				return cond2b, nil
			}
			return nil, fmt.Errorf("and arg must be a bool")
		}
		return nil, fmt.Errorf("and arg must be a bool")
	},
	"or": func(args []Expr, env *Env) (Val, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("or takes 2 args")
		}
		cond1, err := args[0].Eval(env)
		if err != nil {
			return nil, err
		}
		if cond1b, ok := cond1.(Bool); ok {
			if cond1b {
				return cond1, nil
			}
			cond2, err := args[1].Eval(env)
			if err != nil {
				return nil, err
			}
			if cond2b, ok := cond2.(Bool); ok {
				return cond2b, nil
			}
			return nil, fmt.Errorf("or arg must be a bool")
		}
		return nil, fmt.Errorf("or arg must be a bool")
	},
}

// 辅助函数：将 Number 转换逻辑抽离
func assertNums(args []Val) ([]float64, error) {
	nums := make([]float64, len(args))
	for i, arg := range args {
		n, ok := arg.(Number)
		if !ok {
			return nil, fmt.Errorf("expected number, got %T", arg)
		}
		nums[i] = float64(n)
	}
	return nums, nil
}

// 辅助函数：创建二元数值操作符 (减少重复代码)
func binaryNumOp(op func(a, b float64) float64) Builtin {
	return func(args []Val) (Val, error) {
		nums, err := assertNums(args)
		if err != nil {
			return nil, err
		}
		if len(nums) < 2 {
			// 这里简单处理，严格来说 - 可以是一元操作符
			return nil, fmt.Errorf("arithmetic op requires at least 2 args")
		}
		res := nums[0]
		for _, n := range nums[1:] {
			res = op(res, n)
		}
		return Number(res), nil
	}
}

// 辅助函数：创建比较操作符
func compareOp(op func(a, b float64) bool) Builtin {
	return func(args []Val) (Val, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("comparison requires 2 args")
		}
		nums, err := assertNums(args)
		if err != nil {
			return nil, err
		}
		return Bool(op(nums[0], nums[1])), nil
	}
}

func loadBuiltins(env *Env) {
	env.Set("+", binaryNumOp(func(a, b float64) float64 { return a + b }))
	env.Set("-", binaryNumOp(func(a, b float64) float64 { return a - b }))
	env.Set("*", binaryNumOp(func(a, b float64) float64 { return a * b }))
	env.Set("/", binaryNumOp(func(a, b float64) float64 { return a / b }))
	env.Set("=", compareOp(func(a, b float64) bool { return a == b }))
	env.Set("<", compareOp(func(a, b float64) bool { return a < b }))
	env.Set("<=", compareOp(func(a, b float64) bool { return a <= b }))
	env.Set(">=", compareOp(func(a, b float64) bool { return a >= b }))
	env.Set(">", compareOp(func(a, b float64) bool { return a > b }))
	env.Set("not", Builtin(func(args []Val) (Val, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("not takes 1 args")
		}
		if val, ok := args[0].(Bool); ok {
			return !val, nil
		}
		return nil, fmt.Errorf("not takes bool arg")
	}))
}

// ==========================================
// 4. 解析器 (Parser) - 使用 Tokenizer 改进
// ==========================================

func readExpr(code string) (Expr, string, error) {
	code = strings.TrimSpace(code)

	if len(code) == 0 {
		return nil, "", errors.New("unexpected EOF")
	}

	first := code[0]
	if first == '(' {
		remain := code[1:]
		list := List{}
		for {
			remain = strings.TrimSpace(remain)
			if len(remain) == 0 {
				return nil, "", errors.New("unbalanced parenthesis: missing ')'")
			}
			if remain[0] == ')' {
				return list, remain[1:], nil
			}
			expr, remainNext, err := readExpr(remain)
			if err != nil {
				return nil, "", err
			}
			list = append(list, expr)
			remain = remainNext
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
		return Bool(true), nextRemain, nil
	} else if token == "#f" {
		return Bool(false), nextRemain, nil
	} else if num, err := strconv.ParseFloat(token, 64); err == nil {
		return Number(num), nextRemain, nil
	} else {
		return Symbol(token), nextRemain, nil
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

		res, err := expr.Eval(env)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else if res != nil {
			fmt.Println(res.String())
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

		res, err := expr.Eval(env)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else if res != nil {
			fmt.Println(res.String())
		}

		fmt.Print("scheme> ")
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	fmt.Println("Exiting...")
}

// ==========================================
// 5. Main Loop
// ==========================================

// todo：scheme 添加 future / promise 并发支持
func main() {
	env := NewEnv(nil)
	loadBuiltins(env)

	if len(os.Args) > 1 {
		// 直接调用封装好的 runFile
		fileMode(os.Args[1], env)
	} else {
		// REPL 模式 (逻辑也可以复用 tokenize)
		replMode(env)
	}
}
