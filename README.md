
## Simple Scheme Interpreter 规范

本解释器旨在处理一个精简的 Scheme 子集，它支持两种交互模式：

1. REPL（Read-Eval-Print Loop）模式：接收用户的交互式输入，对每一行表达式进行求值并立即打印结果。

2. 文件模式：从指定文件中读取表达式，对每一行进行求值并将结果逐行打印。


实现进度：

- [x] Python
- [X] C
- [X] TypeScript
- [X] Golang

### IO格式

1. repl
2. 文件 -> 逐行打印每个 expression 的值

### 语言定义

#### 基本构件

我们的解释器将处理三种基本数据类型：

  - **数字（float）**：用于算术运算，例如 `10.0`、`3.14`。
  - **布尔值**：`#t`（真）和 `#f`（假），用于控制流程。
  - **符号**：变量名或函数名，例如 `x`, `sum`。

#### 函数与特殊形式

在 Scheme 中，所有操作（包括算术和控制流）都是通过函数调用或特殊形式来完成的。

  - 算术:
    `(+ 1 2 3)` 将计算结果为 `6`。
    `(- 10 4)` 将计算结果为 `6`。

  - 条件语句 (`if`):
    `if` 表达式用于条件分支，它会先求值第一个参数，如果结果为 `#f` 以外的值，则求值第二个参数，否则求值第三个参数。

    ```scheme
    (if #t (+ 1 1) (- 1 1))  ; => 2
    (if #f (+ 1 1) (- 1 1))  ; => 0
    ```

  - 变量定义 (`define`):
    `define` 用于创建全局变量绑定。

    ```scheme
    (define x 10)  ; 定义变量x，并赋值10
    (+ x 5)       ; => 15
    ```

  - 匿名函数 (`lambda`):
    `lambda` 表达式创建一个匿名函数，它会捕获定义时的环境，因此可以访问外部变量。

    ```scheme
    (define adder (lambda (x y) (+ x y)))
    (adder 5 6)    ; => 11

    ; 嵌套示例
    (define make-adder
      (lambda (x)
        (lambda (y) (+ x y))))
    (define add-five (make-adder 5))
    (add-five 10)   ; => 15
    ```

    在上面的例子中，`add-five` 是一个闭包，它记住了 `make-adder` 被调用时 `x` 的值是 `5`。

这些简单的规则和构件，构成了我们解释器的全部核心功能，足以实现一个图灵完备的子集。

### Syntax

我们使用 Backus-Naur Form (BNF) 来形式化描述这个 Scheme 子集的语法。

#### 原子表达式 (Atoms)

  - 数字 ($<number>$): 一个浮点数，如 `123`, `3.14`。
  - 布尔值 ($<boolean>$): `#t` 或 `#f`。
  - 符号 ($<symbol>$): 一个以字母开头的字符串，可以包含字母、数字和 `-`。例如 `x`, `+`, `my-var`。

#### 列表表达式 (Lists)

  - 列表 ($<list>$): 由圆括号括起来的一系列表达式。
    $$
    <list> ::= \text{( } <expression>^* \text{ )}
    $$

#### 完整表达式 (Expressions)

一个 Scheme 程序由以下表达式组成：

$$
<expression> ::= <atom> \mid <list>
$$

$$
<atom> ::= <number> \mid <boolean> \mid <symbol>
$$

#### 特殊形式与函数调用 (Special Forms & Procedure Application)

这些列表表达式具有特殊的求值规则：

  - **过程应用 (Procedure Application)**:
    $$
    <application> ::= \text{( } <procedure> \ <argument>^* \text{ )}
    $$
    其中 $<procedure>$ 是一个求值后得到函数的表达式，$<argument>$ 是一个表达式。

  - **`define`**: 用于定义变量。
    $$
    <define> ::= \text{( define } <symbol> \ <expression> \text{ )}
    $$

  - **`if`**: 条件分支。
    $$
    <if> ::= \text{( if } <test> \ <consequent> \ <alternative> \text{ )}
    $$

  - **`lambda`**: 创建匿名函数（闭包）。
    $$
    <lambda> ::= \text{( lambda } ( <symbol>^* ) \ <expression> \text{ )}
    $$
  - 基本运算符：
    - 算术（`+`, `-`, `*`, `/`, `=`, `<`, `>`, `<=`, `>=`）
    - 逻辑（`and`, `or`, `not`）

### Semantics

我们用一个求值函数 $E(expr, env)$ 来描述表达式的求值规则，其中 $expr$ 是待求值的表达式，$env$ 是当前的环境。

#### **环境 (Environment)**

一个**环境**（$\rho$）是一个从**符号**到**值**的映射。它支持：

  - **查询**: $\rho(s)$ 查找符号 $s$ 对应的值。
  - **更新**: $\rho[s \leftarrow v]$ 在环境中添加或更新映射。

#### **求值规则 (Evaluation Rules)**

1.  **原子表达式**:

      - **数字**: $E(n, \rho) = n$
      - **布尔值**: $E(b, \rho) = b$
      - **符号**: $E(s, \rho) = \rho(s)$，即在环境中查找绑定的值。

2.  **`if` 表达式**:

      - 如果 $E(test, \rho)$ 的结果为真（非 `#f`），则结果是 $E(consequent, \rho)$。
      - 否则，结果是 $E(alternative, \rho)$。

3.  **`define` 表达式**:

      - 求值 $expr$ 得到值 $v = E(expr, \rho)$。
      - 更新当前环境：$\rho_{\text{new}} = \rho[s \leftarrow v]$。
      - 求值结果通常为不可见或空值。

4.  **`lambda` 表达式**:

      - 求值结果是一个**闭包**。这个闭包是一个特殊的数据结构，包含三个部分：参数列表、函数体以及**定义时的环境**（即 $\\rho$）。

5.  **过程应用 (函数调用)**:

      - 求值过程：首先求值过程表达式 $P = E(proc, \rho)$。
      - 求值参数：然后逐一求值所有参数 $V_i = E(arg_i, \rho)$。
      - 如果 $P$ 是一个闭包，它会创建一个**新的环境**，将参数与实参绑定，并以外层环境为闭包定义时的环境。
      - 最后，在新环境中求值闭包的函数体。


### 扩展

1. pair
2. string
3. macro
4. begin
5. future/promise

## 编程语言评价维度

1. 语言设计
   1. 类型系统（影响安全，表达能力等）
   2. 抽象能力（宏，代码生成）
   3. 编程范式
2. 运行时
   1. 内存模型
   2. 并发模型
   3. 性能
3. 开发体验
   1. 工具链
   2. 生态系统
4. 语言哲学
   3. 目标用户？核心思想？适用场景？

