# Go 开发反思

## 零、总体体验


1. 表达能力：没有宏，而且静态类型，因此代码可能冗余，违反DRY原则。可以采取高阶函数/泛型/go generate消除部分冗余
2. 错误处理：没有exception，需要手动处理
3. 类型系统：
   4. 没有sum type，需手动判断nil，不安全
   5. 静态 Duck Typing
   6. 以组合替代子类型
7. 其他特性：并发友好（但scheme.go中没有涉及），工具链完善

## 一、 命名与可见性 (Naming & Visibility)

### 1\. 大小写控制可见性 (camelCase vs PascalCase)

**核心规则**：Go 语言没有 `public` / `private` 关键字，首字母大小写决定了跨包访问权限。

* **PascalCase (大写开头)**: **导出 (Exported)**。其他包可以访问。通常用于对外提供的 API、Struct 的字段。
* **camelCase (小写开头)**: **未导出 (Unexported)**。仅当前包内可见。用于内部实现细节、辅助函数。

**示例：**

```go
package interpreter

// User 是导出的，外部可用
type User struct {
    Name string // 导出的字段，JSON 序列化需要
    age  int    // 未导出的字段，仅包内逻辑可用，JSON 序列化会忽略
}

// NewUser 是“构造函数”，需要导出
func NewUser(name string, age int) *User {
    return &User{Name: name, age: age}
}

// validate 是内部辅助逻辑，无需暴露给外部
func validate(u *User) bool {
    return u.age >= 0
}
```

### 2\. 去除冗余前缀 (Package Context)

**反思**：在该项目中，你最初使用了 `SNumber`, `SBool`, `SLambda`（S 代表 Scheme）。
**改进**：在 Go 中，包名本身就是上下文。应当移除类型名称中重复的上下文信息，保持代码清爽。

**示例：**

```go
package scheme

// Bad: 命名冗余，调用时变成 scheme.SNumber
type SNumber float64
type SBool bool

// Good: 简洁，调用时变成 scheme.Number, scheme.Bool
type Number float64
type Bool bool

// 如果是在 main 包内部，为了区分，可以直接叫 Number 而不加 S
// 只要不跟标准库冲突，名字越短越好
```

-----

## 二、 错误处理 (Error Handling)

### 3\. fmt.Errorf vs errors.New

**反思**：`errors.New` 只能返回静态文本。`fmt.Errorf` 可以动态拼接上下文信息，这在 Debug 时至关重要。

**最佳实践**：

* **定义哨兵错误**：用于逻辑判断（如 `io.EOF`）。
* **使用 fmt.Errorf**：用于返回带数据的错误描述。

**示例：**

```go
import ("errors"; "fmt")

var ErrEmptyList = errors.New("empty list") // 哨兵错误，可用于 == 判断

func parse(input string) error {
    if len(input) == 0 {
        return ErrEmptyList
    }
    if input == "error" {
        // Bad: return errors.New("parse error") // 丢失了具体是什么输入错了
        
        // Good: 携带上下文
        return fmt.Errorf("parsing failed for input: '%s'", input)
    }
    return nil
}
```

-----

## 三、 类型系统与断言 (Type System & Assertions)

### 4\. 安全的类型断言 (Comma-ok Idiom)

**核心规则**：永远不要直接使用 `val.(Type)`，除非你 100% 确定类型，否则一旦类型不匹配会直接 **Panic** 导致程序崩溃。

**示例：**

```go
// Bad: 假设 args[0] 一定是 Symbol，如果传了 Number 程序直接挂掉
// sym := args[0].(Symbol) 

// Good: 优雅降级
if sym, ok := args[0].(Symbol); ok {
    // 只有在类型匹配时才执行
    env.Define(string(sym), val)
} else {
    // 类型不匹配时的处理
    return nil, fmt.Errorf("expected Symbol, got %T", args[0])
}
```

### 5\. 类型分支 (Type Switch)

**反思**：当你需要对一个接口做多种类型的判断时，使用 `switch type` 比写一堆 `if-else` 更高效且易读。

**示例：**

```go
func (l List) Eval(env *Env) (Val, error) {
    fnVal, _ := l[0].Eval(env)

    // 针对 fnVal 的不同底层类型做不同处理
    switch fn := fnVal.(type) {
    case Builtin:
        return fn(args) // fn 自动被转为 Builtin 类型
    case Lambda:
        return evalLambda(fn, args) // fn 自动被转为 Lambda 类型
    default:
        return nil, fmt.Errorf("not a function: %T", fnVal)
    }
}
```

-----

## 四、 代码结构与复用 (Structure & DRY)

### 6\. 高阶函数抽离逻辑 (Higher-Order Functions)

**反思**：如果你发现自己在复制粘贴代码（比如 `+`, `-`, `*`, `/` 的逻辑非常像），就应该写一个“生成函数的函数”。

**示例：**

```go
// 原始痛点：加减乘除写了 4 遍类似的逻辑
// func Add(...) { checkNumber; return a + b }
// func Sub(...) { checkNumber; return a - b }

// 改进：抽离公共逻辑（参数检查、循环），只传入核心差异（操作符）
func binaryNumOp(op func(float64, float64) float64) Builtin {
    return func(args []Val) (Val, error) {
        // 统一的参数检查逻辑
        nums, err := assertNums(args) 
        if err != nil { return nil, err }
        
        // 统一的计算逻辑
        res := nums[0]
        for _, n := range nums[1:] {
            res = op(res, n) // 调用传入的核心逻辑
        }
        return Number(res), nil
    }
}

// 使用
env.Set("+", binaryNumOp(func(a, b float64) float64 { return a + b }))
env.Set("*", binaryNumOp(func(a, b float64) float64 { return a * b }))
```

### 7\. Map 直接初始化 (Map Literals)

**反思**：与其先 `make` 再一行行 `m[k]=v`，不如使用字面量直接初始化，结构一目了然。这也便于维护查找表（Table-Driven Design）。

**示例：**

```go
// Bad: 这种写法在 main 函数里显得很啰嗦
// forms := make(map[string]func(...))
// forms["if"] = func(...) { ... }
// forms["define"] = func(...) { ... }

// Good: 声明式写法，结构清晰，易于阅读
var specialForms = map[string]SpecialForm{
    "if": func(args []Expr, env *Env) (Val, error) {
        // if logic
        return nil, nil
    },
    "define": func(args []Expr, env *Env) (Val, error) {
        // define logic
        return nil, nil
    },
    "lambda": func(args []Expr, env *Env) (Val, error) {
        // lambda logic
        return nil, nil
    },
}
```