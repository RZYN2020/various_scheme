package main

import (
	"errors"
	"fmt"
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
type SLambda func([]Val) Val

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

	fun_wrapped, err := n[0].evaluate(env)
	if err != nil {
		return nil, err
	}

	fun, ok := fun_wrapped.(SLambda)
	if !ok {
		return nil, errors.New("first element is not a function")
	}

	args := make([]Val, len(n)-1)
	for i, e := range n[1:] {
		val, err := e.evaluate(env)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}

	return fun(args), nil
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

// todo:
// 1. special forms
// 2. global functions
// 3. IO and system test
// 4. judge and reflection

func main() {
	global_env := NewEnv(nil)

	expr_s := "afafa"
	global_env.Define("afafa", SNumber(121))
	expr, _, err := readExpr(expr_s)
	val, err := expr.evaluate(global_env)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(val)
	}
}
