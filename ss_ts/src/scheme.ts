import * as fs from "fs";
import * as readline from "readline";

// Evaluator
type Lambda = (args: Val[], env: Env) => Val;
type Val = number | boolean | Lambda;

function vprint(val: Val) {
  if (typeof val === "number") {
    console.log(val);
  } else if (typeof val === "function") {
    console.log("<lambda function>");
  } else if (val === true) {
    console.log("#t");
  } else {
    console.log("#f");
  }
}

class Env {
  parent: Env | null;
  bindings: Map<string, Val>;

  constructor(parent: Env | null, bindings: Map<string, Val>) {
    this.parent = parent;
    this.bindings = bindings;
  }

  find(name: string): Env | null {
    if (this.bindings.has(name)) {
      return this;
    } else if (this.parent !== null) {
      return this.parent.find(name);
    } else {
      return null;
    }
  }

  get(name: string): Val {
    const env = this.find(name);
    if (env !== null) {
      return env.bindings.get(name) as Val;
    } else {
      throw Error(`${name} is not defined`);
    }
  }

  set(name: string, val: Val): void {
    this.bindings.set(name, val);
  }
}

function _define(expr: Expr, env: Env) {
  if (!Array.isArray(expr) || expr.length !== 3) {
    throw Error("Syntax error: 'define' expects exactly 2 arguments (name and value).");
  }
  const name = expr[1];
  if (typeof name !== "string") {
    throw Error(`Syntax error: 'define' expects a symbol as the first argument, but got ${JSON.stringify(name)}`);
  }
  const value = evaluate(expr[2] as Expr, env);
  env.set(name, value);
  return;
}

function _if(expr: Expr, env: Env): Val {
  if (!Array.isArray(expr) || (expr.length !== 3 && expr.length !== 4)) {
    throw Error("Syntax error: 'if' expects 2 or 3 arguments.");
  }
  const test = evaluate(expr[1] as Expr, env);
  if (typeof test !== "boolean") {
    throw Error("Type error: 'if' test expression must evaluate to a boolean.");
  }
  if (test) {
    return evaluate(expr[2] as Expr, env);
  } else {
    if (expr.length === 4) {
      return evaluate(expr[3] as Expr, env);
    } else {
      return false;
    }
  }
}

function _and(expr: Expr, env: Env): Val {
  if (!Array.isArray(expr) || expr.length < 2) {
    throw Error("Syntax error: 'and' expects at least one argument.");
  }
  let result: Val = true;
  for (let i = 1; i < expr.length; i++) {
    result = evaluate(expr[i] as Expr, env);
    if (result === false) {
      return false;
    }
  }
  return result;
}

function _or(expr: Expr, env: Env): Val {
  if (!Array.isArray(expr) || expr.length < 2) {
    throw Error("Syntax error: 'or' expects at least one argument.");
  }
  let result: Val = false;
  for (let i = 1; i < expr.length; i++) {
    result = evaluate(expr[i] as Expr, env);
    if (result !== false) {
      return result;
    }
  }
  return false;
}

function _lambda(expr: Expr, env: Env): Val {
  if (!Array.isArray(expr) || expr.length < 3) {
    throw Error("Syntax error: 'lambda' expects at least 2 arguments (parameters and body).");
  }
  const params = expr[1];
  if (!Array.isArray(params) || !params.every((p) => typeof p === "string")) {
    throw Error("Syntax error: 'lambda' parameters must be a list of symbols.");
  }
  const body = expr.slice(2);
  return function (args: Val[], subenv: Env): Val {
    if (args.length !== params.length) {
      throw Error(`Argument count mismatch: Lambda function expected ${params.length} arguments but got ${args.length}.`);
    }
    const lambdaEnv = new Env(env, new Map<string, Val>());
    for (let i = 0; i < params.length; i++) {
      lambdaEnv.set(params[i] as string, args[i] as Val);
    }
    let result: Val = false;
    for (const expr of body) {
      result = evaluate(expr as Expr, lambdaEnv);
    }
    return result;
  };
}

const special_forms = new Map<string, (args: Expr[], env: Env) => Val>([
  ["if", _if],
  ["lambda", _lambda],
  ["and", _and],
  ["or", _or],
]);

function evaluate(expr: Expr, env: Env): Val {
  if (typeof expr === "number" || typeof expr === "boolean") {
    return expr;
  }
  if (typeof expr === "string") {
    return env.get(expr);
  }
  if (!Array.isArray(expr) || expr.length === 0) {
    throw Error("Evaluation error: Cannot evaluate an empty list.");
  }

  const form = expr[0] as Expr;
  if (typeof form === "string" && special_forms.has(form)) {
    return special_forms.get(form)!(expr, env);
  } else {
    const func = evaluate(form, env);
    if (typeof func !== "function") {
      throw Error(`Type error: Expression starts with non-callable value: ${JSON.stringify(form)}`);
    }

    const args = expr.slice(1).map((arg) => evaluate(arg as Expr, env));
    return func(args, env);
  }
}

// Parser
type Sym = string;
type Expr = number | boolean | Sym | Expr[];

function readNumber(code: string): [number, string] {
  const numberPattern = /^[-+]?\d+(\.\d+)?/;
  const match = code.match(numberPattern);

  if (match) {
    const numberStr = match[0];
    const parsedNumber = parseFloat(numberStr);
    const remain = code.substring(numberStr.length);
    return [parsedNumber, remain];
  }

  throw Error("Parser error: Invalid number format.");
}

function readSym(code: string): [string, string] {
  for (let i = 1; i <= code.length; i++) {
    if (code[i]?.trim().length === 0 || code[i] === ")" ) {
      const sym = code.slice(0, i);
      return [sym, code.slice(i)];
    }
  }
  return [code, ""]
}

function readExpr(code: string): [Expr, string] {
  const code_trimmed = code.trim();
  if (code_trimmed.length === 0) {
    throw Error("Parser error: Unexpected end of input.");
  }
  const first_char = code_trimmed[0] as string;
  if (first_char === "(") {
    let remain = code_trimmed.slice(1);
    let exprs = [] as Expr[];
    while (remain[0] != ")") {
      let expr;
      [expr, remain] = readExpr(remain);
      exprs.push(expr);
    }
    return [exprs, remain.slice(1)];
  } else if (first_char === "#") {
    if (code_trimmed.length < 2 || (code_trimmed[1] !== "t" && code_trimmed[1] !== "f")) {
      throw Error("Parser error: Invalid boolean literal. Use #t or #f.");
    }
    return [code_trimmed[1] == "t" ? true : false, code_trimmed.slice(2)];
  } else if (/^[+-]?\d$/.test(first_char) || (
    (first_char === "-" || first_char === "+") && code_trimmed.length > 1 &&
    /^\d/.test(code_trimmed[1] as string)
  )) {
    return readNumber(code_trimmed);
  } else {
    return readSym(code_trimmed);
  }
}


const globalEnv = new Env(null, new Map<string, Lambda>([
  ["+", (args: Val[], _: Env) => args.reduce((a, b) => (a as number) + (b as number))],
  ["*", (args: Val[], _: Env) => args.reduce((a, b) => (a as number) * (b as number))],
  ["-", (args: Val[], _: Env) => args.reduce((a, b) => (a as number) - (b as number))],
  ["/", (args: Val[], _: Env) => args.reduce((a, b) => (a as number) / (b as number))],
  ["=", (args: Val[], _: Env) => args.reduce((a, b) => a === b)],
  [">", (args: Val[], _: Env) => args.reduce((a, b) => a > b)],
  ["<", (args: Val[], _: Env) => args.reduce((a, b) => a < b)],
  [">=", (args: Val[], _: Env) => args.reduce((a, b) => a >= b)],
  ["<=", (args: Val[], _: Env) => args.reduce((a, b) => a <= b)],
  ["not", (args: Val[], _: Env) => {
    if (args.length !== 1 || typeof args[0] !== "boolean") {
      throw Error(`Type error: 'not' expects a single boolean argument, but got ${args.length} arguments.`);
    }
    return !args[0];
  }],
]));

function file_mode(test_file: string) {
  const code = fs.readFileSync(test_file, "utf-8");
  let remain = code.trim();
  remain = remain.replace(/;.*$/gm, "").trim();

  // console.log(`Running file: ${test_file}`);
  // console.log(`Remain: ${remain}`);

  while (remain.trim().length != 0) {
    let expr;
    [expr, remain] = readExpr(remain);
    // console.log(expr);
    if (Array.isArray(expr) && expr[0] === "define") {
      _define(expr, globalEnv);
    } else {
      vprint(evaluate(expr, globalEnv));
    }
  }
}

function repl_mode() {
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    prompt: "scheme> ",
  });
  rl.on("line", (line) => {
    if (line.trim().length === 0) {
      rl.prompt();
      return;
    }
    const [expr, remain] = readExpr(line);
    if (remain.trim().length != 0) {
      console.log("Only one expression per line!");
    }
    if (Array.isArray(expr) && expr[0] === "define") {
      _define(expr, globalEnv);
    } else {
      vprint(evaluate(expr, globalEnv));
    }
    rl.prompt();
  }).on("close", () => {
    console.log("Exiting...");
    process.exit(0);
  });
  rl.prompt();
}

// process.argv[1] = "../tests/basic_arithmetic.ss"; // For testing in VSCode
if (process.argv.length >= 3) {
  file_mode(process.argv[2] as string);
} else {
  repl_mode();
}